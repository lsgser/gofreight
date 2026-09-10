package jobs

/*
|--------------------------------------------------------------------------
| Retry
|--------------------------------------------------------------------------
|
| Implements Retry as part of the jobs package in the Gofreight framework.
| Key symbols: JobPayload, FailedJob, DispatchNamed, ProcessOne, Failed,
| RetryFailed.
| 
| The jobs package defines queue interfaces, in-memory and Redis drivers,
| retries, and named job registration.
| 
| Dispatch struct jobs or func workers from controllers; process with
| gofreight queue:work.
| 
| Queued mail and long-running tasks should use this layer instead of
| blocking HTTP handlers.
| 
| Symbols defined here include: JobPayload (exported type); FailedJob
| (exported type); DispatchNamed (DispatchNamed enqueues a registered job
| with retry support.); ProcessOne (ProcessOne runs a single job with
| retries and failed-job recording.); Failed (Failed returns failed jobs
| from Redis.); RetryFailed (RetryFailed re-queues a failed job by id.);
| FlushFailed (FlushFailed removes all failed jobs.); ForgetFailed
| (ForgetFailed removes a single failed job by payload id.).
| 
*/

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultMaxAttempts = 3

// JobPayload is a serializable queued job with retry metadata.
type JobPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Attempts    int    `json:"attempts"`
	MaxAttempts int    `json:"max_attempts"`
	RunAt       int64  `json:"run_at,omitempty"`
	LastError   string `json:"last_error,omitempty"`
}

// FailedJob is a job that exceeded max attempts.
type FailedJob struct {
	Payload   JobPayload `json:"payload"`
	FailedAt  int64      `json:"failed_at"`
	Exception string     `json:"exception"`
}

func (q *RedisQueue) queueKey() string  { return q.key }
func (q *RedisQueue) failedKey() string { return q.key + ":failed" }
func (q *RedisQueue) delayedKey() string { return q.key + ":delayed" }

// DispatchNamed enqueues a registered job with retry support.
func (q *RedisQueue) DispatchNamed(name string, maxAttempts int) {
	if maxAttempts < 1 {
		maxAttempts = defaultMaxAttempts
	}
	payload := JobPayload{
		ID:          randomID(),
		Name:        name,
		MaxAttempts: maxAttempts,
	}
	raw, _ := json.Marshal(payload)
	q.client.RPush(context.Background(), q.queueKey(), raw)
}

// ProcessOne runs a single job with retries and failed-job recording.
func (q *RedisQueue) ProcessOne(ctx context.Context) error {
	raw, err := q.client.LPop(ctx, q.queueKey()).Bytes()
	if err == redis.Nil {
		return redis.Nil
	}
	if err != nil {
		return err
	}
	var payload JobPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	if payload.MaxAttempts < 1 {
		payload.MaxAttempts = defaultMaxAttempts
	}

	registry.mu.RLock()
	fn := registry.named[payload.Name]
	registry.mu.RUnlock()
	if fn == nil {
		return nil
	}

	if err := fn(ctx); err != nil {
		payload.Attempts++
		payload.LastError = err.Error()
		if payload.Attempts >= payload.MaxAttempts {
			q.recordFailed(payload, err)
			return err
		}
		delay := backoff(payload.Attempts)
		payload.RunAt = time.Now().Add(delay).Unix()
		retryRaw, _ := json.Marshal(payload)
		q.client.RPush(ctx, q.queueKey(), retryRaw)
		return err
	}
	return nil
}

func (q *RedisQueue) recordFailed(payload JobPayload, err error) {
	fj := FailedJob{
		Payload:   payload,
		FailedAt:  time.Now().Unix(),
		Exception: err.Error(),
	}
	raw, _ := json.Marshal(fj)
	q.client.LPush(context.Background(), q.failedKey(), raw)
}

func backoff(attempt int) time.Duration {
	sec := 1 << attempt
	if sec > 300 {
		sec = 300
	}
	return time.Duration(sec) * time.Second
}

// Failed returns failed jobs from Redis.
func (q *RedisQueue) Failed(ctx context.Context) ([]FailedJob, error) {
	raws, err := q.client.LRange(ctx, q.failedKey(), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	var out []FailedJob
	for _, raw := range raws {
		var fj FailedJob
		if json.Unmarshal([]byte(raw), &fj) == nil {
			out = append(out, fj)
		}
	}
	return out, nil
}

// RetryFailed re-queues a failed job by id.
func (q *RedisQueue) RetryFailed(ctx context.Context, id string) bool {
	raws, _ := q.client.LRange(ctx, q.failedKey(), 0, -1).Result()
	for i, raw := range raws {
		var fj FailedJob
		if json.Unmarshal([]byte(raw), &fj) != nil || fj.Payload.ID != id {
			continue
		}
		fj.Payload.Attempts = 0
		fj.Payload.LastError = ""
		requeue, _ := json.Marshal(fj.Payload)
		q.client.RPush(ctx, q.queueKey(), requeue)
		q.client.LRem(ctx, q.failedKey(), 1, raw)
		_ = i
		return true
	}
	return false
}

// FlushFailed removes all failed jobs.
func (q *RedisQueue) FlushFailed(ctx context.Context) error {
	return q.client.Del(ctx, q.failedKey()).Err()
}

// ForgetFailed removes a single failed job by payload id.
func (q *RedisQueue) ForgetFailed(ctx context.Context, id string) bool {
	raws, _ := q.client.LRange(ctx, q.failedKey(), 0, -1).Result()
	for _, raw := range raws {
		var fj FailedJob
		if json.Unmarshal([]byte(raw), &fj) != nil || fj.Payload.ID != id {
			continue
		}
		q.client.LRem(ctx, q.failedKey(), 1, raw)
		return true
	}
	return false
}

// Clear removes all pending jobs from the queue.
func (q *RedisQueue) Clear(ctx context.Context) error {
	return q.client.Del(ctx, q.queueKey()).Err()
}

func randomID() string {
	return time.Now().Format("20060102150405.000000000")
}
