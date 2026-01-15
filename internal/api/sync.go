package api

import (
	"net/http"
	"strconv"

	"github.com/fishdivinity/BeeCount-Cloud/internal/auth"
	"github.com/fishdivinity/BeeCount-Cloud/internal/service"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
	"github.com/gin-gonic/gin"
)

// SyncHandler 同步处理器
// 负责处理账本数据的上传、下载和同步状态查询
type SyncHandler struct {
	syncService service.SyncService
}

// NewSyncHandler 创建同步处理器实例
func NewSyncHandler(syncService service.SyncService) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
	}
}

// UploadLedger 上传账本数据
// @Summary 上传账本数据
// @Description 将账本数据导出为JSON格式，用于客户端同步
// @Tags 同步
// @Accept json
// @Produce json
// @Security Bearer
// @Param ledger_id path int true "账本ID"
// @Success 200 {string} string "上传成功，返回JSON数据"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers/{ledger_id}/sync/upload [post]
func (h *SyncHandler) UploadLedger(c *gin.Context) {
	userID := auth.GetUserID(c)

	ledgerID, err := strconv.ParseUint(c.Param("ledger_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid ledger id"))
		return
	}

	jsonData, err := h.syncService.UploadLedger(userID, uint(ledgerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, jsonData)
}

// DownloadLedger 下载账本数据
// @Summary 下载账本数据
// @Description 返回账本数据的JSON格式，用于客户端同步
// @Tags 同步
// @Accept json
// @Produce json
// @Security Bearer
// @Param ledger_id path int true "账本ID"
// @Success 200 {string} string "下载成功，返回JSON数据"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "账本不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers/{ledger_id}/sync/download [get]
func (h *SyncHandler) DownloadLedger(c *gin.Context) {
	userID := auth.GetUserID(c)

	ledgerID, err := strconv.ParseUint(c.Param("ledger_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid ledger id"))
		return
	}

	jsonData, err := h.syncService.DownloadLedger(userID, uint(ledgerID))
	if err != nil {
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, utils.NewNotFoundError("ledger not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, jsonData)
}

// GetSyncStatus 获取同步状态
// @Summary 获取同步状态
// @Description 返回账本的同步状态信息，包括交易数量、最后更新时间等
// @Tags 同步
// @Accept json
// @Produce json
// @Security Bearer
// @Param ledger_id path int true "账本ID"
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "账本不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers/{ledger_id}/sync/status [get]
func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
	userID := auth.GetUserID(c)

	ledgerID, err := strconv.ParseUint(c.Param("ledger_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid ledger id"))
		return
	}

	status, err := h.syncService.GetSyncStatus(userID, uint(ledgerID))
	if err != nil {
		if err == utils.ErrNotFound {
			c.JSON(http.StatusNotFound, utils.NewNotFoundError("ledger not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, status)
}
