package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/config"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConfigService 配置服务实现
type ConfigService struct {
	config.UnimplementedConfigServiceServer
	common.UnimplementedHealthCheckServiceServer
	configs    map[string]*config.ConfigItem
	viper      *viper.Viper
	watcher    *fsnotify.Watcher
	configPath string
	version    string
}

// NewConfigService 创建配置服务实例
func NewConfigService() *ConfigService {
	return &ConfigService{
		configs: make(map[string]*config.ConfigItem),
		viper:   viper.New(),
		version: "v1.0.0",
	}
}

// LoadConfig 加载配置文件
func (s *ConfigService) LoadConfig(configPath string) error {
	s.configPath = configPath

	// 设置配置文件路径
	s.viper.AddConfigPath(configPath)
	s.viper.SetConfigType("yaml")

	// 读取所有配置文件
	s.viper.AutomaticEnv()
	s.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 读取目录下所有yaml文件
	files, err := os.ReadDir(configPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) == ".yaml" {
			// 读取配置文件
			configName := strings.TrimSuffix(file.Name(), ".yaml")
			s.viper.SetConfigName(configName)

			if err := s.viper.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return err
				}
			}
		}
	}

	// 将配置加载到内存
	s.loadConfigsToMemory()

	return nil
}

// loadConfigsToMemory 将配置加载到内存
func (s *ConfigService) loadConfigsToMemory() {
	// 清空现有配置
	for k := range s.configs {
		delete(s.configs, k)
	}

	// 遍历所有配置键
	for _, key := range s.viper.AllKeys() {
		value := s.viper.Get(key)
		valueStr := fmt.Sprintf("%v", value)
		valueType := fmt.Sprintf("%T", value)

		s.configs[key] = &config.ConfigItem{
			Key:   key,
			Value: valueStr,
			Type:  valueType,
		}
	}
}

// StartWatcher 启动配置文件监听
func (s *ConfigService) StartWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	s.watcher = watcher

	// 监听配置目录
	if err := watcher.Add(s.configPath); err != nil {
		log.Fatalf("Failed to watch config directory: %v", err)
	}

	// 启动监听协程
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					// 重新加载配置
					if err := s.LoadConfig(s.configPath); err != nil {
						log.Printf("Failed to reload config: %v", err)
					} else {
						log.Printf("Config reloaded successfully")
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()
}

// GetConfig 获取配置
func (s *ConfigService) GetConfig(ctx context.Context, req *config.GetConfigRequest) (*config.GetConfigResponse, error) {
	configs := make(map[string]*config.ConfigItem)

	if len(req.Keys) == 0 {
		// 返回所有配置
		for k, v := range s.configs {
			configs[k] = v
		}
	} else {
		// 返回指定配置
		for _, key := range req.Keys {
			if v, ok := s.configs[key]; ok {
				configs[key] = v
			}
		}
	}

	return &config.GetConfigResponse{
		Configs: configs,
		Version: s.version,
	}, nil
}

// WatchConfig 监听配置变化
func (s *ConfigService) WatchConfig(req *config.WatchConfigRequest, stream config.ConfigService_WatchConfigServer) error {
	// 实现配置监听逻辑
	return status.Errorf(codes.Unimplemented, "method WatchConfig not implemented")
}

// ReloadConfig 重新加载配置
func (s *ConfigService) ReloadConfig(ctx context.Context, req *config.ReloadConfigRequest) (*common.Response, error) {
	if err := s.LoadConfig(s.configPath); err != nil {
		return &common.Response{
			Success: false,
			Message: fmt.Sprintf("Failed to reload config: %v", err),
			Code:    500,
		}, nil
	}

	return &common.Response{
		Success: true,
		Message: "Config reloaded successfully",
		Code:    200,
	}, nil
}

// Check 健康检查
func (s *ConfigService) Check(ctx context.Context, req *common.HealthCheckRequest) (*common.HealthCheckResponse, error) {
	return &common.HealthCheckResponse{
		Status: common.HealthCheckResponse_SERVING,
	}, nil
}

// Watch 健康检查监听
func (s *ConfigService) Watch(req *common.HealthCheckRequest, stream common.HealthCheckService_WatchServer) error {
	// 实现健康检查监听逻辑
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}
