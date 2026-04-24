package cmb

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	// LogLevelDebug 调试级别，输出所有日志
	LogLevelDebug LogLevel = iota
	// LogLevelInfo 信息级别
	LogLevelInfo
	// LogLevelWarn 警告级别
	LogLevelWarn
	// LogLevelError 错误级别
	LogLevelError
	// LogLevelSilent 静默级别，不输出任何日志
	LogLevelSilent
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelSilent:
		return "SILENT"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志接口
// 外部可以实现此接口来对接自己的日志系统（如 zap、logrus、slog 等）
//
// 提供两种风格：
//   - 结构化风格（zap-style）：Debug/Info/Warn/Error，使用 key-value 对
//   - 格式化风格（printf-style）：Debugf/Infof/Warnf/Errorf，使用 format + args
type Logger interface {
	// ===== 结构化风格（zap-style key-value） =====

	// Debug 调试日志，用于输出请求/响应详情、签名字符串等调试信息
	Debug(msg string, keysAndValues ...interface{})

	// Info 信息日志，用于输出请求开始/完成等常规信息
	Info(msg string, keysAndValues ...interface{})

	// Warn 警告日志，用于输出可恢复的异常情况
	Warn(msg string, keysAndValues ...interface{})

	// Error 错误日志，用于输出请求失败、解密失败等错误信息
	Error(msg string, keysAndValues ...interface{})

	// ===== 格式化风格（printf-style） =====

	// Debugf 格式化调试日志
	Debugf(format string, args ...interface{})

	// Infof 格式化信息日志
	Infof(format string, args ...interface{})

	// Warnf 格式化警告日志
	Warnf(format string, args ...interface{})

	// Errorf 格式化错误日志
	Errorf(format string, args ...interface{})
}

// ================ 默认日志实现 ================

// defaultLogger 默认日志实现，基于标准库 log
type defaultLogger struct {
	level  LogLevel
	logger *log.Logger
	mu     sync.Mutex
}

// NewDefaultLogger 创建默认日志实例
// level: 日志级别，低于此级别的日志不会输出
// writer: 日志输出目标，传 nil 则使用 os.Stderr
// 如需同时输出到多个目标，可传入 io.MultiWriter(os.Stdout, file)
func NewDefaultLogger(level LogLevel, writer io.Writer) Logger {
	if writer == nil {
		writer = os.Stderr
	}
	return &defaultLogger{
		level:  level,
		logger: log.New(writer, "", 0),
	}
}

// NewDefaultLoggerWithFile 创建同时输出到标准输出和日志文件的日志实例
// level: 日志级别
// filePath: 日志文件路径（自动创建，追加写入）
// 返回: Logger 实例和可能的错误
//
// 使用示例：
//
//	logger, err := cmb.NewDefaultLoggerWithFile(cmb.LogLevelDebug, "/var/log/cmb-sdk.log")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	config.Logger = logger
func NewDefaultLoggerWithFile(level LogLevel, filePath string) (Logger, error) {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file failed: %w", err)
	}
	writer := io.MultiWriter(os.Stdout, f)
	return &defaultLogger{
		level:  level,
		logger: log.New(writer, "", 0),
	}, nil
}

// ---- 结构化风格 ----

func (l *defaultLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.log(LogLevelDebug, msg, keysAndValues...)
}

func (l *defaultLogger) Info(msg string, keysAndValues ...interface{}) {
	l.log(LogLevelInfo, msg, keysAndValues...)
}

func (l *defaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.log(LogLevelWarn, msg, keysAndValues...)
}

func (l *defaultLogger) Error(msg string, keysAndValues ...interface{}) {
	l.log(LogLevelError, msg, keysAndValues...)
}

// ---- 格式化风格 ----

func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	l.logFormatted(LogLevelDebug, format, args...)
}

func (l *defaultLogger) Infof(format string, args ...interface{}) {
	l.logFormatted(LogLevelInfo, format, args...)
}

func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	l.logFormatted(LogLevelWarn, format, args...)
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	l.logFormatted(LogLevelError, format, args...)
}

// ---- 内部方法 ----

func (l *defaultLogger) log(level LogLevel, msg string, keysAndValues ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	kvStr := formatKVPairs(keysAndValues...)

	if kvStr != "" {
		l.logger.Printf("[CMB-SDK] %s [%s] %s %s", timestamp, level.String(), msg, kvStr)
	} else {
		l.logger.Printf("[CMB-SDK] %s [%s] %s", timestamp, level.String(), msg)
	}
}

func (l *defaultLogger) logFormatted(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	l.logger.Printf("[CMB-SDK] %s [%s] %s", timestamp, level.String(), msg)
}

// formatKVPairs 将 key-value 对格式化为字符串
func formatKVPairs(keysAndValues ...interface{}) string {
	if len(keysAndValues) == 0 {
		return ""
	}

	result := ""
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		key := fmt.Sprintf("%v", keysAndValues[i])
		value := keysAndValues[i+1]
		if result != "" {
			result += " "
		}
		result += fmt.Sprintf("%s=%v", key, value)
	}

	// 奇数个参数时，最后一个作为值追加
	if len(keysAndValues)%2 != 0 {
		if result != "" {
			result += " "
		}
		result += fmt.Sprintf("EXTRA_VALUE=%v", keysAndValues[len(keysAndValues)-1])
	}

	return result
}

// ================ 空日志实现（静默） ================

// nopLogger 空日志实现，不输出任何日志
type nopLogger struct{}

// NewNopLogger 创建空日志实例，所有日志都会被丢弃
func NewNopLogger() Logger {
	return &nopLogger{}
}

func (l *nopLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (l *nopLogger) Info(msg string, keysAndValues ...interface{})  {}
func (l *nopLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (l *nopLogger) Error(msg string, keysAndValues ...interface{}) {}
func (l *nopLogger) Debugf(format string, args ...interface{})      {}
func (l *nopLogger) Infof(format string, args ...interface{})       {}
func (l *nopLogger) Warnf(format string, args ...interface{})       {}
func (l *nopLogger) Errorf(format string, args ...interface{})      {}
