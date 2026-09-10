package jobs

/*
|--------------------------------------------------------------------------
| Queue
|--------------------------------------------------------------------------
|
| Implements Queue as part of the jobs package in the Gofreight framework.
| Key symbols: QueueBackend.
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
| Symbols defined here include: QueueBackend (exported type).
| 
*/

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
