package repository

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// transactionRepository 交易仓储实现
type transactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository 创建交易仓储实例
func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

// Create 创建交易
func (r *transactionRepository) Create(tx *models.Transaction) error {
	return r.db.Create(tx).Error
}

// GetByID 根据ID获取交易
func (r *transactionRepository) GetByID(id uint) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.
		Preload("Category").
		Preload("Account").
		Preload("ToAccount").
		Preload("Tags").
		Preload("Attachments").
		First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// GetByLedgerID 根据账本ID获取交易列表
func (r *transactionRepository) GetByLedgerID(ledgerID uint, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.
		Preload("Category").
		Preload("Account").
		Preload("ToAccount").
		Preload("Tags").
		Preload("Attachments").
		Where("ledger_id = ?", ledgerID).
		Order("happened_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// Update 更新交易
func (r *transactionRepository) Update(tx *models.Transaction) error {
	return r.db.Save(tx).Error
}

// Delete 删除交易
func (r *transactionRepository) Delete(id uint) error {
	return r.db.Delete(&models.Transaction{}, id).Error
}

// GetByDateRange 根据日期范围获取交易
func (r *transactionRepository) GetByDateRange(ledgerID uint, startDate, endDate string) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.
		Preload("Category").
		Preload("Account").
		Preload("Tags").
		Where("ledger_id = ? AND happened_at >= ? AND happened_at <= ?", ledgerID, startDate, endDate).
		Order("happened_at DESC").
		Find(&transactions).Error
	return transactions, err
}

