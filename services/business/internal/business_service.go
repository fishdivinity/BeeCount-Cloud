package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/business"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type           string
	SQLiteConfig   SQLiteConfig
	MySQLConfig    MySQLConfig
	PostgresConfig PostgresConfig
}

// SQLiteConfig SQLite配置
type SQLiteConfig struct {
	Path string
}

// MySQLConfig MySQL配置
type MySQLConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// PostgresConfig Postgres配置
type PostgresConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

// Ledger 账本模型
type Ledger struct {
	ID          string    `gorm:"type:varchar(36);primaryKey"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	UserID      string    `gorm:"type:varchar(36);not null;index"`
	Currency    string    `gorm:"type:varchar(10);default:'CNY'"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// Transaction 交易模型
type Transaction struct {
	ID              string            `gorm:"type:varchar(36);primaryKey"`
	LedgerID        string            `gorm:"type:varchar(36);not null;index"`
	UserID          string            `gorm:"type:varchar(36);not null;index"`
	Type            string            `gorm:"type:varchar(20);not null"` // income, expense, transfer
	CategoryID      string            `gorm:"type:varchar(36)"`
	SubcategoryID   string            `gorm:"type:varchar(36)"`
	AccountID       string            `gorm:"type:varchar(36);not null"`
	TargetAccountID string            `gorm:"type:varchar(36)"`
	Amount          string            `gorm:"type:decimal(20,2);not null"`
	Description     string            `gorm:"type:text"`
	Date            string            `gorm:"type:varchar(10);not null;index"`
	CreatedAt       time.Time         `gorm:"autoCreateTime;index"`
	UpdatedAt       time.Time         `gorm:"autoUpdateTime"`
	Tags            map[string]string `gorm:"type:json"`
	SyncTime        int64             `gorm:"not null;index"`
	DeviceID        string            `gorm:"type:varchar(36);not null"`
}

// BusinessService 业务服务实现
type BusinessService struct {
	business.UnimplementedBusinessServiceServer
	common.UnimplementedHealthCheckServiceServer
	db     *gorm.DB
	config DatabaseConfig
}

// NewBusinessService 创建业务服务实例
func NewBusinessService() *BusinessService {
	return &BusinessService{}
}

// ConfigureDatabase 配置数据库
func (s *BusinessService) ConfigureDatabase(config DatabaseConfig) error {
	s.config = config

	// 初始化数据库连接
	var err error
	var db *gorm.DB

	switch config.Type {
	case "sqlite3":
		// 确保数据目录存在
		if err := os.MkdirAll(filepath.Dir(config.SQLiteConfig.Path), 0755); err != nil {
			return err
		}

		// 连接SQLite数据库
		db, err = gorm.Open(sqlite.Open(config.SQLiteConfig.Path), &gorm.Config{})
		if err != nil {
			return err
		}

		// 获取SQLite连接池
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		// 配置连接池
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)

	case "mysql":
		// 连接MySQL数据库
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.MySQLConfig.Username,
			config.MySQLConfig.Password,
			config.MySQLConfig.Host,
			config.MySQLConfig.Port,
			config.MySQLConfig.Database,
		)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}

		// 获取MySQL连接池
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		// 配置连接池
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)

	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	s.db = db
	return nil
}

// InitDatabase 初始化数据库
func (s *BusinessService) InitDatabase() error {
	// 自动迁移模型
	if err := s.db.AutoMigrate(&Ledger{}, &Transaction{}); err != nil {
		return err
	}

	log.Println("Database migrated successfully")
	return nil
}

