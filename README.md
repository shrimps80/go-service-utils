# Go Service Utils
这是一个基于 Gin 框架的微服务工具包，提供了常用的运维功能，包括健康检查、指标监控、日志管理、缓存、数据库集成和追踪等特性。

## 目录

- [Go Service Utils](#go-service-utils)
  - [目录](#目录)
  - [功能特性](#功能特性)
  - [快速开始](#快速开始)
    - [安装](#安装)
    - [基础使用](#基础使用)
  - [核心组件](#核心组件)
    - [配置管理](#配置管理)
    - [缓存集成](#缓存集成)
  - [部署运维](#部署运维)
    - [Docker 部署](#docker-部署)
    - [监控告警](#监控告警)
  - [贡献](#贡献)

## 功能特性

- ✨ 健康检查中间件 - 提供服务健康状态监控
- 📊 Prometheus 指标收集 - 支持自定义指标和默认服务指标
- 🛡️ Panic 恢复和日志记录 - 自动捕获并记录异常
- 📝 日志轮转和压缩 - 支持按大小、时间的日志管理
- 🚀 Redis 缓存集成 - 支持集群和哨兵模式
- 💾 数据库连接管理 - 支持 MySQL、PostgreSQL、SQLite
- 🔍 分布式追踪 - OpenTelemetry 集成
- 🎯 服务核心引擎 - 基于 Gin 的增强功能
- ⚙️ 配置管理 - 支持多种格式和动态加载
- 📧 邮件通知 - 支持模板和 HTML 格式

## 快速开始

### 安装

```bash
go get github.com/shrimps80/go-service-utils
```

### 基础使用

创建一个简单的 HTTP 服务：

```go
package main

import (
    "github.com/shrimps80/go-service-utils/core"
    "github.com/shrimps80/go-service-utils/logger"
    "github.com/gin-gonic/gin"
)

func main() {
    // 配置服务引擎
    opts := core.DefaultEngineOptions()
    opts.Mode = "debug"
    opts.Log = &logger.Config{
        Filename:   "/var/log/app.log",
        MaxSize:    100,  // 100MB
        MaxBackups: 3,
        MaxAge:     7,    // 7天
        Compress:   true,
        Level:      "info",
    }

    // 创建服务引擎
    engine, err := core.NewEngine(opts)
    if err != nil {
        panic(err)
    }

    // 注册路由
    engine.GET("/hello", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello World!"})
    })

    // 启动服务
    engine.Run(":8080")
}
```

## 核心组件

### 配置管理

支持 YAML、JSON、TOML 等多种格式的配置文件，支持配置热重载：

```go
package main

import (
    "fmt"
    "github.com/fsnotify/fsnotify"
    "github.com/shrimps80/go-service-utils/config"
)

func main() {
    opts := &config.Options{
        ConfigType: "yaml",
        ConfigName: "config",
        ConfigPaths: []string{".", "./config"},
    }

    cfg, err := config.New(opts)
    if err != nil {
        panic(err)
    }

    // 监听配置变化
    cfg.OnConfigChange(func(e fsnotify.Event) {
        fmt.Printf("配置文件已更新: %s\n", e.Name)
    })
    cfg.StartWatch()

    // 获取配置
    appName := cfg.GetString("app.name")
    port := cfg.GetInt("server.port")
    fmt.Printf("应用: %s, 端口: %d\n", appName, port)
}
```

### 缓存集成

 Redis 缓存支持：

```go
package main

import (
    "context"
    "github.com/shrimps80/go-service-utils/cache"
    "time"
)

func main() {
    redisConfig := &cache.RedisConfig{
        Addrs:      []string{"localhost:6379"},
        Password:   "",
        DB:         0,
        PoolSize:   10,
        MaxRetries: 3,
        Timeout:    time.Second * 5,
    }

    redis, err := cache.NewRedis(redisConfig)
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    
    // 设置缓存
    err = redis.Set(ctx, "key", "value", time.Hour)
    if err != nil {
        panic(err)
    }

    // 获取缓存
    val, err := redis.Get(ctx, "key")
    if err != nil {
        panic(err)
    }
}
```

## 部署运维

### Docker 部署

```dockerfile
FROM golang:1.22-alpine

WORKDIR /app
COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["/app/main"]
```

### 监控告警

Prometheus 配置示例：

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'your-service'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']

# alerts.yml
groups:
- name: service_alerts
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: High error rate detected
```

## 贡献

欢迎提交 Issue 和 Pull Request！
