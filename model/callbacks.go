package model

import "context"

// CallbackType represents a lifecycle hook point.
type CallbackType string

const (
	BeforeValidate CallbackType = "before_validate"
	AfterValidate  CallbackType = "after_validate"
	BeforeSave     CallbackType = "before_save"
	AfterSave      CallbackType = "after_save"
	BeforeCreate   CallbackType = "before_create"
	AfterCreate    CallbackType = "after_create"
	BeforeUpdate   CallbackType = "before_update"
	AfterUpdate    CallbackType = "after_update"
	BeforeDelete   CallbackType = "before_delete"
	AfterDelete    CallbackType = "after_delete"
)

// Callback is a function invoked during the model lifecycle.
type Callback func(ctx context.Context, record any) error

// Callbacks holds registered lifecycle hooks.
type Callbacks struct {
	hooks map[CallbackType][]Callback
}

// NewCallbacks creates an empty callback registry.
func NewCallbacks() *Callbacks {
	return &Callbacks{hooks: make(map[CallbackType][]Callback)}
}

// Register adds a callback for the given lifecycle event.
func (c *Callbacks) Register(event CallbackType, fn Callback) {
	c.hooks[event] = append(c.hooks[event], fn)
}

// Run executes all callbacks for an event in registration order.
func (c *Callbacks) Run(ctx context.Context, event CallbackType, record any) error {
	for _, fn := range c.hooks[event] {
		if err := fn(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

// Callbackable models can register lifecycle hooks.
type Callbackable interface {
	Callbacks() *Callbacks
}

// RunCallbacks executes callbacks for a model if it implements Callbackable.
func RunCallbacks(ctx context.Context, record any, event CallbackType) error {
	if cb, ok := record.(Callbackable); ok {
		return cb.Callbacks().Run(ctx, event, record)
	}
	return nil
}
