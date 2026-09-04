package jobs

import (
	"context"
	"sync"
	"time"
)

// Worker processes jobs asynchronously in background goroutines.
type Worker struct {
	queue   *Queue
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	started bool
	mu      sync.Mutex
}

// NewWorker creates a background worker for a queue.
func NewWorker(q *Queue) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{queue: q, ctx: ctx, cancel: cancel}
}

// Start begins processing jobs in the background.
func (w *Worker) Start(concurrency int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.started {
		return
	}
	if concurrency < 1 {
		concurrency = 1
	}
	w.started = true
	for i := 0; i < concurrency; i++ {
		w.wg.Add(1)
		go w.loop()
	}
}

// Stop gracefully stops the worker.
func (w *Worker) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *Worker) loop() {
	defer w.wg.Done()
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
		}

		w.queue.mu.Lock()
		if len(w.queue.pending) == 0 {
			w.queue.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		job := w.queue.pending[0]
		w.queue.pending = w.queue.pending[1:]
		w.queue.mu.Unlock()

		if err := job.Handle(w.ctx); err != nil {
			w.queue.mu.Lock()
			w.queue.errors = append(w.queue.errors, err)
			w.queue.mu.Unlock()
		}
	}
}

// DispatchAsync adds a job and ensures worker is running.
func DispatchAsync(q *Queue, w *Worker, job Job) {
	q.Dispatch(job)
	if w != nil {
		w.Start(2)
	}
}
