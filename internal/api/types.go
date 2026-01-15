package api

import (
	_ "github.com/fishdivinity/BeeCount-Cloud/docs/swagger"
)

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"johndoe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  UserResponse `json:"user"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	ID          uint   `json:"id" example:"1"`
	Username    string `json:"username" example:"johndoe"`
	Email       string `json:"email" example:"john@example.com"`
	DisplayName string `json:"display_name" example:"John Doe"`
	Avatar      string `json:"avatar" example:"https://example.com/avatar.jpg"`
	CreatedAt  string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt  string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// CreateLedgerRequest 创建账本请求
type CreateLedgerRequest struct {
	Name     string `json:"name" binding:"required,max=100" example:"My Ledger"`
	Currency string `json:"currency" binding:"omitempty,len=3" example:"CNY"`
	Type     string `json:"type" binding:"omitempty,oneof=personal shared" example:"personal"`
}

// UpdateLedgerRequest 更新账本请求
type UpdateLedgerRequest struct {
	Name     *string `json:"name" binding:"omitempty,max=100" example:"Updated Ledger"`
	Currency *string `json:"currency" binding:"omitempty,len=3" example:"USD"`
	Type     *string `json:"type" binding:"omitempty,oneof=personal shared" example:"shared"`
}

// LedgerResponse 账本响应
type LedgerResponse struct {
	ID        uint   `json:"id" example:"1"`
	UserID    uint   `json:"user_id" example:"1"`
	Name      string `json:"name" example:"My Ledger"`
	Currency  string `json:"currency" example:"CNY"`
	Type      string `json:"type" example:"personal"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// CreateTransactionRequest 创建交易请求
type CreateTransactionRequest struct {
	LedgerID    uint   `json:"ledger_id" binding:"required" example:"1"`
	Type         string `json:"type" binding:"required,oneof=expense income transfer" example:"expense"`
	Amount       float64 `json:"amount" binding:"required,gt=0" example:"100.50"`
	CategoryID   *uint  `json:"category_id" example:"1"`
	AccountID    *uint  `json:"account_id" example:"1"`
	ToAccountID  *uint  `json:"to_account_id" example:"2"`
	HappenedAt   string `json:"happened_at" binding:"required" example:"2024-01-01T12:00:00Z"`
	Note         string `json:"note" example:"Lunch"`
}

// UpdateTransactionRequest 更新交易请求
type UpdateTransactionRequest struct {
	Type         *string  `json:"type" binding:"omitempty,oneof=expense income transfer" example:"income"`
	Amount       *float64 `json:"amount" binding:"omitempty,gt=0" example:"200.00"`
	CategoryID   *uint    `json:"category_id" example:"2"`
	AccountID    *uint    `json:"account_id" example:"1"`
	ToAccountID  *uint    `json:"to_account_id" example:"2"`
	HappenedAt   *string  `json:"happened_at" example:"2024-01-02T12:00:00Z"`
	Note         *string  `json:"note" example:"Updated note"`
}

// TransactionResponse 交易响应
type TransactionResponse struct {
	ID           uint                 `json:"id" example:"1"`
	UserID       uint                 `json:"user_id" example:"1"`
	LedgerID     uint                 `json:"ledger_id" example:"1"`
	Type         string               `json:"type" example:"expense"`
	Amount       float64              `json:"amount" example:"100.50"`
	CategoryID   *uint                `json:"category_id" example:"1"`
	AccountID    *uint                `json:"account_id" example:"1"`
	ToAccountID  *uint                `json:"to_account_id" example:"2"`
	HappenedAt   string               `json:"happened_at" example:"2024-01-01T12:00:00Z"`
	Note         string               `json:"note" example:"Lunch"`
	RecurringID  *uint                `json:"recurring_id"`
	CreatedAt    string               `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt    string               `json:"updated_at" example:"2024-01-01T12:00:00Z"`
	Category     *CategoryResponse      `json:"category"`
	Account      *AccountResponse       `json:"account"`
	ToAccount    *AccountResponse       `json:"to_account"`
	Tags         []TagResponse        `json:"tags"`
	Attachments  []AttachmentResponse  `json:"attachments"`
}

// CategoryResponse 分类响应
type CategoryResponse struct {
	ID              uint                `json:"id" example:"1"`
	UserID          uint                `json:"user_id" example:"1"`
	Name            string              `json:"name" example:"Food"`
	Kind            string              `json:"kind" example:"expense"`
	Icon            string              `json:"icon" example:"restaurant"`
	SortOrder       int                 `json:"sort_order" example:"0"`
	ParentID        *uint               `json:"parent_id" example:"1"`
	Level           int                 `json:"level" example:"2"`
	IconType        string              `json:"icon_type" example:"material"`
	CustomIconPath  string              `json:"custom_icon_path" example:"/custom/icons/food.svg"`
	CommunityIconID string              `json:"community_icon_id" example:"community_123"`
	CreatedAt       string              `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt       string              `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// AccountResponse 账户响应
type AccountResponse struct {
	ID             uint   `json:"id" example:"1"`
	UserID         uint   `json:"user_id" example:"1"`
	Name           string `json:"name" example:"Cash"`
	Type           string `json:"type" example:"cash"`
	Currency       string `json:"currency" example:"CNY"`
	InitialBalance float64 `json:"initial_balance" example:"1000.00"`
	CreatedAt      string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt      string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// TagResponse 标签响应
type TagResponse struct {
	ID        uint   `json:"id" example:"1"`
	UserID    uint   `json:"user_id" example:"1"`
	Name      string `json:"name" example:"Business"`
	Color     string `json:"color" example:"#FF5722"`
	SortOrder int    `json:"sort_order" example:"0"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// AttachmentResponse 附件响应
type AttachmentResponse struct {
	ID           uint   `json:"id" example:"1"`
	TransactionID uint   `json:"transaction_id" example:"1"`
	FileName      string `json:"file_name" example:"attachments/1/1_photo.jpg"`
	OriginalName  string `json:"original_name" example:"photo.jpg"`
	FileSize      *int   `json:"file_size" example:"102400"`
	Width         *int   `json:"width" example:"1920"`
	Height        *int   `json:"height" example:"1080"`
	SortOrder     int    `json:"sort_order" example:"0"`
	CreatedAt     string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt     string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error" example:"bad request"`
	Message string `json:"message" example:"Invalid input"`
	Code    int    `json:"code" example:"400"`
}

