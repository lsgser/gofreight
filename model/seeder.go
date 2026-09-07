package model

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
