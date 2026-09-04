package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lsgser/gofreight/database"
)

// Checker verifies subsystem health.
type Checker struct {
	Checks map[string]func(context.Context) error
}

// New creates a health checker with a database check.
func New() *Checker {
	c := &Checker{Checks: make(map[string]func(context.Context) error)}
	c.Checks["database"] = func(ctx context.Context) error {
		return database.DB().PingContext(ctx)
	}
	return c
}

// Handler returns an HTTP handler for GET /health.
func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		status := "ok"
		checks := make(map[string]string)
		code := http.StatusOK

		for name, fn := range c.Checks {
			if err := fn(ctx); err != nil {
				checks[name] = err.Error()
				status = "degraded"
				code = http.StatusServiceUnavailable
			} else {
				checks[name] = "ok"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]any{
			"status": status,
			"checks": checks,
		})
	}
}

// Mount registers /health on a mux-compatible handler.
func Mount(mux interface{ HandleFunc(string, http.HandlerFunc) }) {
	c := New()
	mux.HandleFunc("/health", c.Handler())
}
