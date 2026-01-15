package storage

import (
	"fmt"

	"github.com/fishdivinity/BeeCount-Cloud/internal/config"
)

// NewStorage 创建存储实例
// 根据配置类型返回对应的存储实现
func NewStorage(cfg *config.StorageConfig) (Storage, error) {
	switch cfg.Type {
	case "local":
		return NewLocalStorage(&cfg.Local), nil
	case "s3":
		return NewS3Storage(&cfg.S3)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
}

