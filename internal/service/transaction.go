package service

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"github.com/fishdivinity/BeeCount-Cloud/internal/repository"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
)

// transactionService 交易服务实现
type transactionService struct {
	txRepo repository.TransactionRepository
}

// NewTransactionService 创建交易服务实例
func NewTransactionService(txRepo repository.TransactionRepository) TransactionService {
	return &transactionService{
		txRepo: txRepo,
	}
}

// CreateTransaction 创建交易
// 为指定用户创建新交易
func (s *transactionService) CreateTransaction(userID uint, tx *models.Transaction) error {
	tx.UserID = userID
	if err := s.txRepo.Create(tx); err != nil {
		return utils.WrapError(err, "failed to create transaction")
	}
	return nil
}

// GetTransaction 获取交易
// 根据ID返回交易信息
func (s *transactionService) GetTransaction(id uint) (*models.Transaction, error) {
	tx, err := s.txRepo.GetByID(id)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	return tx, nil
}

// GetLedgerTransactions 获取账本的所有交易
// 返回指定账本中的交易列表，支持分页
func (s *transactionService) GetLedgerTransactions(ledgerID uint, limit, offset int) ([]models.Transaction, error) {
	transactions, err := s.txRepo.GetByLedgerID(ledgerID, limit, offset)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get transactions")
	}
	return transactions, nil
}

// UpdateTransaction 更新交易
// 更新交易信息
func (s *transactionService) UpdateTransaction(tx *models.Transaction) error {
	if err := s.txRepo.Update(tx); err != nil {
		return utils.WrapError(err, "failed to update transaction")
	}
	return nil
}

// DeleteTransaction 删除交易
// 根据ID删除交易
func (s *transactionService) DeleteTransaction(id uint) error {
	if err := s.txRepo.Delete(id); err != nil {
		return utils.WrapError(err, "failed to delete transaction")
	}
	return nil
}

