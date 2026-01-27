package sync

import (
	"fmt"
	"os"

	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
)

// ExternalSyncHandler 外部同步处理器
// 处理当前服务请求其他服务的配置同步
func ExternalSyncHandler(cfg *model.Config, source model.ConfigSource, configPath string) error {
	// 根据配置来源执行相应的外部同步操作
	switch source {
	case model.ConfigSourceFile:
		// 配置文件变化，同步到环境变量
		return SyncConfigToEnv(cfg)
	case model.ConfigSourceEnv:
		// 环境变量变化，同步到配置文件
		return SyncConfigToFile(cfg, configPath)
	case model.ConfigSourceGRPC:
		// gRPC请求变化，同步到环境变量和配置文件
		if err := SyncConfigToEnv(cfg); err != nil {
			return err
		}
		return SyncConfigToFile(cfg, configPath)
	default:
		return fmt.Errorf("unknown config source: %v", source)
	}
}

// SyncConfigToEnv 将配置同步到环境变量
func SyncConfigToEnv(cfg *model.Config) error {
	// 设置环境变量
	envMap := map[string]string{
		"DATABASE_TYPE":  cfg.Database.Active,
		"STORAGE_TYPE":   cfg.Storage.Active,
		"ADMIN_PASSWORD": cfg.Server.AdminAccount.Password,
		"SERVER_PORT":    fmt.Sprintf("%d", cfg.Server.Port),
	}

	// 遍历设置环境变量
	for key, value := range envMap {
		if value != "" {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("failed to set env %s: %w", key, err)
			}
		}
	}

	return nil
}

// SyncConfigToFile 将配置同步到配置文件
func SyncConfigToFile(cfg *model.Config, configPath string) error {
	// 生成配置文件内容
	configContent := generateConfigContent(cfg)

	// 写入配置文件
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
