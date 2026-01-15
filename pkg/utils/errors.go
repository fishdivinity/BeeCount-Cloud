package utils

import (
	"errors"
	"fmt"
)

// 预定义错误
var (
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidCredentials 无效的凭证
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrLedgerNotFound 账本不存在
	ErrLedgerNotFound = errors.New("ledger not found")
	// ErrTransactionNotFound 交易不存在
	ErrTransactionNotFound = errors.New("transaction not found")
	// ErrNotFound 资源不存在
	ErrNotFound = errors.New("resource not found")
	// ErrForbidden 权限不足
	ErrForbidden = errors.New("access forbidden")
	// ErrBadRequest 请求参数错误
	ErrBadRequest = errors.New("bad request")
	// ErrUnauthorized 未授权
	ErrUnauthorized = errors.New("unauthorized")
)

// AppError 应用错误
type AppError struct {
	Code    int
	Message string
	Err     error
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewBadRequestError 创建400错误
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:    400,
		Message: message,
	}
}

// NewUnauthorizedError 创建401错误
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    401,
		Message: message,
	}
}

// NewForbiddenError 创建403错误
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:    403,
		Message: message,
	}
}

// NewNotFoundError 创建404错误
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:    404,
		Message: message,
	}
}

// NewInternalServerError 创建500错误
func NewInternalServerError(message string) *AppError {
	return &AppError{
		Code:    500,
		Message: message,
	}
}

// WrapError 包装错误
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

