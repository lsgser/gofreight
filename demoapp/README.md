# demoapp

A Gofreight web application (SQLite by default).

## Run

```bash
go mod tidy
gofreight db:create
gofreight migrate
gofreight serve          # http://localhost:5000
```

Admin (development): http://localhost:5000/admin

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

Full structure: https://github.com/lsgser/gofreight/blob/main/docs/project-structure.md
