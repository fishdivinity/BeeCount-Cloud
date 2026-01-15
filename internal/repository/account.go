package repository

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// accountRepository 账户仓储实现
type accountRepository struct {
	db *gorm.DB
}

// NewAccountRepository 创建账户仓储实例
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepository{db: db}
}

// Create 创建账户
func (r *accountRepository) Create(account *models.Account) error {
	return r.db.Create(account).Error
}

// GetByID 根据ID获取账户
func (r *accountRepository) GetByID(id uint) (*models.Account, error) {
	var account models.Account
	err := r.db.First(&account, id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetByUserID 根据用户ID获取账户列表
func (r *accountRepository) GetByUserID(userID uint) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.Where("user_id = ?", userID).Find(&accounts).Error
	return accounts, err
}

// Update 更新账户
func (r *accountRepository) Update(account *models.Account) error {
	return r.db.Save(account).Error
}

// Delete 删除账户
func (r *accountRepository) Delete(id uint) error {
	return r.db.Delete(&models.Account{}, id).Error
}

