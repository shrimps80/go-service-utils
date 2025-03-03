package core

import (
	"github.com/gin-gonic/gin"
	"github.com/shrimps80/go-service-utils/middleware"
	"github.com/shrimps80/go-service-utils/logger"
)

// EngineOptions 定义Gin引擎的配置选项
type EngineOptions struct {
	Mode string // gin模式：debug, release, test
	Log  *logger.Config
}

// DefaultEngineOptions 返回默认的引擎配置
func DefaultEngineOptions() *EngineOptions {
	return &EngineOptions{
		Mode: gin.ReleaseMode,
		Log: &logger.Config{
			Filename:   "app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
			Level:      "info",
		},
	}
}

// NewEngine 创建一个预配置的Gin引擎
func NewEngine(opts *EngineOptions) (*gin.Engine, error) {
	if opts == nil {
		opts = DefaultEngineOptions()
	}

	// 设置Gin模式
	gin.SetMode(opts.Mode)

	// 初始化日志
	log, err := logger.NewLogger(opts.Log)
	if err != nil {
		return nil, err
	}

	// 创建gin引擎
	engine := gin.New()

	// 注册默认中间件
	engine.Use(
		middleware.Recovery(log),  // panic恢复
		middleware.Metrics(),      // 指标收集
		middleware.Health(),       // 健康检查
	)

	return engine, nil
}