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
    - [协程池](#协程池)
  - [部署运维](#部署运维)
    - [Docker 部署](#docker-部署)
    - [监控告警](#监控告警)
  - [协程池 API 文档](#协程池-api-文档)
    - [协程池选项](#协程池选项)
    - [任务选项](#任务选项)
    - [协程池接口](#协程池接口)
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
- 🧵 协程池 - 控制并发任务数量，支持任务优先级和超时控制

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

### 协程池

协程池用于控制并发任务数量，让协程排队等待执行：

```go
package main

import (
    "fmt"
    "log"
    "time"
    "context"

    "github.com/shrimps80/go-service-utils/pool"
)

func main() {
    // 创建一个大小为5的协程池
    p, err := pool.New(5)
    if err != nil {
        log.Fatalf("创建协程池失败: %v", err)
    }

    // 提交普通任务
    for i := 1; i <= 10; i++ {
        taskID := i
        err := p.Submit(func() error {
            fmt.Printf("执行任务 %d\n", taskID)
            time.Sleep(time.Second)
            return nil
        })
        if err != nil {
            log.Printf("提交任务失败: %v", err)
        }
    }

    // 提交带结果的任务
    future := p.SubmitFunc(func() (int, error) {
        return 42, nil
    })

    // 获取任务结果
    result, err := future.Get(context.Background())
    if err != nil {
        log.Fatalf("获取结果失败: %v", err)
    }
    fmt.Printf("任务结果: %v\n", result)

    // 提交带优先级的任务
    p.SubmitWithOptions(func() error {
        fmt.Println("执行高优先级任务")
        return nil
    }, pool.WithTaskPriority(pool.PriorityHigh))

    // 提交带超时的任务
    p.SubmitWithOptions(func() error {
        fmt.Println("执行长时间任务")
        time.Sleep(5 * time.Second)
        return nil
    }, pool.WithTimeout(2 * time.Second))

    // 等待所有任务完成
    p.Wait()

    // 获取统计信息
    stats := p.Stats()
    fmt.Printf("协程池统计: 运行任务=%d, 等待任务=%d, 已完成任务=%d\n",
        stats.RunningTasks, stats.WaitingTasks, stats.CompletedTasks)

    // 关闭协程池
    p.Close()
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

## 协程池 API 文档

### 协程池选项

```go
// Options 协程池选项
type Options struct {
    // QueueSize 任务队列大小，默认为 size * 10
    QueueSize int

    // EnablePriority 是否启用优先级功能
    EnablePriority bool

    // PanicHandler 处理任务中的 panic
    PanicHandler func(interface{})
}

// WithQueueSize 设置任务队列大小
func WithQueueSize(size int) Option

// WithPriority 启用优先级功能
func WithPriority() Option

// WithPanicHandler 设置 panic 处理函数
func WithPanicHandler(handler func(interface{})) Option
```

### 任务选项

```go
// TaskOptions 任务选项
type TaskOptions struct {
    // Priority 任务优先级，默认为 PriorityNormal
    Priority Priority

    // Timeout 任务超时时间，默认为 0（不超时）
    Timeout time.Duration
}

// WithTaskPriority 设置任务优先级
func WithTaskPriority(priority Priority) TaskOption

// WithTimeout 设置任务超时时间
func WithTimeout(timeout time.Duration) TaskOption
```

### 协程池接口

```go
// Pool 协程池接口
type Pool interface {
    // Submit 提交一个任务到协程池
    Submit(task func() error) error

    // SubmitWithContext 提交一个带上下文的任务到协程池
    SubmitWithContext(ctx context.Context, task func() error) error

    // SubmitWithOptions 提交一个带选项的任务到协程池
    SubmitWithOptions(task func() error, options ...TaskOption) error

    // SubmitFunc 提交一个带结果的任务到协程池，返回Future对象
    SubmitFunc(task interface{}) Future

    // SubmitFuncWithContext 提交一个带上下文和结果的任务到协程池，返回Future对象
    SubmitFuncWithContext(ctx context.Context, task interface{}) Future

    // Wait 等待所有任务完成
    Wait()

    // Close 关闭协程池，不再接受新任务
    Close()

    // IsClosed 检查协程池是否已关闭
    IsClosed() bool

    // Stats 返回协程池的统计信息
    Stats() Stats
}

// Stats 协程池统计信息
type Stats struct {
    // Size 协程池大小（最大并发数）
    Size int

    // RunningTasks 当前正在运行的任务数
    RunningTasks int

    // WaitingTasks 当前等待中的任务数
    WaitingTasks int

    // CompletedTasks 已完成的任务数
    CompletedTasks int

    // TimeoutTasks 超时的任务数
    TimeoutTasks int
}

// Future 表示一个异步任务的未来结果
type Future interface {
    // Get 获取任务结果，阻塞直到任务完成或上下文取消
    Get(ctx context.Context) (interface{}, error)

    // GetWithTimeout 获取任务结果，阻塞直到任务完成、超时或上下文取消
    GetWithTimeout(timeout time.Duration) (interface{}, error)

    // IsDone 检查任务是否已完成
    IsDone() bool
}

// New 创建一个新的协程池
func New(size int, options ...Option) (Pool, error)
```

## 贡献

欢迎提交 Issue 和 Pull Request！