// Sync 同步数据
func (s *BusinessService) Sync(ctx context.Context, req *business.SyncRequest) (*business.SyncResponse, error) {
	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, status.Errorf(codes.Internal, "Failed to start transaction: %v", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 同步时间
	syncTime := time.Now().Unix()

	// 处理账本
	var syncedLedgers []*business.Ledger
	for _, ledger := range req.Ledgers {
		var existingLedger Ledger
		result := tx.First(&existingLedger, "id = ?", ledger.Id)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// 创建新账本
				newLedger := Ledger{
					ID:          ledger.Id,
					Name:        ledger.Name,
					Description: ledger.Description,
					UserID:      req.UserId,
					Currency:    ledger.Currency,
				}

				if err := tx.Create(&newLedger).Error; err != nil {
					tx.Rollback()
					return nil, status.Errorf(codes.Internal, "Failed to create ledger: %v", err)
				}

				syncedLedgers = append(syncedLedgers, ledger)
			} else {
				tx.Rollback()
				return nil, status.Errorf(codes.Internal, "Failed to query ledger: %v", result.Error)
			}
		} else {
			// 更新现有账本
			existingLedger.Name = ledger.Name
			existingLedger.Description = ledger.Description
			existingLedger.Currency = ledger.Currency

			if err := tx.Save(&existingLedger).Error; err != nil {
				tx.Rollback()
				return nil, status.Errorf(codes.Internal, "Failed to update ledger: %v", err)
			}

			syncedLedgers = append(syncedLedgers, ledger)
		}
	}

	// 处理交易
	var syncedTransactions []*business.Transaction
	for _, transaction := range req.Transactions {
		var existingTransaction Transaction
		result := tx.First(&existingTransaction, "id = ?", transaction.Id)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				// 创建新交易
				newTransaction := Transaction{
					ID:              transaction.Id,
					LedgerID:        transaction.LedgerId,
					UserID:          req.UserId,
					Type:            transaction.Type,
					CategoryID:      transaction.CategoryId,
					SubcategoryID:   transaction.SubcategoryId,
					AccountID:       transaction.AccountId,
					TargetAccountID: transaction.TargetAccountId,
					Amount:          transaction.Amount,
					Description:     transaction.Description,
					Date:            transaction.Date,
					Tags:            transaction.Tags,
					SyncTime:        syncTime,
					DeviceID:        req.DeviceId,
				}

				if err := tx.Create(&newTransaction).Error; err != nil {
					tx.Rollback()
					return nil, status.Errorf(codes.Internal, "Failed to create transaction: %v", err)
				}

				syncedTransactions = append(syncedTransactions, transaction)
			} else {
				tx.Rollback()
				return nil, status.Errorf(codes.Internal, "Failed to query transaction: %v", result.Error)
			}
		} else {
			// 更新现有交易
			existingTransaction.Type = transaction.Type
			existingTransaction.CategoryID = transaction.CategoryId
			existingTransaction.SubcategoryID = transaction.SubcategoryId
			existingTransaction.AccountID = transaction.AccountId
			existingTransaction.TargetAccountID = transaction.TargetAccountId
			existingTransaction.Amount = transaction.Amount
			existingTransaction.Description = transaction.Description
			existingTransaction.Date = transaction.Date
			existingTransaction.Tags = transaction.Tags
			existingTransaction.SyncTime = syncTime

			if err := tx.Save(&existingTransaction).Error; err != nil {
				tx.Rollback()
				return nil, status.Errorf(codes.Internal, "Failed to update transaction: %v", err)
			}

			syncedTransactions = append(syncedTransactions, transaction)
		}
	}

	// 查询需要同步的新增或更新的账本
	var ledgers []Ledger
	if err := tx.Where("user_id = ? AND updated_at > ?", req.UserId, time.Unix(req.LastSyncTime, 0)).Find(&ledgers).Error; err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "Failed to query ledgers: %v", err)
	}

	// 查询需要同步的新增或更新的交易
	var transactions []Transaction
	if err := tx.Where("user_id = ? AND sync_time > ?", req.UserId, req.LastSyncTime).Find(&transactions).Error; err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "Failed to query transactions: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to commit transaction: %v", err)
	}

	// 转换为proto响应格式
	var responseLedgers []*business.Ledger
	for _, ledger := range ledgers {
		responseLedgers = append(responseLedgers, &business.Ledger{
			Id:          ledger.ID,
			Name:        ledger.Name,
			Description: ledger.Description,
			UserId:      ledger.UserID,
			Currency:    ledger.Currency,
			CreatedAt:   ledger.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   ledger.UpdatedAt.Format(time.RFC3339),
		})
	}

	var responseTransactions []*business.Transaction
	for _, transaction := range transactions {
		responseTransactions = append(responseTransactions, &business.Transaction{
			Id:              transaction.ID,
			LedgerId:        transaction.LedgerID,
			UserId:          transaction.UserID,
			Type:            transaction.Type,
			CategoryId:      transaction.CategoryID,
			SubcategoryId:   transaction.SubcategoryID,
			AccountId:       transaction.AccountID,
			TargetAccountId: transaction.TargetAccountID,
			Amount:          transaction.Amount,
			Description:     transaction.Description,
			Date:            transaction.Date,
			CreatedAt:       transaction.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       transaction.UpdatedAt.Format(time.RFC3339),
			Tags:            transaction.Tags,
		})
	}

	// 返回同步响应
	return &business.SyncResponse{
		SyncTime:              syncTime,
		Transactions:          responseTransactions,
		Ledgers:               responseLedgers,
		DeletedTransactionIds: []string{},
		DeletedLedgerIds:      []string{},
	}, nil
}

