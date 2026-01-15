package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/internal/auth"
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"github.com/fishdivinity/BeeCount-Cloud/internal/repository"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
)

// syncService 同步服务实现
type syncService struct {
	repos       *repository.Repositories
	authService auth.AuthService
}

// NewSyncService 创建同步服务实例
func NewSyncService(repos *repository.Repositories, authService auth.AuthService) SyncService {
	return &syncService{
		repos:       repos,
		authService: authService,
	}
}

// SyncData 同步数据结构
// 用于客户端和服务器之间的数据同步
type SyncData struct {
	Version    int               `json:"version"`
	ExportedAt string            `json:"exported_at"`
	LedgerID   int               `json:"ledger_id"`
	LedgerName string            `json:"ledger_name"`
	Currency   string            `json:"currency"`
	Count      int               `json:"count"`
	Accounts   []SyncAccount     `json:"accounts"`
	Categories []SyncCategory    `json:"categories"`
	Tags       []SyncTag         `json:"tags"`
	Items      []SyncTransaction `json:"items"`
}

// SyncAccount 同步账户数据
type SyncAccount struct {
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Currency       string  `json:"currency"`
	InitialBalance float64 `json:"initial_balance"`
}

// SyncCategory 同步分类数据
type SyncCategory struct {
	Name            string `json:"name"`
	Kind            string `json:"kind"`
	Level           int    `json:"level"`
	SortOrder       int    `json:"sort_order"`
	Icon            string `json:"icon"`
	IconType        string `json:"icon_type"`
	CustomIconPath  string `json:"custom_icon_path"`
	CommunityIconID string `json:"community_icon_id"`
	ParentName      string `json:"parent_name"`
}

// SyncTag 同步标签数据
type SyncTag struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// SyncTransaction 同步交易数据
type SyncTransaction struct {
	Type            string           `json:"type"`
	Amount          float64          `json:"amount"`
	CategoryName    string           `json:"category_name"`
	CategoryKind    string           `json:"category_kind"`
	HappenedAt      string           `json:"happened_at"`
	Note            string           `json:"note"`
	AccountName     string           `json:"account_name"`
	FromAccountName string           `json:"from_account_name"`
	ToAccountName   string           `json:"to_account_name"`
	Tags            []string         `json:"tags"`
	Attachments     []SyncAttachment `json:"attachments"`
}

// SyncAttachment 同步附件数据
type SyncAttachment struct {
	FileName     string `json:"file_name"`
	OriginalName string `json:"original_name"`
	FileSize     *int   `json:"file_size"`
	Width        *int   `json:"width"`
	Height       *int   `json:"height"`
	SortOrder    int    `json:"sort_order"`
}

// UploadLedger 上传账本数据
// 将账本数据导出为JSON格式，用于客户端同步
func (s *syncService) UploadLedger(userID, ledgerID uint) (string, error) {
	ledger, err := s.repos.Ledger.GetByID(ledgerID)
	if err != nil {
		return "", utils.ErrNotFound
	}

	if ledger.UserID != userID {
		return "", utils.ErrForbidden
	}

	transactions, err := s.repos.Transaction.GetByLedgerID(ledgerID, 0, 0)
	if err != nil {
		return "", utils.WrapError(err, "failed to get transactions")
	}

	accounts, err := s.repos.Account.GetByUserID(userID)
	if err != nil {
		return "", utils.WrapError(err, "failed to get accounts")
	}

	categories, err := s.repos.Category.GetByUserID(userID)
	if err != nil {
		return "", utils.WrapError(err, "failed to get categories")
	}

	tags, err := s.repos.Tag.GetByUserID(userID)
	if err != nil {
		return "", utils.WrapError(err, "failed to get tags")
	}

	syncData := SyncData{
		Version:    5,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		LedgerID:   int(ledgerID),
		LedgerName: ledger.Name,
		Currency:   ledger.Currency,
		Count:      len(transactions),
		Accounts:   convertAccounts(accounts),
		Categories: convertCategories(categories),
		Tags:       convertTags(tags),
		Items:      convertTransactions(transactions),
	}

	jsonData, err := json.Marshal(syncData)
	if err != nil {
		return "", utils.WrapError(err, "failed to marshal sync data")
	}

	return string(jsonData), nil
}

// DownloadLedger 下载账本数据
// 返回账本数据的JSON格式，用于客户端同步
func (s *syncService) DownloadLedger(userID, ledgerID uint) (string, error) {
	ledger, err := s.repos.Ledger.GetByID(ledgerID)
	if err != nil {
		return "", utils.ErrNotFound
	}

	if ledger.UserID != userID {
		return "", utils.ErrForbidden
	}

	syncData, err := s.UploadLedger(userID, ledgerID)
	if err != nil {
		return "", err
	}

	return syncData, nil
}

