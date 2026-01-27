package loader

import (
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
	"github.com/spf13/viper"
)

// LoadJWTConfig 加载JWT配置
func LoadJWTConfig(v *viper.Viper) (*model.JWTConfig, error) {
	var cfg model.JWTConfig
	if err := v.Sub("jwt").Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// BindJWTEnv 绑定JWT相关环境变量
func BindJWTEnv(v *viper.Viper) {
	v.BindEnv("jwt.secret", "JWT_SECRET")
	v.BindEnv("jwt.expire_hours", "JWT_EXPIRE_HOURS")
}