// GetLedgers 获取账本列表
func (s *BusinessService) GetLedgers(ctx context.Context, req *business.GetLedgersRequest) (*business.GetLedgersResponse, error) {
	// 设置默认分页
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 查询账本列表
	var ledgers []Ledger
	var total int64

	// 计算总数
	if err := s.db.Model(&Ledger{}).Where("user_id = ?", req.UserId).Count(&total).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to count ledgers: %v", err)
	}

	// 查询列表
	if err := s.db.Where("user_id = ?", req.UserId).Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&ledgers).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to query ledgers: %v", err)
	}

	// 转换为proto格式
	var responseLedgers []*business.Ledger
	for _, ledger := range ledgers {
		responseLedgers = append(responseLedgers, &business.Ledger{
			Id:          ledger.ID,
			Name:        ledger.Name,
			Description: ledger.Description,
			UserId:      ledger.UserID,
			Currency:    ledger.Currency,
			CreatedAt:   ledger.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   ledger.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &business.GetLedgersResponse{
		Ledgers:  responseLedgers,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// CreateLedger 创建账本
func (s *BusinessService) CreateLedger(ctx context.Context, req *business.Ledger) (*business.Ledger, error) {
	// 生成UUID
	ledgerID := req.Id
	if ledgerID == "" {
		ledgerID = uuid.New().String()
	}

	// 创建账本
	ledger := Ledger{
		ID:          ledgerID,
		Name:        req.Name,
		Description: req.Description,
		UserID:      req.UserId,
		Currency:    req.Currency,
	}

	if err := s.db.Create(&ledger).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create ledger: %v", err)
	}

	// 返回创建的账本
	return &business.Ledger{
		Id:          ledger.ID,
		Name:        ledger.Name,
		Description: ledger.Description,
		UserId:      ledger.UserID,
		Currency:    ledger.Currency,
		CreatedAt:   ledger.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   ledger.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// UpdateLedger 更新账本
func (s *BusinessService) UpdateLedger(ctx context.Context, req *business.Ledger) (*business.Ledger, error) {
	// 查询账本
	var ledger Ledger
	result := s.db.First(&ledger, "id = ?", req.Id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "Ledger not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to query ledger: %v", result.Error)
	}

	// 更新账本
	ledger.Name = req.Name
	ledger.Description = req.Description
	ledger.Currency = req.Currency

	if err := s.db.Save(&ledger).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update ledger: %v", err)
	}

	// 返回更新后的账本
	return &business.Ledger{
		Id:          ledger.ID,
		Name:        ledger.Name,
		Description: ledger.Description,
		UserId:      ledger.UserID,
		Currency:    ledger.Currency,
		CreatedAt:   ledger.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   ledger.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteLedger 删除账本
func (s *BusinessService) DeleteLedger(ctx context.Context, req *business.Ledger) (*common.Response, error) {
	// 删除账本
	result := s.db.Delete(&Ledger{}, "id = ?", req.Id)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete ledger: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "Ledger not found")
	}

	return &common.Response{
		Success: true,
		Message: "Ledger deleted successfully",
		Code:    200,
	}, nil
}

// CreateTransaction 创建交易
func (s *BusinessService) CreateTransaction(ctx context.Context, req *business.Transaction) (*business.Transaction, error) {
	// 生成UUID
	transactionID := req.Id
	if transactionID == "" {
		transactionID = uuid.New().String()
	}

	// 创建交易
	transaction := Transaction{
		ID:              transactionID,
		LedgerID:        req.LedgerId,
		UserID:          req.UserId,
		Type:            req.Type,
		CategoryID:      req.CategoryId,
		SubcategoryID:   req.SubcategoryId,
		AccountID:       req.AccountId,
		TargetAccountID: req.TargetAccountId,
		Amount:          req.Amount,
		Description:     req.Description,
		Date:            req.Date,
		Tags:            req.Tags,
		SyncTime:        time.Now().Unix(),
	}

	if err := s.db.Create(&transaction).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create transaction: %v", err)
	}

	// 返回创建的交易
	return &business.Transaction{
		Id:              transaction.ID,
		LedgerId:        transaction.LedgerID,
		UserId:          transaction.UserID,
		Type:            transaction.Type,
		CategoryId:      transaction.CategoryID,
		SubcategoryId:   transaction.SubcategoryID,
		AccountId:       transaction.AccountID,
		TargetAccountId: transaction.TargetAccountID,
		Amount:          transaction.Amount,
		Description:     transaction.Description,
		Date:            transaction.Date,
		CreatedAt:       transaction.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       transaction.UpdatedAt.Format(time.RFC3339),
		Tags:            transaction.Tags,
	}, nil
}

// UpdateTransaction 更新交易
func (s *BusinessService) UpdateTransaction(ctx context.Context, req *business.Transaction) (*business.Transaction, error) {
	// 查询交易
	var transaction Transaction
	result := s.db.First(&transaction, "id = ?", req.Id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "Transaction not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to query transaction: %v", result.Error)
	}

	// 更新交易
	transaction.LedgerID = req.LedgerId
	transaction.Type = req.Type
	transaction.CategoryID = req.CategoryId
	transaction.SubcategoryID = req.SubcategoryId
	transaction.AccountID = req.AccountId
	transaction.TargetAccountID = req.TargetAccountId
	transaction.Amount = req.Amount
	transaction.Description = req.Description
	transaction.Date = req.Date
	transaction.Tags = req.Tags
	transaction.SyncTime = time.Now().Unix()

	if err := s.db.Save(&transaction).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update transaction: %v", err)
	}

	// 返回更新后的交易
	return &business.Transaction{
		Id:              transaction.ID,
		LedgerId:        transaction.LedgerID,
		UserId:          transaction.UserID,
		Type:            transaction.Type,
		CategoryId:      transaction.CategoryID,
		SubcategoryId:   transaction.SubcategoryID,
		AccountId:       transaction.AccountID,
		TargetAccountId: transaction.TargetAccountID,
		Amount:          transaction.Amount,
		Description:     transaction.Description,
		Date:            transaction.Date,
		CreatedAt:       transaction.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       transaction.UpdatedAt.Format(time.RFC3339),
		Tags:            transaction.Tags,
	}, nil
}

// DeleteTransaction 删除交易
func (s *BusinessService) DeleteTransaction(ctx context.Context, req *business.Transaction) (*common.Response, error) {
	// 删除交易
	result := s.db.Delete(&Transaction{}, "id = ?", req.Id)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete transaction: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "Transaction not found")
	}

	return &common.Response{
		Success: true,
		Message: "Transaction deleted successfully",
		Code:    200,
	}, nil
}

// Check 健康检查
func (s *BusinessService) Check(ctx context.Context, req *common.HealthCheckRequest) (*common.HealthCheckResponse, error) {
	// 检查数据库连接
	sqlDB, err := s.db.DB()
	if err != nil {
		return &common.HealthCheckResponse{
			Status: common.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	// 执行简单查询
	if err := sqlDB.Ping(); err != nil {
		return &common.HealthCheckResponse{
			Status: common.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	return &common.HealthCheckResponse{
		Status: common.HealthCheckResponse_SERVING,
	}, nil
}

// Watch 健康检查监听
func (s *BusinessService) Watch(req *common.HealthCheckRequest, stream common.HealthCheckService_WatchServer) error {
	// 实现健康检查监听逻辑
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}
