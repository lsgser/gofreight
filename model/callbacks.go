package model

/*
|--------------------------------------------------------------------------
| Callbacks
|--------------------------------------------------------------------------
|
| Implements Callbacks as part of the model package in the Gofreight
| framework. Key symbols: CallbackType, Callback, Callbacks, NewCallbacks,
| Register, Run.
| 
| The model package is the ORM layer: repositories, queries, associations,
| soft deletes, validation, serialization, collections, and pagination.
| 
| Models map to tables via struct tags; migrations define schema
| separately in db/migrate.
| 
| See docs/models.md, docs/orm.md, and docs/factories.md for
| Laravel-aligned patterns.
| 
| Symbols defined here include: CallbackType (exported type);
| BeforeValidate (exported value); AfterValidate (exported value);
| BeforeSave (exported value); AfterSave (exported value); BeforeCreate
| (exported value); AfterCreate (exported value); BeforeUpdate (exported
| value); AfterUpdate (exported value); BeforeDelete (exported value);
| AfterDelete (exported value).
| 
*/

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
