# Deployment

## Docker

```bash
docker compose up --build
```

The app listens on port 5000. PostgreSQL and Redis are included for local/staging stacks.

## Production checklist

1. Set `GOFREIGHT_ENV=production`
2. Run `gofreight key:generate` to set a unique `APP_KEY`
3. Configure database (`DB_*` in `.env` or managed PostgreSQL — see `.env.example`)
4. Run migrations: `gofreight migrate`
5. Precompile assets (automatic in `app.Run()` when production)
6. Mount health check: `GET /health` (automatic via `app.Run()`)
7. Configure integrations via environment variables (see [integrations.md](integrations.md))
8. Enable CORS if serving a separate frontend: `app.UseCORS("https://yourdomain.com")`
9. Enable rate limiting: `app.UseRateLimit(100, time.Minute)`
10. **Never** expose `/admin` in production (disabled by default)

## Environment variables

See [.env.example](../.env.example) for the full list.

## Health checks

```bash
curl http://localhost:5000/health
```

Returns JSON with database status:

```json
{"status":"ok","checks":{"database":"ok"}}
```

## Structured logging

In production, Gofreight uses JSON structured logging via `slog` for request logs.

## Scaling

- Run multiple app instances behind a load balancer
- Use Redis for shared cache (`REDIS_URL`)
- Use a managed job worker or call `app.StartJobs(4)` for background processing
