package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
	key    string
}

func NewRedisQueue(client *redis.Client, key string) *RedisQueue {
	if key == "" {
		key = "fileai:jobs"
	}
	return &RedisQueue{client: client, key: key}
}

func (q *RedisQueue) Enqueue(ctx context.Context, job Job) error {
	if q == nil || q.client == nil {
		return errors.New("redis queue is not configured")
	}
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, q.key, payload).Err()
}

func (q *RedisQueue) Dequeue(ctx context.Context) (Job, error) {
	if q == nil || q.client == nil {
		return Job{}, errors.New("redis queue is not configured")
	}

	result, err := q.client.BRPop(ctx, 5*time.Second, q.key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Job{}, err
		}
		return Job{}, err
	}
	if len(result) < 2 {
		return Job{}, errors.New("invalid redis queue payload")
	}

	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return Job{}, err
	}
	return job, nil
}