// GetSyncStatus 获取同步状态
// 返回账本的同步状态信息，包括交易数量、最后更新时间等
func (s *syncService) GetSyncStatus(userID, ledgerID uint) (map[string]interface{}, error) {
	ledger, err := s.repos.Ledger.GetByID(ledgerID)
	if err != nil {
		return nil, utils.ErrNotFound
	}

	if ledger.UserID != userID {
		return nil, utils.ErrForbidden
	}

	transactions, err := s.repos.Transaction.GetByLedgerID(ledgerID, 0, 0)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get transactions")
	}

	status := map[string]interface{}{
		"ledger_id":    ledgerID,
		"ledger_name":  ledger.Name,
		"currency":     ledger.Currency,
		"count":        len(transactions),
		"last_updated": ledger.UpdatedAt,
		"fingerprint":  calculateFingerprint(transactions),
	}

	return status, nil
}

// convertAccounts 转换账户数据为同步格式
func convertAccounts(accounts []models.Account) []SyncAccount {
	result := make([]SyncAccount, len(accounts))
	for i, acc := range accounts {
		result[i] = SyncAccount{
			Name:           acc.Name,
			Type:           acc.Type,
			Currency:       acc.Currency,
			InitialBalance: acc.InitialBalance,
		}
	}
	return result
}

// convertCategories 转换分类数据为同步格式
func convertCategories(categories []models.Category) []SyncCategory {
	result := make([]SyncCategory, len(categories))
	for i, cat := range categories {
		parentName := ""
		if cat.Parent != nil {
			parentName = cat.Parent.Name
		}
		result[i] = SyncCategory{
			Name:            cat.Name,
			Kind:            cat.Kind,
			Level:           cat.Level,
			SortOrder:       cat.SortOrder,
			Icon:            cat.Icon,
			IconType:        cat.IconType,
			CustomIconPath:  cat.CustomIconPath,
			CommunityIconID: cat.CommunityIconID,
			ParentName:      parentName,
		}
	}
	return result
}

// convertTags 转换标签数据为同步格式
func convertTags(tags []models.Tag) []SyncTag {
	result := make([]SyncTag, len(tags))
	for i, tag := range tags {
		result[i] = SyncTag{
			Name:  tag.Name,
			Color: tag.Color,
		}
	}
	return result
}

// convertTransactions 转换交易数据为同步格式
func convertTransactions(transactions []models.Transaction) []SyncTransaction {
	result := make([]SyncTransaction, len(transactions))
	for i, tx := range transactions {
		tags := make([]string, len(tx.Tags))
		for j, tag := range tx.Tags {
			tags[j] = tag.Name
		}

		attachments := make([]SyncAttachment, len(tx.Attachments))
		for j, att := range tx.Attachments {
			attachments[j] = SyncAttachment{
				FileName:     att.FileName,
				OriginalName: att.OriginalName,
				FileSize:     att.FileSize,
				Width:        att.Width,
				Height:       att.Height,
				SortOrder:    att.SortOrder,
			}
		}

		accountName := ""
		fromAccountName := ""
		toAccountName := ""

		if tx.Type == "transfer" {
			if tx.Account != nil {
				fromAccountName = tx.Account.Name
			}
			if tx.ToAccount != nil {
				toAccountName = tx.ToAccount.Name
			}
		} else {
			if tx.Account != nil {
				accountName = tx.Account.Name
			}
		}

		categoryName := ""
		categoryKind := ""
		if tx.Category != nil {
			categoryName = tx.Category.Name
			categoryKind = tx.Category.Kind
		}

		result[i] = SyncTransaction{
			Type:            tx.Type,
			Amount:          tx.Amount,
			CategoryName:    categoryName,
			CategoryKind:    categoryKind,
			HappenedAt:      tx.HappenedAt.UTC().Format(time.RFC3339),
			Note:            tx.Note,
			AccountName:     accountName,
			FromAccountName: fromAccountName,
			ToAccountName:   toAccountName,
			Tags:            tags,
			Attachments:     attachments,
		}
	}
	return result
}

// calculateFingerprint 计算数据指纹
// 用于检测数据是否发生变化
func calculateFingerprint(transactions []models.Transaction) string {
	data := fmt.Sprintf("%d", len(transactions))
	return utils.CalculateSHA256Bytes([]byte(data))
}

