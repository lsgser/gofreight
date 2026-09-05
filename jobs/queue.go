package jobs

import "context"

// QueueBackend dispatches and processes background jobs.
type QueueBackend interface {
	Dispatch(job Job)
	DispatchFunc(fn func(ctx context.Context) error)
	Process(ctx context.Context) []error
	Pending() int
	Errors() []error
	Flush()
}

// Ensure in-memory Queue implements QueueBackend.
var _ QueueBackend = (*Queue)(nil)
