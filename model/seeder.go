package model

/*
|--------------------------------------------------------------------------
| Seeder
|--------------------------------------------------------------------------
|
| Implements Seeder as part of the model package in the Gofreight
| framework. Key symbols: SeederRunner, Seeder, NewSeeder, SetContext,
| Context, Call.
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
| Symbols defined here include: SeederRunner (exported type); Seeder
| (exported type); NewSeeder (NewSeeder creates a base seeder. Set context
| with SetContext before Call.); SetContext (SetContext sets the context
| used by Call.); Context (Context returns the seeder context.); Call
| (Call runs one or more seeders in order (Laravel
| $this->call([UserSeeder::class, ...])).).
| 
*/

import "context"

// SeederRunner is implemented by seeder classes (Laravel Database\Seeder).
type SeederRunner interface {
	Run(ctx context.Context) error
}

// Seeder provides Laravel-style seeder orchestration ($this->call([...])).
type Seeder struct {
	ctx context.Context
}

// NewSeeder creates a base seeder. Set context with SetContext before Call.
func NewSeeder() Seeder {
	return Seeder{ctx: context.Background()}
}

// SetContext sets the context used by Call.
func (s *Seeder) SetContext(ctx context.Context) {
	if ctx != nil {
		s.ctx = ctx
	}
}

// Context returns the seeder context.
func (s *Seeder) Context() context.Context {
	if s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

// Call runs one or more seeders in order (Laravel $this->call([UserSeeder::class, ...])).
func (s *Seeder) Call(seeders ...SeederRunner) error {
	for _, seeder := range seeders {
		if seeder == nil {
			continue
		}
		if err := seeder.Run(s.Context()); err != nil {
			return err
		}
	}
	return nil
}
