package loader

import (
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
	"github.com/spf13/viper"
)

// LoadStorageConfig 加载存储配置
func LoadStorageConfig(v *viper.Viper) (*model.StorageConfig, error) {
	var cfg model.StorageConfig
	if err := v.Sub("storage").Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// BindStorageEnv 绑定存储相关环境变量
func BindStorageEnv(v *viper.Viper) {
	v.BindEnv("storage.active", "STORAGE_TYPE")
	v.BindEnv("storage.s3.access_key_id", "S3_ACCESS_KEY")
	v.BindEnv("storage.s3.secret_access_key", "S3_SECRET_KEY")
	v.BindEnv("storage.s3.bucket", "S3_BUCKET")
	v.BindEnv("storage.s3.region", "S3_REGION")
	v.BindEnv("storage.s3.endpoint", "S3_ENDPOINT")
}
