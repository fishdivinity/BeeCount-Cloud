package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fishdivinity/BeeCount-Cloud/internal/auth"
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"github.com/fishdivinity/BeeCount-Cloud/internal/repository"
	"github.com/fishdivinity/BeeCount-Cloud/internal/storage"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AttachmentHandler 附件处理器
// 负责处理交易附件的上传、下载和删除操作
type AttachmentHandler struct {
	attachmentRepo repository.AttachmentRepository
	storage        storage.Storage
	maxFileSize    int64
	allowedTypes   []string
}

func NewAttachmentHandler(attachmentRepo repository.AttachmentRepository, storage storage.Storage, maxFileSize int64, allowedTypes []string) *AttachmentHandler {
	if maxFileSize <= 0 {
		maxFileSize = 10 * 1024 * 1024
	}
	if len(allowedTypes) == 0 {
		allowedTypes = []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	}
	return &AttachmentHandler{
		attachmentRepo: attachmentRepo,
		storage:        storage,
		maxFileSize:    maxFileSize,
		allowedTypes:   allowedTypes,
	}
}

// UploadAttachment 上传附件
// @Summary 上传附件
// @Description 上传交易附件，支持图片格式（JPEG、PNG、GIF、WebP），最大10MB
// @Tags 附件
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param ledger_id path int true "账本ID"
// @Param transaction_id path int true "交易ID"
// @Param file formData file true "附件文件"
// @Success 201 {object} map[string]interface{} "上传成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/ledgers/{ledger_id}/transactions/{transaction_id}/attachments [post]
func (h *AttachmentHandler) UploadAttachment(c *gin.Context) {
	userID := auth.GetUserID(c)
	transactionID, err := strconv.ParseUint(c.Param("transaction_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid transaction id"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("file is required"))
		return
	}

	contentType := storage.GetFileContentType(file)
	if !h.isAllowedFileType(contentType) {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid file type"))
		return
	}

	if file.Size > h.maxFileSize {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError(fmt.Sprintf("file size exceeds %dMB limit", h.maxFileSize/(1024*1024))))
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError("failed to open file"))
		return
	}
	defer src.Close()

	fileName := fmt.Sprintf("attachments/%d/%d_%s", userID, transactionID, file.Filename)

	metadata := &storage.Metadata{
		ContentType: contentType,
		Size:        file.Size,
	}

	if err := h.storage.Upload(c.Request.Context(), fileName, src, metadata); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError("failed to upload file"))
		return
	}

	url, err := h.storage.GetURL(c.Request.Context(), fileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError("failed to get file URL"))
		return
	}

	fileSize := int(file.Size)
	attachment := &models.TransactionAttachment{
		TransactionID: uint(transactionID),
		FileName:      fileName,
		OriginalName:  file.Filename,
		FileSize:      &fileSize,
	}

	if err := h.attachmentRepo.Create(attachment); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError("failed to create attachment record"))
		return
	}

	response := map[string]interface{}{
		"id":        attachment.ID,
		"file_name": attachment.FileName,
		"url":       url,
	}

	c.JSON(http.StatusCreated, response)
}

// GetAttachment 获取附件
// @Summary 获取附件
// @Description 根据附件ID下载文件并返回给客户端
// @Tags 附件
// @Accept json
// @Produce application/octet-stream
// @Security Bearer
// @Param id path int true "附件ID"
// @Success 200 {file} binary "文件内容"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "附件不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/attachments/{id} [get]
func (h *AttachmentHandler) GetAttachment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid attachment id"))
		return
	}

	attachment, err := h.attachmentRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewNotFoundError("attachment not found"))
		return
	}

	fileSize, err := h.storage.GetSize(c.Request.Context(), attachment.FileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError("failed to get file size"))
		return
	}

	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		var from, to int64
		_, err := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &from, &to)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid range header"))
			return
		}

		if from < 0 || from >= fileSize {
			c.JSON(http.StatusRequestedRangeNotSatisfiable, utils.NewBadRequestError("invalid range"))
			return
		}

		if to == 0 || to >= fileSize {
			to = fileSize - 1
		}

		length := to - from + 1
		reader, err := h.storage.DownloadRange(c.Request.Context(), attachment.FileName, from, length)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.NewInternalServerError("failed to download file range"))
			return
		}
		defer reader.Close()

		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", from, to, fileSize))
		c.Header("Content-Length", strconv.FormatInt(length, 10))
		c.Header("Accept-Ranges", "bytes")
		c.DataFromReader(http.StatusPartialContent, length, "application/octet-stream", reader, nil)
		return
	}

	reader, err := h.storage.Download(c.Request.Context(), attachment.FileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError("failed to download file"))
		return
	}
	defer reader.Close()

	c.Header("Content-Length", strconv.FormatInt(fileSize, 10))
	c.Header("Accept-Ranges", "bytes")
	c.DataFromReader(http.StatusOK, fileSize, "application/octet-stream", reader, nil)
}

// DeleteAttachment 删除附件
// @Summary 删除附件
// @Description 从存储后端删除文件，并删除数据库中的附件记录
// @Tags 附件
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "附件ID"
// @Success 204 "删除成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "附件不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/attachments/{id} [delete]
func (h *AttachmentHandler) DeleteAttachment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError("invalid attachment id"))
		return
	}

	attachment, err := h.attachmentRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewNotFoundError("attachment not found"))
		return
	}

	if err := h.storage.Delete(c.Request.Context(), attachment.FileName); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError("failed to delete file"))
		return
	}

	if err := h.attachmentRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError("failed to delete attachment record"))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *AttachmentHandler) isAllowedFileType(contentType string) bool {
	for _, allowedType := range h.allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}
