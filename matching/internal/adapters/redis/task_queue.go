package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type taskQueue struct {
	client *redis.Client
}

func NewTaskQueue(client *redis.Client) *taskQueue {
	return &taskQueue{client: client}
}

func (q *taskQueue) Push(ctx context.Context, key string, payload []byte) error {
	return q.client.LPush(ctx, key, payload).Err()
}

// BlockPop blocks up to timeoutSecs seconds waiting for a task.
// Returns (nil, nil) when the timeout expires with no task.
func (q *taskQueue) BlockPop(ctx context.Context, key string, timeoutSecs int) ([]byte, error) {
	result, err := q.client.BRPop(ctx, time.Duration(timeoutSecs)*time.Second, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	// BRPop returns [key, value]
	if len(result) < 2 {
		return nil, nil
	}
	return []byte(result[1]), nil
}
