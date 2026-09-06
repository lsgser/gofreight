package jobs

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisQueue persists jobs in Redis for multi-process workers.
type RedisQueue struct {
	client *redis.Client
	key    string
	mu     sync.Mutex
	errors []error
}

// NewRedisQueue creates a Redis-backed job queue.
func NewRedisQueue(url, key string) (*RedisQueue, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	if key == "" {
		key = "gofreight:jobs"
	}
	return &RedisQueue{
		client: redis.NewClient(opts),
		key:    key,
	}, nil
}

// DispatchFuncJob wraps a named function job for Redis serialization boundary.
type DispatchFuncJob struct {
	Name string
	Fn   func(ctx context.Context) error
}

func (j DispatchFuncJob) Handle(ctx context.Context) error {
	if j.Fn == nil {
		return nil
	}
	return j.Fn(ctx)
}

var registry = struct {
	mu    sync.RWMutex
	named map[string]func(ctx context.Context) error
}{named: make(map[string]func(ctx context.Context) error)}

// RegisterJob registers a named job handler for Redis queue workers.
func RegisterJob(name string, fn func(ctx context.Context) error) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.named[name] = fn
}

func (q *RedisQueue) enqueue(payload JobPayload) {
	if payload.MaxAttempts < 1 {
		payload.MaxAttempts = defaultMaxAttempts
	}
	if payload.ID == "" {
		payload.ID = randomID()
	}
	raw, _ := json.Marshal(payload)
	q.client.RPush(context.Background(), q.key, raw)
}

func (q *RedisQueue) Dispatch(job Job) {
	if named, ok := job.(DispatchFuncJob); ok && named.Name != "" {
		registry.mu.Lock()
		registry.named[named.Name] = named.Fn
		registry.mu.Unlock()
		q.enqueue(JobPayload{Name: named.Name, MaxAttempts: defaultMaxAttempts})
		return
	}
	_ = job.Handle(context.Background())
}

func (q *RedisQueue) DispatchFunc(fn func(ctx context.Context) error) {
	q.Dispatch(JobFunc(fn))
}

func (q *RedisQueue) Process(ctx context.Context) []error {
	var errs []error
	for {
		err := q.ProcessOne(ctx)
		if err == redis.Nil {
			break
		}
		if err != nil {
			errs = append(errs, err)
		}
	}
	q.mu.Lock()
	q.errors = append(q.errors, errs...)
	q.mu.Unlock()
	return errs
}

func (q *RedisQueue) Pending() int {
	n, _ := q.client.LLen(context.Background(), q.key).Result()
	return int(n)
}

func (q *RedisQueue) Errors() []error {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]error{}, q.errors...)
}

func (q *RedisQueue) Flush() {
	q.client.Del(context.Background(), q.key)
	q.client.Del(context.Background(), q.failedKey())
	q.mu.Lock()
	q.errors = nil
	q.mu.Unlock()
}

// StartWorker polls Redis and processes jobs until ctx is cancelled.
func (q *RedisQueue) StartWorker(ctx context.Context, concurrency int) {
	if concurrency < 1 {
		concurrency = 1
	}
	for i := 0; i < concurrency; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					_ = q.ProcessOne(ctx)
					time.Sleep(200 * time.Millisecond)
				}
			}
		}()
	}
}
