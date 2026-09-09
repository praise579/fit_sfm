package logger

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestNewLoggerDefaultConsole(t *testing.T) {
	lg, err := NewLogger(Config{Level: "info"})
	if err != nil {
		t.Fatalf("NewLogger 失败: %v", err)
	}
	lg.Info("hello")
	_ = lg.Sync()

	if L() == nil || S() == nil || Nop() == nil {
		t.Fatal("全局句柄 L/S/Nop 不应为 nil")
	}
	SetLevel(zapcore.DebugLevel) // 不应 panic
}

func TestNewLoggerCreatesFileAndParentDir(t *testing.T) {
	p := filepath.Join(t.TempDir(), "logs", "app.log")
	lg, err := NewLogger(Config{Level: "debug", File: p})
	if err != nil {
		t.Fatalf("NewLogger 失败: %v", err)
	}
	lg.Info("write to file")
	_ = lg.Sync()

	if _, err := os.Stat(p); err != nil {
		t.Fatalf("日志文件未创建: %v", err)
	}
}

func TestNewLoggerInvalidLevel(t *testing.T) {
	if _, err := NewLogger(Config{Level: "not-a-level"}); err == nil {
		t.Fatal("非法日志级别应返回错误")
	}
}
