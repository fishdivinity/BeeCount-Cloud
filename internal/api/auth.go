package api

import (
	"net/http"

	"github.com/fishdivinity/BeeCount-Cloud/internal/auth"
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"github.com/fishdivinity/BeeCount-Cloud/internal/service"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
// 负责处理用户注册、登录和获取当前用户信息
type AuthHandler struct {
	userService      service.UserService
	authService      auth.AuthService
	allowRegistration bool
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(userService service.UserService, authService auth.AuthService, allowRegistration bool) *AuthHandler {
	return &AuthHandler{
		userService:      userService,
		authService:      authService,
		allowRegistration: allowRegistration,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 201 {object} UserResponse "注册成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 403 {object} ErrorResponse "注册已禁用"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	if !h.allowRegistration {
		c.JSON(http.StatusForbidden, utils.NewForbiddenError("user registration is disabled"))
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError(err.Error()))
		return
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password,
	}

	if err := h.userService.Register(user); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewInternalServerError(err.Error()))
		return
	}

	user.PasswordHash = ""
	c.JSON(http.StatusCreated, user)
}

// Login 用户登录
// @Summary 用户登录
// @Description 验证用户凭证，返回JWT token和用户信息
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} AuthResponse "登录成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "认证失败"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewBadRequestError(err.Error()))
		return
	}

	user, token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewUnauthorizedError(err.Error()))
		return
	}

	response := AuthResponse{
		Token: token,
		User: UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Avatar:      user.Avatar,
			CreatedAt:   user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}

	c.JSON(http.StatusOK, response)
}

// Me 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 根据JWT token返回当前登录用户的详细信息
// @Tags 认证
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} UserResponse "获取成功"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Router /api/v1/users/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := auth.GetUserID(c)
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewNotFoundError("user not found"))
		return
	}

	c.JSON(http.StatusOK, user)
}
