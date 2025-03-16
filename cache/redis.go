package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Addrs      []string      // Redis地址列表
	Password   string        // 密码
	DB         int           // 数据库编号
	PoolSize   int           // 连接池大小
	MaxRetries int           // 最大重试次数
	Timeout    time.Duration // 超时时间
}

// DefaultRedisConfig 返回默认Redis配置
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addrs:      []string{"localhost:6379"},
		DB:         0,
		PoolSize:   10,
		MaxRetries: 3,
		Timeout:    time.Second * 5,
	}
}

// Redis Redis客户端封装
type Redis struct {
	client redis.UniversalClient
}

// NewRedis 创建Redis客户端
func NewRedis(cfg *RedisConfig) (*Redis, error) {
	if cfg == nil {
		cfg = DefaultRedisConfig()
	}

	// 创建通用客户端（自动判断单机/集群模式）
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:      cfg.Addrs,
		Password:   cfg.Password,
		DB:         cfg.DB,
		PoolSize:   cfg.PoolSize,
		MaxRetries: cfg.MaxRetries,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Redis{client: client}, nil
}

// Get 获取缓存
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set 设置缓存
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Del 删除缓存
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Close 关闭连接
func (r *Redis) Close() error {
	return r.client.Close()
}

// Pipeline 管道操作
func (r *Redis) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// Subscribe 订阅频道
func (r *Redis) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// Publish 发布消息
func (r *Redis) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Hset 设置哈希表字段值
func (r *Redis) Hset(ctx context.Context, key string, field string, value interface{}) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

// Hget 获取哈希表字段值
func (r *Redis) Hget(ctx context.Context, key string, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取哈希表所有字段和值
func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// ExpireAt 设置过期时间
func (r *Redis) ExpireAt(ctx context.Context, key string, expiration time.Time) error {
	return r.client.ExpireAt(ctx, key, expiration).Err()
}
