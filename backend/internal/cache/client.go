package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	Available() bool
}

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(redisURL string) Client {
	if redisURL == "" {
		log.Printf("cache disabled: REDIS_URL not set")
		return &NoopClient{}
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("cache disabled: invalid REDIS_URL: %v", err)
		return &NoopClient{}
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Printf("cache disabled: redis ping failed: %v", err)
		return &NoopClient{}
	}
	log.Printf("cache ready: redis")
	return &RedisClient{client: client}
}

func (c *RedisClient) Available() bool { return c != nil && c.client != nil }

func (c *RedisClient) Get(ctx context.Context, key string) ([]byte, bool, error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return value, true, nil
}

func (c *RedisClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *RedisClient) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 100 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

type NoopClient struct{}

func (c *NoopClient) Available() bool { return false }
func (c *NoopClient) Get(ctx context.Context, key string) ([]byte, bool, error) {
	return nil, false, nil
}
func (c *NoopClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}
func (c *NoopClient) Delete(ctx context.Context, keys ...string) error { return nil }
func (c *NoopClient) DeleteByPattern(ctx context.Context, pattern string) error {
	return nil
}
