package jobs

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
