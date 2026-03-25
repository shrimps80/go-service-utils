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

// Exists 检查键是否存在
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// ExpireAt 设置过期时间
func (r *Redis) ExpireAt(ctx context.Context, key string, expiration time.Time) error {
	return r.client.ExpireAt(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Keys 获取所有匹配的键（生产环境慎用，性能较差）
func (r *Redis) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// Scan 使用游标遍历键（推荐用于生产环境）
func (r *Redis) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.client.Scan(ctx, cursor, match, count).Result()
}

// ScanKeys 使用 Scan 获取所有匹配的键（封装好的方法）
func (r *Redis) ScanKeys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		var batch []string
		var err error

		// 每次扫描 100 个键
		batch, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, batch...)

		// 游标为 0 表示扫描完成
		if cursor == 0 {
			break
		}
	}

	return keys, nil
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
func (r *Redis) Hset(ctx context.Context, key string, value ...interface{}) error {
	return r.client.HSet(ctx, key, value).Err()
}

// Hget 获取哈希表字段值
func (r *Redis) Hget(ctx context.Context, key string, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取哈希表所有字段和值
func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希表字段
func (r *Redis) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// Incr 自增
func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrBy 增加指定值
func (r *Redis) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// GetClient 获取底层客户端（供高级操作使用）
func (r *Redis) GetClient() redis.UniversalClient {
	return r.client
}
