package main

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps go-redis Client to implement RedisCache interface
// (for testability and abstraction)
type RedisClient struct {
	*redis.Client
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.Client.Set(ctx, key, value, expiration)
}

// NewRedisClient returns a RedisCache (or nil if no redisAddr)
func NewRedisClient(redisAddr string) RedisCache {
	if redisAddr == "" {
		return nil
	}
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &RedisClient{client}
}
