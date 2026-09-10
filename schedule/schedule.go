package schedule

/*
|--------------------------------------------------------------------------
| Schedule
|--------------------------------------------------------------------------
|
| Implements Schedule as part of the schedule package in the Gofreight
| framework. Key symbols: Task, Scheduler, New, Every, Daily, Tasks.
| 
| Schedule defines cron-like tasks invoked by gofreight schedule:run,
| suitable for system crontab or Kubernetes CronJob.
| 
| Register closures or command strings in bootstrap/schedule.go.
| 
| Symbols defined here include: Task (exported type); Scheduler (exported
| type); New (New creates a scheduler.); Every (Every registers a task at
| an interval.); Daily (Daily registers a task at HH:MM each day.); Tasks
| (Tasks returns registered tasks.); RunDue (RunDue executes tasks that
| are due now.).
| 
*/

import (
	"context"
	"sync"
	"time"
)

// Task runs on a schedule.
type Task struct {
	Name     string
	Interval time.Duration
	At       string // HH:MM daily (optional, overrides Interval when set)
	Run      func(ctx context.Context) error
}

// Scheduler manages recurring tasks.
type Scheduler struct {
	mu    sync.Mutex
	tasks []Task
	last  map[string]time.Time
}

// New creates a scheduler.
func New() *Scheduler {
	return &Scheduler{last: make(map[string]time.Time)}
}

// Every registers a task at an interval.
func (s *Scheduler) Every(interval time.Duration, name string, fn func(ctx context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, Task{Name: name, Interval: interval, Run: fn})
}

// Daily registers a task at HH:MM each day.
func (s *Scheduler) Daily(at, name string, fn func(ctx context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, Task{Name: name, At: at, Run: fn})
}

// Tasks returns registered tasks.
func (s *Scheduler) Tasks() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Task{}, s.tasks...)
}

// RunDue executes tasks that are due now.
func (s *Scheduler) RunDue(ctx context.Context) []error {
	s.mu.Lock()
	tasks := append([]Task{}, s.tasks...)
	s.mu.Unlock()

	var errs []error
	now := time.Now()
	for _, t := range tasks {
		if !s.isDue(t, now) {
			continue
		}
		if t.Run != nil {
			if err := t.Run(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		s.mu.Lock()
		s.last[t.Name] = now
		s.mu.Unlock()
	}
	return errs
}

func (s *Scheduler) isDue(t Task, now time.Time) bool {
	s.mu.Lock()
	last, ok := s.last[t.Name]
	s.mu.Unlock()

	if t.At != "" {
		parsed, err := time.Parse("15:04", t.At)
		if err != nil {
			return false
		}
		scheduled := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
		if now.Before(scheduled) {
			return false
		}
		if ok && last.After(scheduled) {
			return false
		}
		return true
	}

	if t.Interval <= 0 {
		return false
	}
	if !ok {
		return true
	}
	return now.Sub(last) >= t.Interval
}
