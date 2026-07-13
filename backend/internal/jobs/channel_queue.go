package jobs

import (
	"context"
)

type ChannelQueue struct {
	jobs chan Job
}

func NewChannelQueue(buffer int) *ChannelQueue {
	if buffer < 1 {
		buffer = 64
	}
	return &ChannelQueue{jobs: make(chan Job, buffer)}
}

func (q *ChannelQueue) Enqueue(ctx context.Context, job Job) error {
	select {
	case q.jobs <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *ChannelQueue) Dequeue(ctx context.Context) (Job, error) {
	select {
	case job := <-q.jobs:
		return job, nil
	case <-ctx.Done():
		return Job{}, ctx.Err()
	}
}
