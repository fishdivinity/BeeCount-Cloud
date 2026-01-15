package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/internal/config"
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库接口
// 定义数据库的通用方法
type Database interface {
	// Connect 连接数据库
	Connect() error
	// Close 关闭数据库连接
	Close() error
	// AutoMigrate 自动迁移数据库表结构
	AutoMigrate(models ...interface{}) error
	// GetDB 获取GORM数据库实例
	GetDB() *gorm.DB
	// TestConnection 测试数据库连接
	TestConnection() error
}

// GormDatabase GORM数据库实现
type GormDatabase struct {
	db     *gorm.DB
	config *config.DatabaseConfig
}

// NewDatabase 创建数据库实例
// 根据配置类型返回对应的数据库实现
func NewDatabase(cfg *config.DatabaseConfig) (Database, error) {
	db := &GormDatabase{
		config: cfg,
	}

	if err := db.Connect(); err != nil {
		return nil, err
	}

	return db, nil
}

// NewDatabaseWithType 根据指定类型创建数据库实例
func NewDatabaseWithType(dbType string, cfg *config.DatabaseConfig) (Database, error) {
	tempConfig := *cfg
	tempConfig.Type = dbType

	db := &GormDatabase{
		config: &tempConfig,
	}

	if err := db.Connect(); err != nil {
		return nil, err
	}

	return db, nil
}

// Connect 连接数据库
func (d *GormDatabase) Connect() error {
	var dialector gorm.Dialector
	var dsn string

	dbType := d.config.Type
	if dbType == "" {
		dbType = d.config.Active
	}

	switch dbType {
	case "sqlite":
		dsn = d.config.SQLite.Path

		absPath, err := filepath.Abs(dsn)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}

		dialector = sqlite.Open(dsn)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
			d.config.MySQL.Username,
			d.config.MySQL.Password,
			d.config.MySQL.Host,
			d.config.MySQL.Port,
			d.config.MySQL.Database,
			d.config.MySQL.Charset,
			d.config.MySQL.ParseTime,
			d.config.MySQL.Loc,
		)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s timezone=%s",
			d.config.Postgres.Host,
			d.config.Postgres.Port,
			d.config.Postgres.Username,
			d.config.Postgres.Password,
			d.config.Postgres.Database,
			d.config.Postgres.SSLMode,
			d.config.Postgres.Timezone,
		)
		dialector = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	d.db = db
	return nil
}

// Close 关闭数据库连接
func (d *GormDatabase) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AutoMigrate 自动迁移数据库表结构
func (d *GormDatabase) AutoMigrate(models ...interface{}) error {
	return d.db.AutoMigrate(models...)
}

// GetDB 获取GORM数据库实例
func (d *GormDatabase) GetDB() *gorm.DB {
	return d.db
}

// TestConnection 测试数据库连接
func (d *GormDatabase) TestConnection() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// AutoMigrateModels 自动迁移所有模型
func AutoMigrateModels(db Database) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Ledger{},
		&models.Account{},
		&models.Category{},
		&models.Transaction{},
		&models.Tag{},
		&models.TransactionTag{},
		&models.TransactionAttachment{},
		&models.Budget{},
		&models.RecurringTransaction{},
		&models.Conversation{},
		&models.Message{},
	)
}
