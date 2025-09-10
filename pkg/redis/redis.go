package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/cheel98/flashcard-backend/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient Redis客户端封装
type RedisClient struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisClient 创建新的Redis客户端
func NewRedisClient(cfg *config.Config, logger *zap.Logger) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis连接成功", zap.String("addr", fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)))

	return &RedisClient{
		client: rdb,
		logger: logger,
	}, nil
}

// Set 设置键值对
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		r.logger.Error("Redis Set失败", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// Get 获取值
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		r.logger.Error("Redis Get失败", zap.String("key", key), zap.Error(err))
		return "", err
	}
	return val, nil
}

// Delete 删除键
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("Redis Delete失败", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// Close 关闭连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// SetCaptcha 设置验证码，过期时间5分钟
func (r *RedisClient) SetCaptcha(ctx context.Context, email, captcha string) error {
	key := fmt.Sprintf("captcha:%s", email)
	return r.Set(ctx, key, captcha, 5*time.Minute)
}

// GetCaptcha 获取验证码
func (r *RedisClient) GetCaptcha(ctx context.Context, email string) (string, error) {
	key := fmt.Sprintf("captcha:%s", email)
	return r.Get(ctx, key)
}

// DeleteCaptcha 删除验证码
func (r *RedisClient) DeleteCaptcha(ctx context.Context, email string) error {
	key := fmt.Sprintf("captcha:%s", email)
	return r.Delete(ctx, key)
}
