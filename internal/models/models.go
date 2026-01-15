package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username     string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	DisplayName  string `gorm:"size:100" json:"display_name"`
	Avatar       string `gorm:"size:255" json:"avatar"`
}

// Ledger 账本模型
type Ledger struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID   uint   `gorm:"index;not null" json:"user_id"`
	Name      string `gorm:"size:100;not null" json:"name"`
	Currency  string `gorm:"size:10;default:'CNY';not null" json:"currency"`
	Type      string `gorm:"size:20;default:'personal';not null" json:"type"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

// Account 账户模型
type Account struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID        uint    `gorm:"index;not null" json:"user_id"`
	Name          string  `gorm:"size:100;not null" json:"name"`
	Type          string  `gorm:"size:50;default:'cash';not null" json:"type"`
	Currency      string  `gorm:"size:10;default:'CNY';not null" json:"currency"`
	InitialBalance float64 `gorm:"type:decimal(15,2);default:0.00" json:"initial_balance"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

// Category 分类模型
type Category struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID          uint   `gorm:"index;not null" json:"user_id"`
	Name            string `gorm:"size:100;not null" json:"name"`
	Kind            string `gorm:"size:20;not null" json:"kind"`
	Icon            string `gorm:"size:50" json:"icon"`
	SortOrder       int    `gorm:"default:0" json:"sort_order"`
	ParentID        *uint  `gorm:"index" json:"parent_id"`
	Level           int    `gorm:"default:1" json:"level"`
	IconType        string `gorm:"size:20;default:'material'" json:"icon_type"`
	CustomIconPath  string `gorm:"size:255" json:"custom_icon_path"`
	CommunityIconID string `gorm:"size:50" json:"community_icon_id"`

	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Parent   *Category `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// Transaction 交易模型
type Transaction struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID       uint      `gorm:"index;not null" json:"user_id"`
	LedgerID     uint      `gorm:"index;not null" json:"ledger_id"`
	Type         string    `gorm:"size:20;not null" json:"type"`
	Amount       float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	CategoryID   *uint     `gorm:"index" json:"category_id"`
	AccountID    *uint     `gorm:"index" json:"account_id"`
	ToAccountID  *uint     `gorm:"index" json:"to_account_id"`
	HappenedAt   time.Time `gorm:"not null" json:"happened_at"`
	Note         string    `gorm:"type:text" json:"note"`
	RecurringID  *uint     `gorm:"index" json:"recurring_id"`

	User       User       `gorm:"foreignKey:UserID" json:"-"`
	Ledger     Ledger     `gorm:"foreignKey:LedgerID" json:"ledger,omitempty"`
	Category   *Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Account    *Account   `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	ToAccount  *Account   `gorm:"foreignKey:ToAccountID" json:"to_account,omitempty"`
	Recurring *RecurringTransaction `gorm:"foreignKey:RecurringID" json:"recurring,omitempty"`
	Tags       []Tag      `gorm:"many2many:transaction_tags;" json:"tags,omitempty"`
	Attachments []TransactionAttachment `gorm:"foreignKey:TransactionID" json:"attachments,omitempty"`
}

// RecurringTransaction 周期交易模型
type RecurringTransaction struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID       uint       `gorm:"index;not null" json:"user_id"`
	LedgerID     uint       `gorm:"index;not null" json:"ledger_id"`
	Type         string     `gorm:"size:20;not null" json:"type"`
	Amount       float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	CategoryID   *uint      `gorm:"index" json:"category_id"`
	AccountID    *uint      `gorm:"index" json:"account_id"`
	ToAccountID  *uint      `gorm:"index" json:"to_account_id"`
	Note         string     `gorm:"type:text" json:"note"`
	Frequency    string     `gorm:"size:20;not null" json:"frequency"`
	Interval     int        `gorm:"default:1" json:"interval"`
	DayOfMonth   *int       `json:"day_of_month"`
	DayOfWeek    *int       `json:"day_of_week"`
	MonthOfYear   *int       `json:"month_of_year"`
	StartDate    time.Time  `gorm:"not null" json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	LastGeneratedDate *time.Time `json:"last_generated_date"`
	Enabled      bool       `gorm:"default:true" json:"enabled"`

	User     User     `gorm:"foreignKey:UserID" json:"-"`
	Ledger   Ledger   `gorm:"foreignKey:LedgerID" json:"ledger,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Account  *Account  `gorm:"foreignKey:AccountID" json:"account,omitempty"`
	ToAccount *Account  `gorm:"foreignKey:ToAccountID" json:"to_account,omitempty"`
}

// Tag 标签模型
type Tag struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID    uint   `gorm:"index;not null" json:"user_id"`
	Name      string `gorm:"size:100;not null" json:"name"`
	Color     string `gorm:"size:20" json:"color"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TransactionTag 交易标签关联模型
type TransactionTag struct {
	ID            uint `gorm:"primarykey" json:"id"`
	TransactionID uint `gorm:"index;not null" json:"transaction_id"`
	TagID         uint `gorm:"index;not null" json:"tag_id"`

	Transaction Transaction `gorm:"foreignKey:TransactionID" json:"-"`
	Tag         Tag         `gorm:"foreignKey:TagID" json:"-"`
}

// TransactionAttachment 交易附件模型
type TransactionAttachment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	TransactionID uint   `gorm:"index;not null" json:"transaction_id"`
	FileName      string `gorm:"size:255;not null" json:"file_name"`
	OriginalName  string `gorm:"size:255" json:"original_name"`
	FileSize      *int   `json:"file_size"`
	Width         *int   `json:"width"`
	Height        *int   `json:"height"`
	SortOrder     int    `gorm:"default:0" json:"sort_order"`

	Transaction Transaction `gorm:"foreignKey:TransactionID" json:"-"`
}

// Budget 预算模型
type Budget struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID    uint   `gorm:"index;not null" json:"user_id"`
	LedgerID  uint   `gorm:"index;not null" json:"ledger_id"`
	Type      string `gorm:"size:20;default:'total';not null" json:"type"`
	CategoryID *uint  `gorm:"index" json:"category_id"`
	Amount    float64 `gorm:"type:decimal(15,2);not null" json:"amount"`
	Period    string `gorm:"size:20;default:'monthly';not null" json:"period"`
	StartDay  int    `gorm:"default:1" json:"start_day"`
	Enabled   bool   `gorm:"default:true" json:"enabled"`

	User     User     `gorm:"foreignKey:UserID" json:"-"`
	Ledger   Ledger   `gorm:"foreignKey:LedgerID" json:"ledger,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// Conversation 对话模型
type Conversation struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uint   `gorm:"index;not null" json:"user_id"`
	Title  string `gorm:"size:200;default:'AI对话'" json:"title"`

	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Messages []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

// Message 消息模型
type Message struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ConversationID uint   `gorm:"index;not null" json:"conversation_id"`
	Role           string `gorm:"size:20;not null" json:"role"`
	Content        string `gorm:"type:text;not null" json:"content"`
	MessageType    string `gorm:"size:20;default:'text';not null" json:"message_type"`
	Metadata       string `gorm:"type:text" json:"metadata"`
	TransactionID  *uint  `gorm:"index" json:"transaction_id"`

	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"-"`
	Transaction  *Transaction  `gorm:"foreignKey:TransactionID" json:"transaction,omitempty"`
}

