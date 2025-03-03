# Go Service Utils

这是一个基于Gin框架的微服务工具包，提供了常用的运维功能，包括健康检查、指标监控、日志管理、缓存、数据库集成和追踪等特性。

## 功能特性

- 健康检查中间件
- Prometheus指标收集
- Panic恢复和日志记录
- 日志轮转和压缩
- Redis缓存集成
- 数据库连接管理
- 分布式追踪
- 服务核心引擎
- 配置管理（支持动态加载）
- 邮件通知功能

## 快速开始

### 安装

```bash
go get github.com/shrimps80/go-service-utils
```

### 基础使用示例

```go
package main

import (
    "github.com/shrimps80/go-service-utils/core"
    "github.com/shrimps80/go-service-utils/logger"
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

## 高级功能

### Redis缓存集成

```go
package main

import (
    "context"
    "github.com/shrimps80/go-service-utils/cache"
    "time"
)

func main() {
    // 配置Redis客户端
    redisConfig := &cache.RedisConfig{
        Addrs:      []string{"localhost:6379"},
        Password:   "",
        DB:         0,
        PoolSize:   10,
        MaxRetries: 3,
        Timeout:    time.Second * 5,
    }

    // 创建Redis客户端
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

### 配置管理

```go
package main

import (
    "fmt"
    "github.com/fsnotify/fsnotify"
    "github.com/shrimps80/go-service-utils/config"
)

func main() {
    // 配置选项
    opts := &config.Options{
        ConfigType: "yaml",
        ConfigName: "config",
        ConfigPaths: []string{".", "./config"},
    }

    // 创建配置管理器
    cfg, err := config.New(opts)
    if err != nil {
        panic(err)
    }

    // 监听配置文件变化
    cfg.OnConfigChange(func(e fsnotify.Event) {
        fmt.Printf("配置文件已更新: %s，事件类型: %s\n", e.Name, e.Op)
        if e.Op&fsnotify.Write == fsnotify.Write {
            // 获取更新后的配置
            newValue := cfg.GetString("app.name")
            fmt.Printf("新的应用名称: %s\n", newValue)
        }
    })

    // 启动配置文件监听
    cfg.StartWatch()

    // 获取配置值
    appName := cfg.GetString("app.name")
    port := cfg.GetInt("server.port")
    debug := cfg.GetBool("app.debug")

    // 获取所有配置
    allConfig := cfg.GetAll()
    fmt.Printf("完整配置: %+v\n", allConfig)

    // 获取当前使用的配置文件
    configFile := cfg.GetConfigFile()
    fmt.Printf("当前配置文件: %s\n", configFile)

    // 修改配置
    cfg.Set("app.version", "1.0.1")
    
    // 保存配置到文件
    if err := cfg.WriteConfig(); err != nil {
        panic(err)
    }

    // 如果需要停止监听配置变更
    // cfg.StopWatch()
}
```

### 数据库集成

```go
package main

import (
    "github.com/shrimps80/go-service-utils/database"
    "time"
)

func main() {
    // 配置数据库连接
    dbConfig := &database.Config{
        Type:         "mysql",
        DSN:          "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
        MaxIdleConns: 10,
        MaxOpenConns: 100,
        MaxLifetime:  time.Hour,
        Debug:        true,
    }

    // 创建数据库连接
    db, err := database.New(dbConfig)
    if err != nil {
        panic(err)
    }

    // 使用GORM进行数据库操作
    type User struct {
        ID   uint
        Name string
    }

    // 自动迁移
    db.AutoMigrate(&User{})

    // 创建记录
    user := User{Name: "test"}
    db.Create(&user)
}
```

### 邮件通知

```go
package main

import (
    "github.com/shrimps80/go-service-utils/notification"
)

func main() {
    // 配置邮件客户端
    mailConfig := &notification.MailConfig{
        Host:     "smtp.example.com",
        Port:     587,
        Username: "your-email@example.com",
        Password: "your-password",
        From:     "your-email@example.com",
        UseTLS:   true,
    }

    // 创建邮件客户端
    mailClient := notification.NewMailClient(mailConfig)

    // 添加邮件模板
    template := &notification.MailTemplate{
        Subject: "告警通知: {{.Title}}",
        Body:    "<h1>{{.Title}}</h1><p>{{.Content}}</p>",
    }
    mailClient.AddTemplate("alert", template)

    // 发送邮件
    data := map[string]string{
        "Title":   "服务异常",
        "Content": "CPU使用率超过90%",
    }
    err := mailClient.SendMail(
        []string{"admin@example.com"},
        "alert",
        data,
    )
    if err != nil {
        panic(err)
    }
}
```

## 部署与运维

### Docker部署

```dockerfile
FROM golang:1.22-alpine

WORKDIR /app
COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["/app/main"]
```

### 资源配置

```yaml
# docker-compose.yml
version: '3'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### 监控配置

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

  - alert: SlowResponses
    expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: Slow response times detected
```

### 负载均衡

```nginx
# nginx.conf
upstream backend {
    server localhost:8080;
    check interval=3000 rise=2 fall=5 timeout=1000 type=http;
    check_http_send "HEAD /health HTTP/1.0\r\n\r\n";
    check_http_expect_alive http_2xx;
}

server {
    listen 80;
    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```