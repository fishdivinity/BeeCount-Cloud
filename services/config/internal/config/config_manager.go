package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	stdsync "sync"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/config"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/generator"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/loader"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/sync"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/watcher"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	config.UnimplementedConfigServiceServer
	common.UnimplementedHealthCheckServiceServer

	configPath  string
	currentCfg  *model.Config
	isActive    bool // 服务是否已激活
	subscribers map[int64]config.ConfigService_WatchConfigServer
	nextSubID   int64
	fileWatcher *watcher.FileWatcher
	envWatcher  *watcher.EnvWatcher
	mu          stdsync.RWMutex // 互斥锁，保护共享资源
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath:  configPath,
		subscribers: make(map[int64]config.ConfigService_WatchConfigServer),
		nextSubID:   1,
	}
}

// Init 初始化配置管理器
func (cm *ConfigManager) Init() error {
	// 确保配置文件所在的目录存在
	configDir := filepath.Dir(cm.configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		log.Printf("Config directory %s not found, creating...", configDir)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// 配置文件不存在，生成默认配置
		log.Println("Config file not found, generating default config...")
		if err := generator.GenerateDefaultConfig(cm.configPath); err != nil {
			return fmt.Errorf("failed to generate default config: %w", err)
		}
	}

	// 加载配置
	cfg, _, err := loader.LoadConfig(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 检查配置完整性
	cfg = sync.CheckConfigIntegrity(cfg)

	// 同步配置到环境变量
	if err := sync.SyncConfig(cfg, model.ConfigSourceFile, cm.configPath); err != nil {
		return fmt.Errorf("failed to sync config to env: %w", err)
	}

	cm.mu.Lock()
	cm.currentCfg = cfg
	cm.mu.Unlock()

	// 启动配置监听器
	if err := cm.startWatchers(); err != nil {
		return fmt.Errorf("failed to start watchers: %w", err)
	}

	log.Println("ConfigManager initialized successfully")
	return nil
}

// startWatchers 启动配置监听器
func (cm *ConfigManager) startWatchers() error {
	// 启动文件监听器
	fileWatcher, err := watcher.NewFileWatcher(cm.configPath, func(cfg *model.Config) {
		cm.handleConfigChange(cfg, model.ConfigSourceFile)
	})
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	cm.fileWatcher = fileWatcher
	fileWatcher.Start()

	// 启动环境变量监听器
	envWatcher := watcher.NewEnvWatcher(5*time.Second, func(envMap map[string]string) {
		cm.handleEnvChange(envMap)
	})
	cm.envWatcher = envWatcher
	envWatcher.Start()

	return nil
}

// handleConfigChange 处理配置变化
func (cm *ConfigManager) handleConfigChange(cfg *model.Config, source model.ConfigSource) {
	// 检查配置完整性
	cfg = sync.CheckConfigIntegrity(cfg)

	// 同步配置
	if err := sync.SyncConfig(cfg, source, cm.configPath); err != nil {
		log.Printf("Failed to sync config: %v", err)
		return
	}

	// 更新当前配置
	cm.mu.Lock()
	oldCfg := cm.currentCfg
	cm.currentCfg = cfg
	cm.mu.Unlock()

	// 通知所有订阅者
	cm.notifySubscribers(oldCfg)
}

// handleEnvChange 处理环境变量变化
func (cm *ConfigManager) handleEnvChange(_ map[string]string) {
	// 重新加载配置，环境变量会自动覆盖配置文件
	cfg, _, err := loader.LoadConfig(cm.configPath)
	if err != nil {
		log.Printf("Failed to reload config from env change: %v", err)
		return
	}

	cm.handleConfigChange(cfg, model.ConfigSourceEnv)
}

// notifySubscribers 通知所有订阅者配置变更
func (cm *ConfigManager) notifySubscribers(_ *model.Config) {
	// 创建配置变更事件
	event := &config.ConfigChangeEvent{
		Version:   "v1.0.0",
		Timestamp: time.Now().Format(time.RFC3339),
		Key:       "config.updated",
	}

	// 遍历所有订阅者
	cm.mu.RLock()
	subscribersCopy := make(map[int64]config.ConfigService_WatchConfigServer, len(cm.subscribers))
	for k, v := range cm.subscribers {
		subscribersCopy[k] = v
	}
	cm.mu.RUnlock()

	for subID, stream := range subscribersCopy {
		if err := stream.Send(event); err != nil {
			log.Printf("Failed to send config change event to subscriber %d: %v", subID, err)
			// 移除无效的订阅者
			cm.mu.Lock()
			delete(cm.subscribers, subID)
			cm.mu.Unlock()
		}
	}
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig(ctx context.Context, req *config.GetConfigRequest) (*config.GetConfigResponse, error) {
	cm.mu.RLock()
	cfg := cm.currentCfg
	cm.mu.RUnlock()

	configs := make(map[string]*config.ConfigItem)

	// 将模型转换为gRPC响应格式
	cm.convertServerConfig(cfg.Server, configs)
	cm.convertDatabaseConfig(cfg.Database, configs)
	cm.convertStorageConfig(cfg.Storage, configs)
	cm.convertJWTConfig(cfg.JWT, configs)
	cm.convertLogConfig(cfg.Log, configs)
	cm.convertCORSConfig(cfg.CORS, configs)
	cm.convertCacheConfig(cfg.Cache, configs)

	return &config.GetConfigResponse{
		Configs: configs,
		Version: "v1.0.0",
	}, nil
}

// convertServerConfig 转换服务器配置
func (cm *ConfigManager) convertServerConfig(server model.ServerConfig, configs map[string]*config.ConfigItem) {
	configs["server.port"] = &config.ConfigItem{
		Key:   "server.port",
		Value: fmt.Sprintf("%d", server.Port),
		Type:  "int",
	}
	configs["server.mode"] = &config.ConfigItem{
		Key:   "server.mode",
		Value: server.Mode,
		Type:  "string",
	}
	configs["server.read_timeout"] = &config.ConfigItem{
		Key:   "server.read_timeout",
		Value: fmt.Sprintf("%v", server.ReadTimeout),
		Type:  "duration",
	}
	configs["server.write_timeout"] = &config.ConfigItem{
		Key:   "server.write_timeout",
		Value: fmt.Sprintf("%v", server.WriteTimeout),
		Type:  "duration",
	}
	configs["server.docs.enabled"] = &config.ConfigItem{
		Key:   "server.docs.enabled",
		Value: fmt.Sprintf("%t", server.Docs.Enabled),
		Type:  "bool",
	}
	configs["server.admin_account.username"] = &config.ConfigItem{
		Key:   "server.admin_account.username",
		Value: server.AdminAccount.Username,
		Type:  "string",
	}
	configs["server.admin_account.password"] = &config.ConfigItem{
		Key:   "server.admin_account.password",
		Value: server.AdminAccount.Password,
		Type:  "string",
	}
}

// convertDatabaseConfig 转换数据库配置
func (cm *ConfigManager) convertDatabaseConfig(db model.DatabaseConfig, configs map[string]*config.ConfigItem) {
	configs["database.active"] = &config.ConfigItem{
		Key:   "database.active",
		Value: db.Active,
		Type:  "string",
	}
	// SQLite配置
	configs["database.sqlite.path"] = &config.ConfigItem{
		Key:   "database.sqlite.path",
		Value: db.SQLite.Path,
		Type:  "string",
	}
	// MySQL配置
	configs["database.mysql.host"] = &config.ConfigItem{
		Key:   "database.mysql.host",
		Value: db.MySQL.Host,
		Type:  "string",
	}
	configs["database.mysql.port"] = &config.ConfigItem{
		Key:   "database.mysql.port",
		Value: fmt.Sprintf("%d", db.MySQL.Port),
		Type:  "int",
	}
	configs["database.mysql.username"] = &config.ConfigItem{
		Key:   "database.mysql.username",
		Value: db.MySQL.Username,
		Type:  "string",
	}
	configs["database.mysql.password"] = &config.ConfigItem{
		Key:   "database.mysql.password",
		Value: db.MySQL.Password,
		Type:  "string",
	}
	configs["database.mysql.database"] = &config.ConfigItem{
		Key:   "database.mysql.database",
		Value: db.MySQL.Database,
		Type:  "string",
	}
	configs["database.mysql.charset"] = &config.ConfigItem{
		Key:   "database.mysql.charset",
		Value: db.MySQL.Charset,
		Type:  "string",
	}
	configs["database.mysql.parse_time"] = &config.ConfigItem{
		Key:   "database.mysql.parse_time",
		Value: fmt.Sprintf("%t", db.MySQL.ParseTime),
		Type:  "bool",
	}
	configs["database.mysql.loc"] = &config.ConfigItem{
		Key:   "database.mysql.loc",
		Value: db.MySQL.Loc,
		Type:  "string",
	}
	// Postgres配置
	configs["database.postgres.host"] = &config.ConfigItem{
		Key:   "database.postgres.host",
		Value: db.Postgres.Host,
		Type:  "string",
	}
	configs["database.postgres.port"] = &config.ConfigItem{
		Key:   "database.postgres.port",
		Value: fmt.Sprintf("%d", db.Postgres.Port),
		Type:  "int",
	}
	configs["database.postgres.username"] = &config.ConfigItem{
		Key:   "database.postgres.username",
		Value: db.Postgres.Username,
		Type:  "string",
	}
	configs["database.postgres.password"] = &config.ConfigItem{
		Key:   "database.postgres.password",
		Value: db.Postgres.Password,
		Type:  "string",
	}
	configs["database.postgres.database"] = &config.ConfigItem{
		Key:   "database.postgres.database",
		Value: db.Postgres.Database,
		Type:  "string",
	}
	configs["database.postgres.sslmode"] = &config.ConfigItem{
		Key:   "database.postgres.sslmode",
		Value: db.Postgres.SSLMode,
		Type:  "string",
	}
	configs["database.postgres.timezone"] = &config.ConfigItem{
		Key:   "database.postgres.timezone",
		Value: db.Postgres.Timezone,
		Type:  "string",
	}
	// 连接池配置
	configs["database.pool.max_idle_conns"] = &config.ConfigItem{
		Key:   "database.pool.max_idle_conns",
		Value: fmt.Sprintf("%d", db.Pool.MaxIdleConns),
		Type:  "int",
	}
	configs["database.pool.max_open_conns"] = &config.ConfigItem{
		Key:   "database.pool.max_open_conns",
		Value: fmt.Sprintf("%d", db.Pool.MaxOpenConns),
		Type:  "int",
	}
	configs["database.pool.conn_max_lifetime"] = &config.ConfigItem{
		Key:   "database.pool.conn_max_lifetime",
		Value: fmt.Sprintf("%v", db.Pool.ConnMaxLifetime),
		Type:  "duration",
	}
	configs["database.pool.conn_max_idle_time"] = &config.ConfigItem{
		Key:   "database.pool.conn_max_idle_time",
		Value: fmt.Sprintf("%v", db.Pool.ConnMaxIdleTime),
		Type:  "duration",
	}
}

// convertStorageConfig 转换存储配置
func (cm *ConfigManager) convertStorageConfig(storage model.StorageConfig, configs map[string]*config.ConfigItem) {
	configs["storage.active"] = &config.ConfigItem{
		Key:   "storage.active",
		Value: storage.Active,
		Type:  "string",
	}
	configs["storage.max_file_size"] = &config.ConfigItem{
		Key:   "storage.max_file_size",
		Value: fmt.Sprintf("%d", storage.MaxFileSize),
		Type:  "int",
	}
	// 本地存储配置
	configs["storage.local.path"] = &config.ConfigItem{
		Key:   "storage.local.path",
		Value: storage.Local.Path,
		Type:  "string",
	}
	configs["storage.local.url_prefix"] = &config.ConfigItem{
		Key:   "storage.local.url_prefix",
		Value: storage.Local.URLPrefix,
		Type:  "string",
	}
	// S3存储配置
	configs["storage.s3.region"] = &config.ConfigItem{
		Key:   "storage.s3.region",
		Value: storage.S3.Region,
		Type:  "string",
	}
	configs["storage.s3.bucket"] = &config.ConfigItem{
		Key:   "storage.s3.bucket",
		Value: storage.S3.Bucket,
		Type:  "string",
	}
	configs["storage.s3.access_key_id"] = &config.ConfigItem{
		Key:   "storage.s3.access_key_id",
		Value: storage.S3.AccessKeyID,
		Type:  "string",
	}
	configs["storage.s3.secret_access_key"] = &config.ConfigItem{
		Key:   "storage.s3.secret_access_key",
		Value: storage.S3.SecretAccessKey,
		Type:  "string",
	}
	configs["storage.s3.endpoint"] = &config.ConfigItem{
		Key:   "storage.s3.endpoint",
		Value: storage.S3.Endpoint,
		Type:  "string",
	}
}

// convertJWTConfig 转换JWT配置
func (cm *ConfigManager) convertJWTConfig(jwt model.JWTConfig, configs map[string]*config.ConfigItem) {
	configs["jwt.secret"] = &config.ConfigItem{
		Key:   "jwt.secret",
		Value: jwt.Secret,
		Type:  "string",
	}
	configs["jwt.expire_hours"] = &config.ConfigItem{
		Key:   "jwt.expire_hours",
		Value: fmt.Sprintf("%d", jwt.ExpireHours),
		Type:  "int",
	}
	configs["jwt.rotation_interval_days"] = &config.ConfigItem{
		Key:   "jwt.rotation_interval_days",
		Value: fmt.Sprintf("%d", jwt.RotationIntervalDays),
		Type:  "int",
	}
	configs["jwt.last_rotation_date"] = &config.ConfigItem{
		Key:   "jwt.last_rotation_date",
		Value: jwt.LastRotationDate,
		Type:  "string",
	}
	configs["jwt.previous_secret"] = &config.ConfigItem{
		Key:   "jwt.previous_secret",
		Value: jwt.PreviousSecret,
		Type:  "string",
	}
}

// convertLogConfig 转换日志配置
func (cm *ConfigManager) convertLogConfig(log model.LogConfig, configs map[string]*config.ConfigItem) {
	configs["log.level"] = &config.ConfigItem{
		Key:   "log.level",
		Value: log.Level,
		Type:  "string",
	}
	configs["log.format"] = &config.ConfigItem{
		Key:   "log.format",
		Value: log.Format,
		Type:  "string",
	}
	configs["log.output"] = &config.ConfigItem{
		Key:   "log.output",
		Value: log.Output,
		Type:  "string",
	}
	// 文件日志配置
	configs["log.file.path"] = &config.ConfigItem{
		Key:   "log.file.path",
		Value: log.File.Path,
		Type:  "string",
	}
	configs["log.file.max_size"] = &config.ConfigItem{
		Key:   "log.file.max_size",
		Value: fmt.Sprintf("%d", log.File.MaxSize),
		Type:  "int",
	}
	configs["log.file.max_backups"] = &config.ConfigItem{
		Key:   "log.file.max_backups",
		Value: fmt.Sprintf("%d", log.File.MaxBackups),
		Type:  "int",
	}
	configs["log.file.max_age"] = &config.ConfigItem{
		Key:   "log.file.max_age",
		Value: fmt.Sprintf("%d", log.File.MaxAge),
		Type:  "int",
	}
	configs["log.file.compress"] = &config.ConfigItem{
		Key:   "log.file.compress",
		Value: fmt.Sprintf("%t", log.File.Compress),
		Type:  "bool",
	}
	configs["log.file.max_total_size_gb"] = &config.ConfigItem{
		Key:   "log.file.max_total_size_gb",
		Value: fmt.Sprintf("%d", log.File.MaxTotalSizeGB),
		Type:  "int",
	}
}

// convertCORSConfig 转换CORS配置
func (cm *ConfigManager) convertCORSConfig(cors model.CORSConfig, configs map[string]*config.ConfigItem) {
	configs["cors.allow_credentials"] = &config.ConfigItem{
		Key:   "cors.allow_credentials",
		Value: fmt.Sprintf("%t", cors.AllowCredentials),
		Type:  "bool",
	}
	configs["cors.max_age"] = &config.ConfigItem{
		Key:   "cors.max_age",
		Value: cors.MaxAge,
		Type:  "string",
	}
}

// convertCacheConfig 转换缓存配置
func (cm *ConfigManager) convertCacheConfig(cache model.CacheConfig, configs map[string]*config.ConfigItem) {
	configs["cache.active"] = &config.ConfigItem{
		Key:   "cache.active",
		Value: cache.Active,
		Type:  "string",
	}
	// 内存缓存配置
	configs["cache.memory.max_size"] = &config.ConfigItem{
		Key:   "cache.memory.max_size",
		Value: fmt.Sprintf("%d", cache.Memory.MaxSize),
		Type:  "int",
	}
	// Redis缓存配置
	configs["cache.redis.host"] = &config.ConfigItem{
		Key:   "cache.redis.host",
		Value: cache.Redis.Host,
		Type:  "string",
	}
	configs["cache.redis.port"] = &config.ConfigItem{
		Key:   "cache.redis.port",
		Value: fmt.Sprintf("%d", cache.Redis.Port),
		Type:  "int",
	}
	configs["cache.redis.password"] = &config.ConfigItem{
		Key:   "cache.redis.password",
		Value: cache.Redis.Password,
		Type:  "string",
	}
	configs["cache.redis.db"] = &config.ConfigItem{
		Key:   "cache.redis.db",
		Value: fmt.Sprintf("%d", cache.Redis.DB),
		Type:  "int",
	}
}

// WatchConfig 监听配置变化
func (cm *ConfigManager) WatchConfig(req *config.WatchConfigRequest, stream config.ConfigService_WatchConfigServer) error {
	// 生成订阅ID
	subID := cm.nextSubID
	cm.nextSubID++

	// 添加订阅者
	cm.mu.Lock()
	cm.subscribers[subID] = stream
	cm.mu.Unlock()
	log.Printf("New subscriber: %d", subID)

	// 保持连接，直到客户端断开
	<-stream.Context().Done()

	// 移除订阅者
	cm.mu.Lock()
	delete(cm.subscribers, subID)
	cm.mu.Unlock()
	log.Printf("Subscriber disconnected: %d", subID)

	return nil
}

// ReloadConfig 重新加载配置
func (cm *ConfigManager) ReloadConfig(ctx context.Context, req *config.ReloadConfigRequest) (*common.Response, error) {
	log.Println("Reloading config...")

	// 重新加载配置
	cfg, _, err := loader.LoadConfig(cm.configPath)
	if err != nil {
		return &common.Response{
			Success: false,
			Message: fmt.Sprintf("Failed to reload config: %v", err),
			Code:    500,
		}, nil
	}

	cm.handleConfigChange(cfg, model.ConfigSourceGRPC)

	return &common.Response{
		Success: true,
		Message: "Config reloaded successfully",
		Code:    200,
	}, nil
}

// StartService 启动服务
func (cm *ConfigManager) StartService(ctx context.Context, req *config.StartServiceRequest) (*config.StartServiceResponse, error) {
	log.Printf("Received StartService request: %v", req)

	// 检查服务是否已经激活
	cm.mu.RLock()
	isActive := cm.isActive
	cm.mu.RUnlock()

	if isActive {
		log.Println("ConfigService is already active")
		return &config.StartServiceResponse{
			Success: true,
			Message: "ConfigService is already active",
			Pid:     int32(os.Getpid()),
		}, nil
	}

	// 初始化配置
	if err := cm.Init(); err != nil {
		log.Printf("Failed to initialize config manager: %v", err)
		return &config.StartServiceResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to initialize config manager: %v", err),
			Pid:     0,
		}, nil
	}

	// 设置服务为激活状态
	cm.mu.Lock()
	cm.isActive = true
	cm.mu.Unlock()

	log.Println("ConfigService has been activated successfully")

	return &config.StartServiceResponse{
		Success: true,
		Message: "ConfigService has been activated successfully",
		Pid:     int32(os.Getpid()),
	}, nil
}

// Check 健康检查
func (cm *ConfigManager) Check(ctx context.Context, req *common.HealthCheckRequest) (*common.HealthCheckResponse, error) {
	return &common.HealthCheckResponse{
		Status: common.HealthCheckResponse_SERVING,
	}, nil
}

// Watch 健康检查监听
func (cm *ConfigManager) Watch(req *common.HealthCheckRequest, stream common.HealthCheckService_WatchServer) error {
	// 发送初始状态
	initialStatus := &common.HealthCheckResponse{
		Status: common.HealthCheckResponse_SERVING,
	}
	if err := stream.Send(initialStatus); err != nil {
		return err
	}

	// 保持连接，直到客户端断开
	<-stream.Context().Done()

	return nil
}

// UpdateConfig 更新配置（内部使用）
func (cm *ConfigManager) UpdateConfig(cfg *model.Config) error {
	cm.handleConfigChange(cfg, model.ConfigSourceGRPC)
	return nil
}

// Shutdown 关闭配置管理器
func (cm *ConfigManager) Shutdown() {
	// 停止监听器
	if cm.fileWatcher != nil {
		cm.fileWatcher.Stop()
	}
	if cm.envWatcher != nil {
		cm.envWatcher.Stop()
	}

	// 关闭所有订阅者连接
	cm.mu.Lock()
	for subID := range cm.subscribers {
		delete(cm.subscribers, subID)
	}
	cm.mu.Unlock()

	log.Println("ConfigManager shutdown successfully")
}
