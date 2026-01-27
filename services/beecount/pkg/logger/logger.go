package logger

import (
	"log"
	"os"
	"strings"
)

// LogLevel 日志级别类型
type LogLevel int

// 定义日志级别
const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	FATAL
)

var (
	// 当前日志级别
	currentLevel = INFO

	// 日志输出前缀
	prefix = "[BeeCount-Cloud] "

	// 日志记录器
	debugLogger = log.New(os.Stdout, prefix+"[DEBUG] ", log.Ldate|log.Ltime)
	infoLogger  = log.New(os.Stdout, prefix+"[INFO] ", log.Ldate|log.Ltime)
	warnLogger  = log.New(os.Stdout, prefix+"[WARNING] ", log.Ldate|log.Ltime)
	errLogger   = log.New(os.Stderr, prefix+"[ERROR] ", log.Ldate|log.Ltime)
	fatalLogger = log.New(os.Stderr, prefix+"[FATAL] ", log.Ldate|log.Ltime)
)

// SetLevel 设置日志级别
func SetLevel(level string) {
	level = strings.ToUpper(level)
	switch level {
	case "DEBUG":
		currentLevel = DEBUG
	case "INFO":
		currentLevel = INFO
	case "WARNING":
		currentLevel = WARNING
	case "ERROR":
		currentLevel = ERROR
	case "FATAL":
		currentLevel = FATAL
	}
}

// Debug 记录调试日志
func Debug(format string, v ...any) {
	if currentLevel <= DEBUG {
		debugLogger.Printf(format, v...)
	}
}

// Info 记录信息日志
func Info(format string, v ...any) {
	if currentLevel <= INFO {
		infoLogger.Printf(format, v...)
	}
}

// Warning 记录警告日志
func Warning(format string, v ...any) {
	if currentLevel <= WARNING {
		warnLogger.Printf(format, v...)
	}
}

// Error 记录错误日志
func Error(format string, v ...any) {
	if currentLevel <= ERROR {
		errLogger.Printf(format, v...)
	}
}

// Fatal 记录致命错误日志并退出
func Fatal(format string, v ...any) {
	if currentLevel <= FATAL {
		fatalLogger.Fatalf(format, v...)
	}
}
