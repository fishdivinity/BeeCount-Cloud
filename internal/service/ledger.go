package service

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"github.com/fishdivinity/BeeCount-Cloud/internal/repository"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
)

// ledgerService 账本服务实现
type ledgerService struct {
	ledgerRepo repository.LedgerRepository
}

// NewLedgerService 创建账本服务实例
func NewLedgerService(ledgerRepo repository.LedgerRepository) LedgerService {
	return &ledgerService{
		ledgerRepo: ledgerRepo,
	}
}

// CreateLedger 创建账本
// 为指定用户创建新账本
func (s *ledgerService) CreateLedger(userID uint, ledger *models.Ledger) error {
	ledger.UserID = userID
	if err := s.ledgerRepo.Create(ledger); err != nil {
		return utils.WrapError(err, "failed to create ledger")
	}
	return nil
}

// GetLedger 获取账本
// 根据ID返回账本信息
func (s *ledgerService) GetLedger(id uint) (*models.Ledger, error) {
	ledger, err := s.ledgerRepo.GetByID(id)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	return ledger, nil
}

// GetUserLedgers 获取用户的所有账本
// 返回指定用户拥有的所有账本列表
func (s *ledgerService) GetUserLedgers(userID uint) ([]models.Ledger, error) {
	ledgers, err := s.ledgerRepo.GetByUserID(userID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get ledgers")
	}
	return ledgers, nil
}

// UpdateLedger 更新账本
// 更新账本信息
func (s *ledgerService) UpdateLedger(ledger *models.Ledger) error {
	if err := s.ledgerRepo.Update(ledger); err != nil {
		return utils.WrapError(err, "failed to update ledger")
	}
	return nil
}

// DeleteLedger 删除账本
// 根据ID删除账本
func (s *ledgerService) DeleteLedger(id uint) error {
	if err := s.ledgerRepo.Delete(id); err != nil {
		return utils.WrapError(err, "failed to delete ledger")
	}
	return nil
}

