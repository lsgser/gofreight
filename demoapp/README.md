# demoapp

A Gofreight web application (SQLite by default).

## Run

```bash
go mod tidy
gofreight key:generate
gofreight db:create
gofreight migrate
gofreight serve          # http://localhost:5000
```

Admin (development): http://localhost:5000/admin

## Routes

Web routes live in `routes/web.go`. API routes live in `routes/api.go` and are registered under `/api/v1` via route groups in `routes/register.go`. See [Routing](../docs/routing.md).

## Layout

```
bootstrap/app.go   Application wiring
routes/            Web and API routes
app/controllers/   HTTP handlers
app/models/        Database models
app/views/         GFT templates (.gft)
db/migrate/        SQL migrations
public/            Static assets
tests/             HTTP tests
```

Documentation: https://lsgser.github.io/gofreight-web/
