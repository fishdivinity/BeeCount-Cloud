package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/internal/auth"
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"github.com/fishdivinity/BeeCount-Cloud/internal/service"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
	"github.com/gin-gonic/gin"
)

// TransactionHandler 交易处理器
// 负责处理交易的创建、查询、更新和删除操作
type TransactionHandler struct {
	txService service.TransactionService
}

// NewTransactionHandler 创建交易处理器实例
func NewTransactionHandler(txService service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		txService: txService,
	}
}

// CreateTransaction 创建交易
// @Summary 创建交易
// @Description 为指定账本创建新的交易记录
// @Tags 交易
// @Accept json
// @Produce json
// @Security Bearer
// @Param ledger_id path int true "账本ID"
// @Param request body CreateTransactionRequest true "交易信息"
// @Success 201 {object} TransactionResponse "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers/{ledger_id}/transactions [post]
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	userID := auth.GetUserID(c)

	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError(err.Error()))
		return
	}

	tx := &models.Transaction{
		LedgerID:    req.LedgerID,
		Type:        req.Type,
		Amount:      req.Amount,
		CategoryID:  req.CategoryID,
		AccountID:   req.AccountID,
		ToAccountID: req.ToAccountID,
		Note:        req.Note,
	}

	parsedTime, err := time.Parse(time.RFC3339, req.HappenedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid happened_at format"))
		return
	}
	tx.HappenedAt = parsedTime

	if err := h.txService.CreateTransaction(userID, tx); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, tx)
}

// GetTransactions 获取账本的所有交易
// @Summary 获取交易列表
// @Description 返回指定账本中的交易列表，支持分页
// @Tags 交易
// @Accept json
// @Produce json
// @Security Bearer
// @Param ledger_id path int true "账本ID"
// @Param limit query int false "每页数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {array} TransactionResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers/{ledger_id}/transactions [get]
func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	ledgerID, err := strconv.ParseUint(c.Param("ledger_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid ledger id"))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	transactions, err := h.txService.GetLedgerTransactions(uint(ledgerID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// GetTransaction 获取单个交易
// @Summary 获取交易详情
// @Description 根据交易ID返回交易详细信息
// @Tags 交易
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "交易ID"
// @Success 200 {object} TransactionResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "交易不存在"
// @Router /api/v1/ledgers/{ledger_id}/transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid transaction id"))
		return
	}

	tx, err := h.txService.GetTransaction(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewNotFoundError("transaction not found"))
		return
	}

	c.JSON(http.StatusOK, tx)
}

// UpdateTransaction 更新交易
// @Summary 更新交易
// @Description 更新指定交易的信息，只有交易所有者可以更新
// @Tags 交易
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "交易ID"
// @Param request body UpdateTransactionRequest true "更新信息"
// @Success 200 {object} TransactionResponse "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 403 {object} ErrorResponse "无权限"
// @Failure 404 {object} ErrorResponse "交易不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers/{ledger_id}/transactions/{id} [put]
func (h *TransactionHandler) UpdateTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid transaction id"))
		return
	}

	tx, err := h.txService.GetTransaction(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewNotFoundError("transaction not found"))
		return
	}

	userID := auth.GetUserID(c)
	if tx.UserID != userID {
		c.JSON(http.StatusForbidden, utils.NewForbiddenError("access denied"))
		return
	}

	var req UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError(err.Error()))
		return
	}

	if req.Type != nil {
		tx.Type = *req.Type
	}
	if req.Amount != nil {
		tx.Amount = *req.Amount
	}
	if req.CategoryID != nil {
		tx.CategoryID = req.CategoryID
	}
	if req.AccountID != nil {
		tx.AccountID = req.AccountID
	}
	if req.ToAccountID != nil {
		tx.ToAccountID = req.ToAccountID
	}
	if req.HappenedAt != nil {
		parsedTime, err := time.Parse(time.RFC3339, *req.HappenedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid happened_at format"))
			return
		}
		tx.HappenedAt = parsedTime
	}
	if req.Note != nil {
		tx.Note = *req.Note
	}

	if err := h.txService.UpdateTransaction(tx); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, tx)
}

// DeleteTransaction 删除交易
// @Summary 删除交易
// @Description 删除指定交易，只有交易所有者可以删除
// @Tags 交易
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "交易ID"
// @Success 204 "删除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 403 {object} ErrorResponse "无权限"
// @Failure 404 {object} ErrorResponse "交易不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers/{ledger_id}/transactions/{id} [delete]
func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid transaction id"))
		return
	}

	tx, err := h.txService.GetTransaction(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewNotFoundError("transaction not found"))
		return
	}

	userID := auth.GetUserID(c)
	if tx.UserID != userID {
		c.JSON(http.StatusForbidden, utils.NewForbiddenError("access denied"))
		return
	}

	if err := h.txService.DeleteTransaction(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
