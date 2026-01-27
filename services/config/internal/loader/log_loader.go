package loader

import (
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
	"github.com/spf13/viper"
)

// LoadLogConfig 加载日志配置
func LoadLogConfig(v *viper.Viper) (*model.LogConfig, error) {
	var cfg model.LogConfig
	if err := v.Sub("log").Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// BindLogEnv 绑定日志相关环境变量
func BindLogEnv(v *viper.Viper) {
	v.BindEnv("log.level", "LOG_LEVEL")
	v.BindEnv("log.format", "LOG_FORMAT")
	v.BindEnv("log.output", "LOG_OUTPUT")
}
