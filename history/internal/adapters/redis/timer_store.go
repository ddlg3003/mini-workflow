package redis

import (
	"context"
	"time"

	"mini-workflow/history/internal/domain"

	goredis "github.com/redis/go-redis/v9"
)

const (
	activeSetKey  = "timers:active_set"
	wakeupChannel = "timers:wakeup_channel"
	wakeupMessage = "NEW_EARLIEST_TIMER"
)

var claimExpiredScript = goredis.NewScript(`
local ids = redis.call('ZRANGEBYSCORE', KEYS[1], 0, ARGV[1], 'LIMIT', 0, ARGV[2])
if #ids > 0 then
    redis.call('ZREM', KEYS[1], unpack(ids))
end
return ids
`)

type redisTimerStore struct {
	client *goredis.Client
}

func NewTimerStore(client *goredis.Client) *redisTimerStore {
	return &redisTimerStore{client: client}
}

func (s *redisTimerStore) ScheduleTimer(ctx context.Context, timerID string, fireTimeMs int64) error {
	if err := s.client.ZAdd(ctx, activeSetKey, goredis.Z{
		Score:  float64(fireTimeMs),
		Member: timerID,
	}).Err(); err != nil {
		return err
	}

	earliest, err := s.client.ZRangeWithScores(ctx, activeSetKey, 0, 0).Result()
	if err != nil {
		return err
	}
	if len(earliest) > 0 && earliest[0].Member.(string) == timerID {
		return s.client.Publish(ctx, wakeupChannel, wakeupMessage).Err()
	}
	return nil
}

func (s *redisTimerStore) ClaimExpired(ctx context.Context, nowMs int64, batchSize int) ([]string, error) {
	result, err := claimExpiredScript.Run(ctx, s.client, []string{activeSetKey}, nowMs, batchSize).StringSlice()
	if err != nil && err != goredis.Nil {
		return nil, err
	}
	return result, nil
}

func (s *redisTimerStore) Subscribe(ctx context.Context) (<-chan struct{}, func()) {
	sub := s.client.Subscribe(ctx, wakeupChannel)
	ch := make(chan struct{}, 1)

	go func() {
		defer sub.Close()
		msgCh := sub.Channel()
		for {
			select {
			case _, ok := <-msgCh:
				if !ok {
					return
				}
				select {
				case ch <- struct{}{}:
				default:
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, func() { sub.Close() }
}

func (s *redisTimerStore) RebuildFromDB(ctx context.Context, timers []domain.Timer) error {
	if len(timers) == 0 {
		return nil
	}
	pipe := s.client.Pipeline()
	for _, t := range timers {
		pipe.ZAdd(ctx, activeSetKey, goredis.Z{
			Score:  float64(t.FireTime.UnixMilli()),
			Member: t.TimerID.String(),
		})
	}
	_, err := pipe.Exec(ctx)
	return err
}

func NewRedisClient(addr string) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
}
