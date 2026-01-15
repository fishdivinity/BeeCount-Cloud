package admin

import (
	"fmt"

	"github.com/fishdivinity/BeeCount-Cloud/internal/auth"
	"github.com/fishdivinity/BeeCount-Cloud/internal/config"
	"github.com/fishdivinity/BeeCount-Cloud/internal/models"
	"gorm.io/gorm"
)

// Manager 管理员管理器
type Manager struct {
	db           *gorm.DB
	authService  auth.AuthService
	adminConfig  config.AdminConfig
}

// NewManager 创建管理员管理器
func NewManager(db *gorm.DB, authService auth.AuthService, adminConfig config.AdminConfig) *Manager {
	return &Manager{
		db:          db,
		authService: authService,
		adminConfig: adminConfig,
	}
}

// EnsureAdmin 确保管理员账户存在
// 如果不存在则创建默认管理员账户
func (m *Manager) EnsureAdmin() error {
	var count int64
	if err := m.db.Model(&models.User{}).Where("username = ?", m.adminConfig.Username).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check admin user: %w", err)
	}

	if count > 0 {
		return nil
	}

	hashedPassword, err := m.authService.HashPassword(m.adminConfig.Password)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	admin := &models.User{
		Username:     m.adminConfig.Username,
		Email:        fmt.Sprintf("%s@localhost", m.adminConfig.Username),
		PasswordHash: hashedPassword,
		DisplayName:  "Administrator",
	}

	if err := m.db.Create(admin).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	return nil
}

// IsAdmin 检查用户是否为管理员
func (m *Manager) IsAdmin(userID uint) bool {
	var user models.User
	if err := m.db.First(&user, userID).Error; err != nil {
		return false
	}

	return user.Username == m.adminConfig.Username
}