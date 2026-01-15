package service

import (
	"github.com/fishdivinity/BeeCount-Cloud/internal/auth"
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"github.com/fishdivinity/BeeCount-Cloud/internal/repository"
	"github.com/fishdivinity/BeeCount-Cloud/pkg/utils"
)

// userService 用户服务实现
type userService struct {
	userRepo    repository.UserRepository
	authService auth.AuthService
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, authService auth.AuthService) UserService {
	return &userService{
		userRepo:    userRepo,
		authService: authService,
	}
}

// Register 用户注册
// 加密密码并创建新用户
func (s *userService) Register(user *models.User) error {
	hashedPassword, err := s.authService.HashPassword(user.PasswordHash)
	if err != nil {
		return utils.WrapError(err, "failed to hash password")
	}
	user.PasswordHash = hashedPassword

	if err := s.userRepo.Create(user); err != nil {
		return utils.WrapError(err, "failed to create user")
	}
	return nil
}

// Login 用户登录
// 验证邮箱和密码，返回用户信息和JWT token
func (s *userService) Login(email, password string) (*models.User, string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", utils.ErrInvalidCredentials
	}

	if err := s.authService.CheckPassword(password, user.PasswordHash); err != nil {
		return nil, "", utils.ErrInvalidCredentials
	}

	token, err := s.authService.GenerateToken(user)
	if err != nil {
		return nil, "", utils.WrapError(err, "failed to generate token")
	}

	user.PasswordHash = ""
	return user, token, nil
}

// GetUserByID 根据ID获取用户
// 返回用户信息，不包含密码哈希
func (s *userService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, utils.ErrUserNotFound
	}
	user.PasswordHash = ""
	return user, nil
}

// UpdateUser 更新用户信息
// 如果包含密码则重新加密
func (s *userService) UpdateUser(user *models.User) error {
	if user.PasswordHash != "" {
		hashedPassword, err := s.authService.HashPassword(user.PasswordHash)
		if err != nil {
			return utils.WrapError(err, "failed to hash password")
		}
		user.PasswordHash = hashedPassword
	}

	if err := s.userRepo.Update(user); err != nil {
		return utils.WrapError(err, "failed to update user")
	}
	return nil
}

