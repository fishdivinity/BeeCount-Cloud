package loader

import (
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
	"github.com/spf13/viper"
)

// LoadServerConfig 加载服务器配置
func LoadServerConfig(v *viper.Viper) (*model.ServerConfig, error) {
	var cfg model.ServerConfig
	if err := v.Sub("server").Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// BindServerEnv 绑定服务器相关环境变量
func BindServerEnv(v *viper.Viper) {
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.mode", "SERVER_MODE")
	v.BindEnv("server.admin_account.password", "ADMIN_PASSWORD")
}
