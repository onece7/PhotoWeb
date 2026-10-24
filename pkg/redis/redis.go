package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"novel2comic/backend/internal/config"
)

// Client Redis 客户端封装
type Client struct {
	rdb *redis.Client
}

// NewClient 创建 Redis 客户端
func NewClient(cfg *config.Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis 连接失败: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Set 设置键值，带过期时间
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

// Get 获取键值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Exists(ctx, keys...).Result()
}

// SetNX 仅当键不存在时设置
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, key, value, expiration).Result()
}

// Incr 自增
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

// Expire 设置过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.rdb.Expire(ctx, key, expiration).Result()
}

// TTL 获取剩余有效期
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, key).Result()
}

// Pipeline 返回管道执行器
func (c *Client) Pipeline() redis.Pipeliner {
	return c.rdb.Pipeline()
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.rdb.Close()
}
