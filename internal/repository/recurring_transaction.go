package repository

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// recurringTransactionRepository 周期交易仓储实现
type recurringTransactionRepository struct {
	db *gorm.DB
}

// NewRecurringTransactionRepository 创建周期交易仓储实例
func NewRecurringTransactionRepository(db *gorm.DB) RecurringTransactionRepository {
	return &recurringTransactionRepository{db: db}
}

// Create 创建周期交易
func (r *recurringTransactionRepository) Create(rt *models.RecurringTransaction) error {
	return r.db.Create(rt).Error
}

// GetByID 根据ID获取周期交易
func (r *recurringTransactionRepository) GetByID(id uint) (*models.RecurringTransaction, error) {
	var rt models.RecurringTransaction
	err := r.db.Preload("Category").Preload("Account").First(&rt, id).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// GetByLedgerID 根据账本ID获取周期交易列表
func (r *recurringTransactionRepository) GetByLedgerID(ledgerID uint) ([]models.RecurringTransaction, error) {
	var rts []models.RecurringTransaction
	err := r.db.Preload("Category").Preload("Account").Where("ledger_id = ?", ledgerID).Find(&rts).Error
	return rts, err
}

// Update 更新周期交易
func (r *recurringTransactionRepository) Update(rt *models.RecurringTransaction) error {
	return r.db.Save(rt).Error
}

// Delete 删除周期交易
func (r *recurringTransactionRepository) Delete(id uint) error {
	return r.db.Delete(&models.RecurringTransaction{}, id).Error
}

// GetEnabled 获取启用的周期交易
func (r *recurringTransactionRepository) GetEnabled(ledgerID uint) ([]models.RecurringTransaction, error) {
	var rts []models.RecurringTransaction
	err := r.db.Preload("Category").Preload("Account").Where("ledger_id = ? AND enabled = ?", ledgerID, true).Find(&rts).Error
	return rts, err
}

