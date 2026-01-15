package repository

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// tagRepository 标签仓储实现
type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository 创建标签仓储实例
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

// Create 创建标签
func (r *tagRepository) Create(tag *models.Tag) error {
	return r.db.Create(tag).Error
}

// GetByID 根据ID获取标签
func (r *tagRepository) GetByID(id uint) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetByUserID 根据用户ID获取标签列表
func (r *tagRepository) GetByUserID(userID uint) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.Where("user_id = ?", userID).Find(&tags).Error
	return tags, err
}

// Update 更新标签
func (r *tagRepository) Update(tag *models.Tag) error {
	return r.db.Save(tag).Error
}

// Delete 删除标签
func (r *tagRepository) Delete(id uint) error {
	return r.db.Delete(&models.Tag{}, id).Error
}

