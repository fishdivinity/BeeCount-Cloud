package repository

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// budgetRepository 预算仓储实现
type budgetRepository struct {
	db *gorm.DB
}

// NewBudgetRepository 创建预算仓储实例
func NewBudgetRepository(db *gorm.DB) BudgetRepository {
	return &budgetRepository{db: db}
}

// Create 创建预算
func (r *budgetRepository) Create(budget *models.Budget) error {
	return r.db.Create(budget).Error
}

// GetByID 根据ID获取预算
func (r *budgetRepository) GetByID(id uint) (*models.Budget, error) {
	var budget models.Budget
	err := r.db.Preload("Category").First(&budget, id).Error
	if err != nil {
		return nil, err
	}
	return &budget, nil
}

// GetByLedgerID 根据账本ID获取预算列表
func (r *budgetRepository) GetByLedgerID(ledgerID uint) ([]models.Budget, error) {
	var budgets []models.Budget
	err := r.db.Preload("Category").Where("ledger_id = ?", ledgerID).Find(&budgets).Error
	return budgets, err
}

// Update 更新预算
func (r *budgetRepository) Update(budget *models.Budget) error {
	return r.db.Save(budget).Error
}

// Delete 删除预算
func (r *budgetRepository) Delete(id uint) error {
	return r.db.Delete(&models.Budget{}, id).Error
}

