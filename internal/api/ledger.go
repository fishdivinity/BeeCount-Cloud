package api

import (
	"net/http"
	"strconv"

	"github.com/fishdivinity/BeeCount-Cloud/internal/auth"
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"github.com/fishdivinity/BeeCount-Cloud/internal/service"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
	"github.com/gin-gonic/gin"
)

// LedgerHandler 账本处理器
// 负责处理账本的创建、查询、更新和删除操作
type LedgerHandler struct {
	ledgerService service.LedgerService
}

// NewLedgerHandler 创建账本处理器实例
func NewLedgerHandler(ledgerService service.LedgerService) *LedgerHandler {
	return &LedgerHandler{
		ledgerService: ledgerService,
	}
}

// CreateLedger 创建账本
// @Summary 创建账本
// @Description 为当前用户创建新的账本
// @Tags 账本
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateLedgerRequest true "账本信息"
// @Success 201 {object} LedgerResponse "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers [post]
func (h *LedgerHandler) CreateLedger(c *gin.Context) {
	userID := auth.GetUserID(c)

	var req CreateLedgerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError(err.Error()))
		return
	}

	ledger := &models.Ledger{
		Name:     req.Name,
		Currency: req.Currency,
		Type:     req.Type,
	}

	if ledger.Currency == "" {
		ledger.Currency = "CNY"
	}
	if ledger.Type == "" {
		ledger.Type = "personal"
	}

	if err := h.ledgerService.CreateLedger(userID, ledger); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, ledger)
}

// GetLedgers 获取当前用户的所有账本
// @Summary 获取账本列表
// @Description 返回当前用户拥有的所有账本列表
// @Tags 账本
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} LedgerResponse "获取成功"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers [get]
func (h *LedgerHandler) GetLedgers(c *gin.Context) {
	userID := auth.GetUserID(c)

	ledgers, err := h.ledgerService.GetUserLedgers(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, ledgers)
}

// GetLedger 获取单个账本
// @Summary 获取账本详情
// @Description 根据账本ID返回账本详细信息
// @Tags 账本
// @Accept json
// @Produce json
// @Security Bearer
// @Param ledger_id path int true "账本ID"
// @Success 200 {object} LedgerResponse "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "账本不存在"
// @Router /api/v1/ledgers/{ledger_id} [get]
func (h *LedgerHandler) GetLedger(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("ledger_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid ledger id"))
		return
	}

	ledger, err := h.ledgerService.GetLedger(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewNotFoundError("ledger not found"))
		return
	}

	c.JSON(http.StatusOK, ledger)
}

// UpdateLedger 更新账本
// @Summary 更新账本
// @Description 更新指定账本的信息，只有账本所有者可以更新
// @Tags 账本
// @Accept json
// @Produce json
// @Security Bearer
// @Param ledger_id path int true "账本ID"
// @Param request body UpdateLedgerRequest true "更新信息"
// @Success 200 {object} LedgerResponse "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 403 {object} ErrorResponse "无权限"
// @Failure 404 {object} ErrorResponse "账本不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers/{ledger_id} [put]
func (h *LedgerHandler) UpdateLedger(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("ledger_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid ledger id"))
		return
	}

	ledger, err := h.ledgerService.GetLedger(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewNotFoundError("ledger not found"))
		return
	}

	userID := auth.GetUserID(c)
	if ledger.UserID != userID {
		c.JSON(http.StatusForbidden, utils.NewForbiddenError("access denied"))
		return
	}

	var req UpdateLedgerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError(err.Error()))
		return
	}

	if req.Name != nil {
		ledger.Name = *req.Name
	}
	if req.Currency != nil {
		ledger.Currency = *req.Currency
	}
	if req.Type != nil {
		ledger.Type = *req.Type
	}

	if err := h.ledgerService.UpdateLedger(ledger); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, ledger)
}

// DeleteLedger 删除账本
// @Summary 删除账本
// @Description 删除指定账本，只有账本所有者可以删除
// @Tags 账本
// @Accept json
// @Produce json
// @Security Bearer
// @Param ledger_id path int true "账本ID"
// @Success 204 "删除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 403 {object} ErrorResponse "无权限"
// @Failure 404 {object} ErrorResponse "账本不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers/{ledger_id} [delete]
func (h *LedgerHandler) DeleteLedger(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("ledger_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid ledger id"))
		return
	}

	ledger, err := h.ledgerService.GetLedger(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewNotFoundError("ledger not found"))
		return
	}

	userID := auth.GetUserID(c)
	if ledger.UserID != userID {
		c.JSON(http.StatusForbidden, utils.NewForbiddenError("access denied"))
		return
	}

	if err := h.ledgerService.DeleteLedger(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
