package repository

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// conversationRepository 对话仓储实现
type conversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository 创建对话仓储实例
func NewConversationRepository(db *gorm.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

// Create 创建对话
func (r *conversationRepository) Create(conv *models.Conversation) error {
	return r.db.Create(conv).Error
}

// GetByID 根据ID获取对话
func (r *conversationRepository) GetByID(id uint) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Preload("Messages").First(&conv, id).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// GetByUserID 根据用户ID获取对话列表
func (r *conversationRepository) GetByUserID(userID uint) ([]models.Conversation, error) {
	var convs []models.Conversation
	err := r.db.Where("user_id = ?", userID).Order("updated_at DESC").Find(&convs).Error
	return convs, err
}

// Update 更新对话
func (r *conversationRepository) Update(conv *models.Conversation) error {
	return r.db.Save(conv).Error
}

// Delete 删除对话
func (r *conversationRepository) Delete(id uint) error {
	return r.db.Delete(&models.Conversation{}, id).Error
}

