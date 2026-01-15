package repository

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// messageRepository 消息仓储实现
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息仓储实例
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

// Create 创建消息
func (r *messageRepository) Create(msg *models.Message) error {
	return r.db.Create(msg).Error
}

// GetByID 根据ID获取消息
func (r *messageRepository) GetByID(id uint) (*models.Message, error) {
	var msg models.Message
	err := r.db.First(&msg, id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetByConversationID 根据对话ID获取消息列表
func (r *messageRepository) GetByConversationID(conversationID uint) ([]models.Message, error) {
	var msgs []models.Message
	err := r.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&msgs).Error
	return msgs, err
}

// Update 更新消息
func (r *messageRepository) Update(msg *models.Message) error {
	return r.db.Save(msg).Error
}

// Delete 删除消息
func (r *messageRepository) Delete(id uint) error {
	return r.db.Delete(&models.Message{}, id).Error
}

