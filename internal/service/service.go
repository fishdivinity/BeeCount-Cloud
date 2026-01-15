package service

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/auth"
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"github.com/fishdivinity/BeeCount-Cloud/internal/repository"
)

// UserService 用户服务接口
// 定义用户相关的业务逻辑方法
type UserService interface {
	Register(user *models.User) error
	Login(email, password string) (*models.User, string, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateUser(user *models.User) error
}

// LedgerService 账本服务接口
// 定义账本相关的业务逻辑方法
type LedgerService interface {
	CreateLedger(userID uint, ledger *models.Ledger) error
	GetLedger(id uint) (*models.Ledger, error)
	GetUserLedgers(userID uint) ([]models.Ledger, error)
	UpdateLedger(ledger *models.Ledger) error
	DeleteLedger(id uint) error
}

// TransactionService 交易服务接口
// 定义交易相关的业务逻辑方法
type TransactionService interface {
	CreateTransaction(userID uint, tx *models.Transaction) error
	GetTransaction(id uint) (*models.Transaction, error)
	GetLedgerTransactions(ledgerID uint, limit, offset int) ([]models.Transaction, error)
	UpdateTransaction(tx *models.Transaction) error
	DeleteTransaction(id uint) error
}

// SyncService 同步服务接口
// 定义数据同步相关的业务逻辑方法
type SyncService interface {
	UploadLedger(userID, ledgerID uint) (string, error)
	DownloadLedger(userID, ledgerID uint) (string, error)
	GetSyncStatus(userID, ledgerID uint) (map[string]interface{}, error)
}

// Services 服务集合
// 包含所有业务服务的实例
type Services struct {
	User        UserService
	Ledger      LedgerService
	Transaction TransactionService
	Sync        SyncService
}

// NewServices 创建所有服务实例
func NewServices(repos *repository.Repositories, authService auth.AuthService) *Services {
	return &Services{
		User:        NewUserService(repos.User, authService),
		Ledger:      NewLedgerService(repos.Ledger),
		Transaction: NewTransactionService(repos.Transaction),
		Sync:        NewSyncService(repos, authService),
	}
}

// AttachmentRepository 附件仓储接口（已废弃，应使用repository包中的接口）
type AttachmentRepository interface {
	Create(attachment *models.TransactionAttachment) error
	GetByID(id uint) (*models.TransactionAttachment, error)
	GetByTransactionID(transactionID uint) ([]models.TransactionAttachment, error)
	Update(attachment *models.TransactionAttachment) error
	Delete(id uint) error
}

