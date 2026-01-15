package repository

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// ledgerRepository 账本仓储实现
type ledgerRepository struct {
	db *gorm.DB
}

// NewLedgerRepository 创建账本仓储实例
func NewLedgerRepository(db *gorm.DB) LedgerRepository {
	return &ledgerRepository{db: db}
}

// Create 创建账本
func (r *ledgerRepository) Create(ledger *models.Ledger) error {
	return r.db.Create(ledger).Error
}

// GetByID 根据ID获取账本
func (r *ledgerRepository) GetByID(id uint) (*models.Ledger, error) {
	var ledger models.Ledger
	err := r.db.Preload("User").First(&ledger, id).Error
	if err != nil {
		return nil, err
	}
	return &ledger, nil
}

// GetByUserID 根据用户ID获取账本列表
func (r *ledgerRepository) GetByUserID(userID uint) ([]models.Ledger, error) {
	var ledgers []models.Ledger
	err := r.db.Where("user_id = ?", userID).Find(&ledgers).Error
	return ledgers, err
}

// Update 更新账本
func (r *ledgerRepository) Update(ledger *models.Ledger) error {
	return r.db.Save(ledger).Error
}

// Delete 删除账本
func (r *ledgerRepository) Delete(id uint) error {
	return r.db.Delete(&models.Ledger{}, id).Error
}

