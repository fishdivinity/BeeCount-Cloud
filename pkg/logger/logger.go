package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 日志接口
// 定义日志记录的通用方法
type Logger interface {
	// Debug 记录调试级别日志
	Debug(msg string, fields ...Field)
	// Info 记录信息级别日志
	Info(msg string, fields ...Field)
	// Warn 记录警告级别日志
	Warn(msg string, fields ...Field)
	// Error 记录错误级别日志
	Error(msg string, fields ...Field)
	// Fatal 记录致命错误日志并退出程序
	Fatal(msg string, fields ...Field)
	// With 添加字段到日志上下文
	With(fields ...Field) Logger
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// ZapLogger Zap日志实现
type ZapLogger struct {
	logger *zap.Logger
}

// NewLogger 创建日志实例
func NewLogger(level, format, output string, fileConfig FileConfig) (Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	var encoder zapcore.Encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	switch format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	var writer zapcore.WriteSyncer
	switch output {
	case "file":
		writer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   fileConfig.Path,
			MaxSize:    fileConfig.MaxSize,
			MaxBackups: fileConfig.MaxBackups,
			MaxAge:     fileConfig.MaxAge,
			Compress:   fileConfig.Compress,
		})
	case "stdout":
		writer = zapcore.AddSync(os.Stdout)
	default:
		writer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writer, zapLevel)
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &ZapLogger{logger: zapLogger}, nil
}

// Debug 记录调试级别日志
func (l *ZapLogger) Debug(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	l.logger.Debug(msg, zapFields...)
}

// Info 记录信息级别日志
func (l *ZapLogger) Info(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	l.logger.Info(msg, zapFields...)
}

// Warn 记录警告级别日志
func (l *ZapLogger) Warn(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	l.logger.Warn(msg, zapFields...)
}

// Error 记录错误级别日志
func (l *ZapLogger) Error(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	l.logger.Error(msg, zapFields...)
}

// Fatal 记录致命错误日志并退出程序
func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	l.logger.Fatal(msg, zapFields...)
}

// With 添加字段到日志上下文
func (l *ZapLogger) With(fields ...Field) Logger {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return &ZapLogger{logger: l.logger.With(zapFields...)}
}

// String 创建字符串字段
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 创建64位整数字段
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 创建64位浮点数字段
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Any 创建任意类型字段
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Error 创建错误字段
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

// FileConfig 文件日志配置
type FileConfig struct {
	Path          string
	MaxSize       int
	MaxBackups    int
	MaxAge        int
	Compress      bool
	MaxTotalSizeGB int
}

// StartCleanupRoutine 启动日志清理例程
func StartCleanupRoutine(fileConfig FileConfig, cleanupIntervalHours int) {
	if cleanupIntervalHours <= 0 {
		cleanupIntervalHours = 24
	}

	ticker := time.NewTicker(time.Duration(cleanupIntervalHours) * time.Hour)
	go func() {
		for range ticker.C {
			cleanupOldLogs(fileConfig)
		}
	}()
}

// cleanupOldLogs 清理旧日志文件
func cleanupOldLogs(fileConfig FileConfig) {
	if fileConfig.MaxTotalSizeGB <= 0 {
		return
	}

	logDir := filepath.Dir(fileConfig.Path)
	maxTotalSize := int64(fileConfig.MaxTotalSizeGB) * 1024 * 1024 * 1024

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return
	}

	var files []os.FileInfo
	var totalSize int64

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, info)
		totalSize += info.Size()
	}

	if totalSize <= maxTotalSize {
		return
	}

	for _, file := range files {
		if totalSize <= maxTotalSize {
			break
		}

		filePath := filepath.Join(logDir, file.Name())
		if err := os.Remove(filePath); err == nil {
			totalSize -= file.Size()
		}
	}
}
