// 包 logger 封装 go.uber.org/zap，提供应用日志实例构建与全局日志句柄。
package logger

import (
	"fmt"
	"os"

	"github.com/praise579/fit_sfm/pkg/fileurl"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config 描述一个日志实例的构建参数。
type Config struct {
	// Level 日志级别，取值参见 zapcore.ParseLevel（如 debug/info/warn/error）。
	Level string `yaml:"level"`

	// File 日志输出文件路径；为空时日志仅输出到 stderr。
	File string `yaml:"file"`

	// Production 为 true 时文件输出采用 JSON 编码，便于日志采集。
	Production bool `yaml:"production"`
}

// 全局默认日志：stderr 控制台输出、级别 info，可通过 SetLevel 动态调整。
var (
	stderr = zapcore.Lock(os.Stderr)
	level  = zap.NewAtomicLevelAt(zap.InfoLevel)
	l      = zap.New(zapcore.NewCore(consoleEncoder(), stderr, level))
	s      = l.Sugar()

	nop = zap.NewNop()
)

func consoleEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
}

// NewLogger 按 Config 构建 zap.Logger：仅有文件时为「stderr 控制台 + 文件」双写，
// 无文件时仅输出到 stderr。
func NewLogger(c Config) (*zap.Logger, error) {
	lvl, err := zapcore.ParseLevel(c.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", c.Level, err)
	}

	if c.File == "" {
		return zap.New(zapcore.NewCore(consoleEncoder(), stderr, lvl)), nil
	}

	// 目标文件可能不存在：先确保其所在目录存在，再由 zap.Open 创建/追加文件。
	if !fileurl.IsExist(c.File) {
		_ = fileurl.CreatePath(c.File, os.ModePerm)
	}
	fileSyncer, _, err := zap.Open(c.File)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	var fileEncoder zapcore.Encoder
	if c.Production {
		fileEncoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	} else {
		fileEncoder = consoleEncoder()
	}

	consoleCore := zapcore.NewCore(consoleEncoder(), stderr, lvl)
	fileCore := zapcore.NewCore(fileEncoder, zapcore.Lock(fileSyncer), lvl)
	return zap.New(zapcore.NewTee(consoleCore, fileCore)), nil
}

// L 返回全局日志实例。
func L() *zap.Logger {
	return l
}

// SetLevel 动态调整全局日志实例的级别。
func SetLevel(lv zapcore.Level) {
	level.SetLevel(lv)
}

// S 返回全局日志实例的 Sugared 版本。
func S() *zap.SugaredLogger {
	return s
}

// Nop 返回一个丢弃所有日志的实例。
func Nop() *zap.Logger {
	return nop
}
