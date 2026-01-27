package loader

import (
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
	"github.com/spf13/viper"
)

// LoadDatabaseConfig 加载数据库配置
func LoadDatabaseConfig(v *viper.Viper) (*model.DatabaseConfig, error) {
	var cfg model.DatabaseConfig
	if err := v.Sub("database").Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// BindDatabaseEnv 绑定数据库相关环境变量
func BindDatabaseEnv(v *viper.Viper) {
	v.BindEnv("database.active", "DATABASE_TYPE")
	v.BindEnv("database.mysql.host", "DB_HOST")
	v.BindEnv("database.mysql.port", "DB_PORT")
	v.BindEnv("database.mysql.username", "DB_USER")
	v.BindEnv("database.mysql.password", "DB_PASSWORD")
	v.BindEnv("database.mysql.database", "DB_NAME")
}
