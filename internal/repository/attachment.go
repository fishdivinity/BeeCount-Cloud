package repository

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// attachmentRepository 附件仓储实现
type attachmentRepository struct {
	db *gorm.DB
}

// NewAttachmentRepository 创建附件仓储实例
func NewAttachmentRepository(db *gorm.DB) AttachmentRepository {
	return &attachmentRepository{db: db}
}

// Create 创建附件
func (r *attachmentRepository) Create(attachment *models.TransactionAttachment) error {
	return r.db.Create(attachment).Error
}

// GetByID 根据ID获取附件
func (r *attachmentRepository) GetByID(id uint) (*models.TransactionAttachment, error) {
	var attachment models.TransactionAttachment
	err := r.db.First(&attachment, id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// GetByTransactionID 根据交易ID获取附件列表
func (r *attachmentRepository) GetByTransactionID(transactionID uint) ([]models.TransactionAttachment, error) {
	var attachments []models.TransactionAttachment
	err := r.db.Where("transaction_id = ?", transactionID).Order("sort_order").Find(&attachments).Error
	return attachments, err
}

// Update 更新附件
func (r *attachmentRepository) Update(attachment *models.TransactionAttachment) error {
	return r.db.Save(attachment).Error
}

// Delete 删除附件
func (r *attachmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.TransactionAttachment{}, id).Error
}

