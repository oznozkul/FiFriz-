package jobs

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewQueue(redisURL, queueKey string) Queue {
	if redisURL != "" {
		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("redis queue disabled: invalid REDIS_URL: %v", err)
			return NewChannelQueue(128)
		}
		client := redis.NewClient(opts)
		if err := client.Ping(context.Background()).Err(); err != nil {
			log.Printf("redis queue fallback to in-memory: ping failed: %v", err)
			return NewChannelQueue(128)
		}
		log.Printf("job queue ready: redis key=%s", queueKey)
		return NewRedisQueue(client, queueKey)
	}

	log.Printf("job queue ready: in-memory channel (REDIS_URL not set)")
	return NewChannelQueue(128)
}
