package internal

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fishdivinity/BeeCount-Cloud/common/proto/common"
	"github.com/fishdivinity/BeeCount-Cloud/common/proto/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LogConfig 日志配置
type LogConfig struct {
	Level      string
	Format     string
	Output     string
	FileConfig FileConfig
}

// FileConfig 文件日志配置
type FileConfig struct {
	Path           string
	MaxSize        int // MB
	MaxBackups     int
	MaxAge         int // days
	Compress       bool
	MaxTotalSizeGB int
}

// LogService 日志服务实现
type LogService struct {
	log.UnimplementedLogServiceServer
	common.UnimplementedHealthCheckServiceServer
	logger     zerolog.Logger
	config     LogConfig
	fileWriter *os.File
	mu         sync.Mutex
	level      zerolog.Level
}

// NewLogService 创建日志服务实例
func NewLogService() *LogService {
	return &LogService{
		level: zerolog.InfoLevel,
	}
}

// Configure 配置日志服务
func (s *LogService) Configure(config LogConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config

	// 设置错误堆栈跟踪
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// 设置日志级别
	setLogLevel(config.Level)

	// 初始化输出
	var output io.Writer

	switch config.Output {
	case "file":
		// 初始化文件输出
		if err := s.initFileOutput(); err != nil {
			return err
		}
		output = s.fileWriter
		// 添加文件轮换
		go s.rotateLogs()
	default:
		// 默认输出到控制台
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// 设置日志格式化
	if config.Format == "json" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			FormatLevel: func(i interface{}) string {
				return fmt.Sprintf("| %-6s |", i)
			},
			FormatMessage: func(i interface{}) string {
				return fmt.Sprintf("%s", i)
			},
			FormatFieldName: func(i interface{}) string {
				return fmt.Sprintf("%s:", i)
			},
			FormatFieldValue: func(i interface{}) string {
				return fmt.Sprintf("%s", i)
			},
		}
	}

	// 创建logger
	s.logger = zerolog.New(output).With().Timestamp().Logger()
	return nil
}

// initFileOutput 初始化文件输出
func (s *LogService) initFileOutput() error {
	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(s.config.FileConfig.Path), 0755); err != nil {
		return err
	}

	// 打开日志文件
	file, err := os.OpenFile(s.config.FileConfig.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	s.fileWriter = file
	return nil
}

// setLogLevel 设置日志级别
func setLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// Log 记录日志
func (s *LogService) Log(ctx context.Context, req *log.LogRequest) (*common.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 将log日志级别转换为zerolog级别
	zlevel := convertToZerologLevel(req.Entry.Level)

	// 创建日志事件
	event := s.logger.WithLevel(zlevel)

	// 添加服务名称
	if req.Entry.ServiceName != "" {
		event = event.Str("service", req.Entry.ServiceName)
	}

	// 添加跟踪ID
	if req.Entry.TraceId != "" {
		event = event.Str("trace_id", req.Entry.TraceId)
	}

	// 添加自定义字段
	for k, v := range req.Entry.Fields {
		event = event.Str(k, v)
	}

	// 输出日志
	event.Msg(req.Entry.Message)

	return &common.Response{
		Success: true,
		Message: "Log recorded successfully",
		Code:    200,
	}, nil
}

// SetLogLevel 设置日志级别
func (s *LogService) SetLogLevel(ctx context.Context, req *log.SetLogLevelRequest) (*common.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 更新配置
	s.config.Level = req.Level.String()

	// 设置日志级别
	setLogLevel(s.config.Level)

	return &common.Response{
		Success: true,
		Message: "Log level updated successfully",
		Code:    200,
	}, nil
}

// Check 健康检查
func (s *LogService) Check(ctx context.Context, req *common.HealthCheckRequest) (*common.HealthCheckResponse, error) {
	return &common.HealthCheckResponse{
		Status: common.HealthCheckResponse_SERVING,
	}, nil
}

// Watch 健康检查监听
func (s *LogService) Watch(req *common.HealthCheckRequest, stream common.HealthCheckService_WatchServer) error {
	// 实现健康检查监听逻辑
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}

// rotateLogs 轮换日志文件
func (s *LogService) rotateLogs() {
	// 每小时检查一次日志文件大小
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		if s.fileWriter != nil {
			// 检查文件大小
			if info, err := s.fileWriter.Stat(); err == nil {
				if info.Size() >= int64(s.config.FileConfig.MaxSize*1024*1024) {
					// 关闭当前文件
					s.fileWriter.Close()
					// 重命名当前文件
					timestamp := time.Now().Format("20060102150405")
					newPath := fmt.Sprintf("%s.%s", s.config.FileConfig.Path, timestamp)
					os.Rename(s.config.FileConfig.Path, newPath)
					// 重新打开文件
					s.initFileOutput()
				}
			}
		}
		s.mu.Unlock()
	}
}

// convertToZerologLevel 将proto日志级别转换为zerolog级别
func convertToZerologLevel(level log.LogLevel) zerolog.Level {
	switch level {
	case log.LogLevel_DEBUG:
		return zerolog.DebugLevel
	case log.LogLevel_INFO:
		return zerolog.InfoLevel
	case log.LogLevel_WARNING:
		return zerolog.WarnLevel
	case log.LogLevel_ERROR:
		return zerolog.ErrorLevel
	case log.LogLevel_FATAL:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}
