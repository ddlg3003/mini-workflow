package ports

import "context"

type TaskQueue interface {
	Push(ctx context.Context, key string, payload []byte) error
	BlockPop(ctx context.Context, key string, timeoutSecs int) ([]byte, error)
}
