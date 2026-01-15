package repository

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// categoryRepository 分类仓储实现
type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建分类仓储实例
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// Create 创建分类
func (r *categoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

// GetByID 根据ID获取分类
func (r *categoryRepository) GetByID(id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.Preload("Children").First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetByUserID 根据用户ID获取分类列表
func (r *categoryRepository) GetByUserID(userID uint) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Where("user_id = ?", userID).Find(&categories).Error
	return categories, err
}

// Update 更新分类
func (r *categoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

// Delete 删除分类
func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}

// GetChildren 获取子分类
func (r *categoryRepository) GetChildren(parentID uint) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Where("parent_id = ?", parentID).Find(&categories).Error
	return categories, err
}

