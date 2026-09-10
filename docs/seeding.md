<!--
| --------------------------------------------------------------------------
| Seeding
| --------------------------------------------------------------------------
|
| Official Gofreight documentation consumed by the docs site and linked
| from the CLI welcome page.
|
| Keep examples aligned with the current CLI and version; sync copies to
| gofreight-web when publishing.
|
| Covers concepts, tutorials, and reference material for this topic.
|

-->

<!--
| --------------------------------------------------------------------------
| Database Seeding
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Database Seeding

Gofreight seeding follows the [Laravel seeding](https://laravel.com/docs/seeding) pattern: Go seeder classes in `db/seeders/`, a root `DatabaseSeeder`, and `gofreight db:seed`.

## Quick start

```bash
gofreight make:seeder UserSeeder
gofreight make:seeder DatabaseSeeder   # root seeder (included in new apps)
gofreight db:seed
gofreight db:seed --class=UserSeeder
gofreight migrate:fresh --seed
```

New apps include `db/seeders/database_seeder.go` and `cmd/seed/main.go`. By default, `gofreight db:seed` runs `DatabaseSeeder`.

## Writing seeders

```go
package seeders

import (
    "context"

    "demoapp/app/models"
    "github.com/lsgser/gofreight/model"
)

type UserSeeder struct{}

func NewUserSeeder() *UserSeeder { return &UserSeeder{} }

func (s *UserSeeder) Run(ctx context.Context) error {
    return models.Users.Create(ctx, &models.User{
        Name:  "Admin",
        Email: "admin@example.com",
    })
}
```

## Calling other seeders (Laravel `$this->call()`)

```go
type DatabaseSeeder struct {
    model.Seeder
}

func NewDatabaseSeeder() *DatabaseSeeder {
    return &DatabaseSeeder{Seeder: model.NewSeeder()}
}

func (s *DatabaseSeeder) Run(ctx context.Context) error {
    s.SetContext(ctx)
    return s.Call(
        NewUserSeeder(),
        NewPostSeeder(),
    )
}
```

`model.Seeder.Call()` runs seeders in order — the Go equivalent of Laravel's `$this->call([UserSeeder::class, PostSeeder::class])`.

## SQL seeds

Place files in `db/seeds/` for raw SQL. `gofreight db:seed` runs SQL seeds first, then Go seeders.

## Factories in seeders

Use test factories or repositories directly:

```go
func (s *UserSeeder) Run(ctx context.Context) error {
    for i := 0; i < 50; i++ {
        if err := models.Users.Create(ctx, &models.User{Name: "User"}); err != nil {
            return err
        }
    }
    return nil
}
```

See **[Database](database.md)** for migration + seed workflow and **[Testing](testing.md)** for `gftest.Seed()`.
