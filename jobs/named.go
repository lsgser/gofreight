package jobs

/*
|--------------------------------------------------------------------------
| Named
|--------------------------------------------------------------------------
|
| Implements Named as part of the jobs package in the Gofreight framework.
| Key symbols: NamedJob, NamedJobFunc, JobName, Handle.
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
| Symbols defined here include: NamedJob (exported type); NamedJobFunc
| (exported type).
| 
*/

import "context"

// NamedJob is a job that can be serialized to Redis by name (register with RegisterJob).
type NamedJob interface {
	Job
	JobName() string
}

// NamedJobFunc wraps a function with a stable job name for Redis workers.
type NamedJobFunc struct {
	Name string
	Fn   func(ctx context.Context) error
}

func (j NamedJobFunc) JobName() string { return j.Name }
func (j NamedJobFunc) Handle(ctx context.Context) error {
	if j.Fn == nil {
		return nil
	}
	return j.Fn(ctx)
}
