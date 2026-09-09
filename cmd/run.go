package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/praise579/fit_sfm/pkg/fileurl"

	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type runFlags struct {
	dir     string // 项目根目录
	port    string // 启动端口
	runMode string // 启动模式
	config  string // 指定要使用的配置文件路径
}

var (
	runEnv = new(runFlags)
)

func init() {
	runCommand := &cobra.Command{
		Use:   "run [-c config_file] [-d working_dir] [-p port]",
		Short: "Run service",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		RunE: runServer,
	}

	rootCmd.AddCommand(runCommand)

	fs := runCommand.Flags()
	fs.StringVarP(&runEnv.dir, "dir", "d", "", "working directory")
	fs.StringVarP(&runEnv.port, "port", "p", "", "server port")
	fs.StringVarP(&runEnv.runMode, "mode", "m", "", "run mode (dev/prod)")
	fs.StringVarP(&runEnv.config, "config", "c", "", "config file path")
}

func runServer(cmd *cobra.Command, args []string) error {
	// 创建主 context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化服务器
	server, err := NewServer(runEnv)
	if err != nil {
		log.Printf("Failed to initialize server: %v\n", err)
		return fmt.Errorf("server initialization failed: %w", err)
	}

	// 启动配置文件监听
	configWatcher, err := startConfigWatcher(ctx, server)
	if err != nil {
		log.Printf("Failed to start config watcher: %v\n", err)
		// 配置监听失败不应该阻止服务器启动
	} else {
		defer configWatcher.Close()
	}

	// 设置信号处理
	setupSignalHandler(cancel, server)

	// 等待关闭信号
	<-ctx.Done()

	// 优雅关闭
	server.logger.Info("Server shutting down...")
	server.sc.SendCloseSignal(nil)

	// 等待服务器完全关闭
	time.Sleep(time.Second)

	server.logger.Info("Server shutdown complete")
	return nil
}

// setupSignalHandler 设置系统信号处理
func setupSignalHandler(cancel context.CancelFunc, server *Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v\n", sig)
		server.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		cancel()
	}()
}

// ConfigWatcher 配置文件监听器
type ConfigWatcher struct {
	watcher *watcher.Watcher
	ctx     context.Context
	cancel  context.CancelFunc
	logger  *zap.Logger
}

// startConfigWatcher 启动配置文件监听
func startConfigWatcher(parentCtx context.Context, server *Server) (*ConfigWatcher, error) {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write)

	// 添加配置文件监听
	if err := w.Add(runEnv.config); err != nil {
		return nil, fmt.Errorf("failed to watch config file: %w", err)
	}

	ctx, cancel := context.WithCancel(parentCtx)
	cw := &ConfigWatcher{
		watcher: w,
		ctx:     ctx,
		cancel:  cancel,
		logger:  server.logger,
	}

	// 启动事件处理
	go cw.handleEvents(server)

	// 启动监听
	go func() {
		if err := w.Start(5 * time.Second); err != nil {
			cw.logger.Error("Config watcher start error", zap.Error(err))
		}
	}()

	cw.logger.Info("Config file watcher started", zap.String("file", runEnv.config))
	return cw, nil
}

// handleEvents 处理配置文件变更事件
func (cw *ConfigWatcher) handleEvents(currentServer *Server) {
	// 使用 debounce 避免频繁重启
	var lastReload time.Time
	reloadInterval := 3 * time.Second

	for {
		select {
		case <-cw.ctx.Done():
			cw.logger.Info("Config watcher stopped")
			return

		case event := <-cw.watcher.Event:
			// 防抖动处理
			if time.Since(lastReload) < reloadInterval {
				cw.logger.Debug("Config change ignored (too soon)",
					zap.Duration("since_last_reload", time.Since(lastReload)))
				continue
			}

			cw.logger.Info("Config file changed, reloading...",
				zap.String("event", event.Op.String()),
				zap.String("file", event.Path))

			// 等待一小段时间确保文件写入完成
			time.Sleep(500 * time.Millisecond)

			if err := cw.reloadServer(currentServer); err != nil {
				cw.logger.Error("Failed to reload server", zap.Error(err))
			} else {
				lastReload = time.Now()
				cw.logger.Info("Server reloaded successfully")
			}

		case err := <-cw.watcher.Error:
			cw.logger.Error("Config watcher error", zap.Error(err))

		case <-cw.watcher.Closed:
			cw.logger.Info("Config watcher closed")
			return
		}
	}
}

// reloadServer 重新加载服务器配置
func (cw *ConfigWatcher) reloadServer(currentServer *Server) error {
	// 通知当前服务器关闭
	currentServer.sc.SendCloseSignal(nil)

	// 等待服务器关闭
	time.Sleep(2 * time.Second)

	// 创建新的服务器实例
	newServer, err := NewServer(runEnv)
	if err != nil {
		return fmt.Errorf("failed to create new server: %w", err)
	}

	// 更新 logger 引用
	cw.logger = newServer.logger

	cw.logger.Info("Server reloaded with new configuration")
	return nil
}

// Close 关闭配置监听器
func (cw *ConfigWatcher) Close() {
	cw.cancel()
	cw.watcher.Close()
}

// initConfig 初始化配置文件
func initConfig() error {
	// 切换工作目录
	if err := changeWorkingDir(runEnv.dir); err != nil {
		return err
	}

	// 查找或创建配置文件
	configPath, err := findOrCreateConfig()
	if err != nil {
		return err
	}

	runEnv.config = configPath
	log.Printf("Using config file: %s\n", configPath)
	return nil
}

// changeWorkingDir 切换工作目录
func changeWorkingDir(dir string) error {
	if dir == "" {
		return nil
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("failed to change working directory to %s: %w", dir, err)
	}

	absDir, _ := filepath.Abs(dir)
	log.Printf("Working directory changed to: %s\n", absDir)
	return nil
}

// findOrCreateConfig 查找或创建配置文件
func findOrCreateConfig() (string, error) {
	// 如果指定了配置文件,直接使用
	if runEnv.config != "" {
		if fileurl.IsExist(runEnv.config) {
			return runEnv.config, nil
		}
		log.Printf("Warning: Specified config file not found: %s\n", runEnv.config)
	}

	// 按优先级查找配置文件
	configPaths := []string{
		"config/config-dev.yaml",
		"config.yaml",
		"config/config.yaml",
	}

	for _, path := range configPaths {
		if fileurl.IsExist(path) {
			return path, nil
		}
	}

	// 如果都不存在,创建默认配置
	defaultPath := "config/config.yaml"
	if err := createDefaultConfig(defaultPath); err != nil {
		return "", err
	}

	return defaultPath, nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(path string) error {
	log.Printf("Config file not found, creating default: %s\n", path)

	// 创建目录
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 创建文件
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// 写入默认配置
	if _, err := file.WriteString(configDefault); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	log.Printf("✓ Default config file created: %s\n", path)
	printConfigInstructions()

	return nil
}

// printConfigInstructions 打印配置说明
func printConfigInstructions() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("⚠️  IMPORTANT: Please configure the following settings")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n📝 Edit the config file and set:")
	fmt.Println("   1. Server password for authentication")
	fmt.Println("   2. node_file_url - URL to your node subscription")
	fmt.Println("   3. template_url - URL to your configuration template")
	fmt.Println("\n💡 After editing, restart the server")
	fmt.Println(strings.Repeat("=", 60) + "\n")
}
