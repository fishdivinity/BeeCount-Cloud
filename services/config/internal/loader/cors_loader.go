package loader

import (
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
	"github.com/spf13/viper"
)

// LoadCORSConfig 加载CORS配置
func LoadCORSConfig(v *viper.Viper) (*model.CORSConfig, error) {
	var cfg model.CORSConfig
	if err := v.Sub("cors").Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// BindCORSEnv 绑定CORS相关环境变量
func BindCORSEnv(v *viper.Viper) {
	v.BindEnv("cors.allowed_origins", "CORS_ALLOWED_ORIGINS")
	v.BindEnv("cors.allowed_methods", "CORS_ALLOWED_METHODS")
	v.BindEnv("cors.allowed_headers", "CORS_ALLOWED_HEADERS")
}
