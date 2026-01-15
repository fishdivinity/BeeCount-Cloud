package repository

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// UserRepository 用户仓储接口
// 定义用户数据访问的方法
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	List(limit, offset int) ([]models.User, error)
}

// LedgerRepository 账本仓储接口
// 定义账本数据访问的方法
type LedgerRepository interface {
	Create(ledger *models.Ledger) error
	GetByID(id uint) (*models.Ledger, error)
	GetByUserID(userID uint) ([]models.Ledger, error)
	Update(ledger *models.Ledger) error
	Delete(id uint) error
}

// TransactionRepository 交易仓储接口
// 定义交易数据访问的方法
type TransactionRepository interface {
	Create(tx *models.Transaction) error
	GetByID(id uint) (*models.Transaction, error)
	GetByLedgerID(ledgerID uint, limit, offset int) ([]models.Transaction, error)
	Update(tx *models.Transaction) error
	Delete(id uint) error
	GetByDateRange(ledgerID uint, startDate, endDate string) ([]models.Transaction, error)
}

// CategoryRepository 分类仓储接口
// 定义分类数据访问的方法
type CategoryRepository interface {
	Create(category *models.Category) error
	GetByID(id uint) (*models.Category, error)
	GetByUserID(userID uint) ([]models.Category, error)
	Update(category *models.Category) error
	Delete(id uint) error
	GetChildren(parentID uint) ([]models.Category, error)
}

// AccountRepository 账户仓储接口
// 定义账户数据访问的方法
type AccountRepository interface {
	Create(account *models.Account) error
	GetByID(id uint) (*models.Account, error)
	GetByUserID(userID uint) ([]models.Account, error)
	Update(account *models.Account) error
	Delete(id uint) error
}

// TagRepository 标签仓储接口
// 定义标签数据访问的方法
type TagRepository interface {
	Create(tag *models.Tag) error
	GetByID(id uint) (*models.Tag, error)
	GetByUserID(userID uint) ([]models.Tag, error)
	Update(tag *models.Tag) error
	Delete(id uint) error
}

// AttachmentRepository 附件仓储接口
// 定义附件数据访问的方法
type AttachmentRepository interface {
	Create(attachment *models.TransactionAttachment) error
	GetByID(id uint) (*models.TransactionAttachment, error)
	GetByTransactionID(transactionID uint) ([]models.TransactionAttachment, error)
	Update(attachment *models.TransactionAttachment) error
	Delete(id uint) error
}

// BudgetRepository 预算仓储接口
// 定义预算数据访问的方法
type BudgetRepository interface {
	Create(budget *models.Budget) error
	GetByID(id uint) (*models.Budget, error)
	GetByLedgerID(ledgerID uint) ([]models.Budget, error)
	Update(budget *models.Budget) error
	Delete(id uint) error
}

// RecurringTransactionRepository 周期交易仓储接口
// 定义周期交易数据访问的方法
type RecurringTransactionRepository interface {
	Create(rt *models.RecurringTransaction) error
	GetByID(id uint) (*models.RecurringTransaction, error)
	GetByLedgerID(ledgerID uint) ([]models.RecurringTransaction, error)
	Update(rt *models.RecurringTransaction) error
	Delete(id uint) error
	GetEnabled(ledgerID uint) ([]models.RecurringTransaction, error)
}

// ConversationRepository 对话仓储接口
// 定义对话数据访问的方法
type ConversationRepository interface {
	Create(conv *models.Conversation) error
	GetByID(id uint) (*models.Conversation, error)
	GetByUserID(userID uint) ([]models.Conversation, error)
	Update(conv *models.Conversation) error
	Delete(id uint) error
}

// MessageRepository 消息仓储接口
// 定义消息数据访问的方法
type MessageRepository interface {
	Create(msg *models.Message) error
	GetByID(id uint) (*models.Message, error)
	GetByConversationID(conversationID uint) ([]models.Message, error)
	Update(msg *models.Message) error
	Delete(id uint) error
}

// Repositories 仓储集合
// 包含所有仓储的实例
type Repositories struct {
	User                 UserRepository
	Ledger               LedgerRepository
	Transaction          TransactionRepository
	Category             CategoryRepository
	Account              AccountRepository
	Tag                  TagRepository
	Attachment           AttachmentRepository
	Budget               BudgetRepository
	RecurringTransaction RecurringTransactionRepository
	Conversation         ConversationRepository
	Message              MessageRepository
}

// NewRepositories 创建所有仓储实例
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:                 NewUserRepository(db),
		Ledger:               NewLedgerRepository(db),
		Transaction:          NewTransactionRepository(db),
		Category:             NewCategoryRepository(db),
		Account:              NewAccountRepository(db),
		Tag:                  NewTagRepository(db),
		Attachment:           NewAttachmentRepository(db),
		Budget:               NewBudgetRepository(db),
		RecurringTransaction: NewRecurringTransactionRepository(db),
		Conversation:         NewConversationRepository(db),
		Message:              NewMessageRepository(db),
	}
}

