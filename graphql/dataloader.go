package graphql

import (
	"context"
	"sync"

	"github.com/graph-gophers/dataloader/v7"
)

type loadersKey struct{}

// LoaderFactory creates a request-scoped dataloader.
type LoaderFactory func() any

// LoaderRegistry holds named dataloaders for a single GraphQL request.
type LoaderRegistry struct {
	mu      sync.Mutex
	factory map[string]LoaderFactory
	cache   map[string]any
}

// NewLoaderRegistry creates an empty loader registry.
func NewLoaderRegistry() *LoaderRegistry {
	return &LoaderRegistry{
		factory: make(map[string]LoaderFactory),
		cache:   make(map[string]any),
	}
}

// Register adds a named loader factory invoked once per request.
func (r *LoaderRegistry) Register(name string, factory LoaderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factory[name] = factory
}

// Get returns a request-scoped loader, creating it on first access.
func (r *LoaderRegistry) Get(name string) (any, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if loader, ok := r.cache[name]; ok {
		return loader, true
	}
	factory, ok := r.factory[name]
	if !ok {
		return nil, false
	}
	loader := factory()
	r.cache[name] = loader
	return loader, true
}

// WithLoaders attaches a loader registry to context.
func WithLoaders(ctx context.Context, reg *LoaderRegistry) context.Context {
	return context.WithValue(ctx, loadersKey{}, reg)
}

// LoadersFromContext returns the loader registry for this request.
func LoadersFromContext(ctx context.Context) (*LoaderRegistry, bool) {
	reg, ok := ctx.Value(loadersKey{}).(*LoaderRegistry)
	return reg, ok
}

// LoaderFromContext returns a typed dataloader from the request context.
func LoaderFromContext[K comparable, V any](ctx context.Context, name string) (*dataloader.Loader[K, V], bool) {
	reg, ok := LoadersFromContext(ctx)
	if !ok {
		return nil, false
	}
	raw, ok := reg.Get(name)
	if !ok {
		return nil, false
	}
	loader, ok := raw.(*dataloader.Loader[K, V])
	return loader, ok
}

// NewLoader is a convenience wrapper around graph-gophers/dataloader.
func NewLoader[K comparable, V any](
	batchFn dataloader.BatchFunc[K, V],
	opts ...dataloader.Option[K, V],
) *dataloader.Loader[K, V] {
	return dataloader.NewBatchedLoader(batchFn, opts...)
}
