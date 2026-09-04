package jobs

import (
	"context"
	"sync"
)

// Job is a unit of background work.
type Job interface {
	Handle(ctx context.Context) error
}

// JobFunc adapts a function to Job.
type JobFunc func(ctx context.Context) error

func (f JobFunc) Handle(ctx context.Context) error {
	return f(ctx)
}

// Queue dispatches jobs for background processing (like Laravel Queue / ActiveJob).
type Queue struct {
	mu      sync.Mutex
	pending []Job
	running bool
	errors  []error
}

// New creates a job queue.
func New() *Queue {
	return &Queue{}
}

// Dispatch adds a job to the queue.
func (q *Queue) Dispatch(job Job) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pending = append(q.pending, job)
}

// DispatchFunc dispatches a function as a job.
func (q *Queue) DispatchFunc(fn func(ctx context.Context) error) {
	q.Dispatch(JobFunc(fn))
}

// Process runs all pending jobs synchronously (useful in tests).
func (q *Queue) Process(ctx context.Context) []error {
	q.mu.Lock()
	jobs := q.pending
	q.pending = nil
	q.mu.Unlock()

	var errs []error
	for _, job := range jobs {
		if err := job.Handle(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	q.mu.Lock()
	q.errors = append(q.errors, errs...)
	q.mu.Unlock()
	return errs
}

// Pending returns the number of pending jobs.
func (q *Queue) Pending() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pending)
}

// Errors returns processing errors.
func (q *Queue) Errors() []error {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]error{}, q.errors...)
}

// Flush clears pending jobs and errors.
func (q *Queue) Flush() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pending = nil
	q.errors = nil
}
