# Deployment

Deploy Gofreight applications as a single compiled binary with optional Redis, PostgreSQL, and reverse proxy.

---

## Build for production

```bash
# From app root
go build -o bin/myapp .

# Cross-compile (Linux from macOS)
GOOS=linux GOARCH=amd64 go build -o bin/myapp .
```

The binary includes your app code and links against Gofreight — no separate runtime needed.

---

## Environment

Production `.env`:

```env
GOFREIGHT_ENV=production
APP_DEBUG=false
APP_URL=https://myapp.example.com
APP_KEY=your-generated-key

HOST=0.0.0.0
PORT=8080

DB_CONNECTION=pgsql
DATABASE_URL=postgres://user:pass@db:5432/myapp?sslmode=require

SESSION_DRIVER=redis
CACHE_STORE=redis
QUEUE_CONNECTION=redis
REDIS_URL=redis://redis:6379

MAIL_DRIVER=smtp
MAIL_HOST=smtp.example.com
MAIL_PORT=587
MAIL_USERNAME=...
MAIL_PASSWORD=...
MAIL_FROM=noreply@example.com
```

Generate a key:

```bash
gofreight key:generate
```

---

## Database migrations

Run before starting the app:

```bash
gofreight migrate --force
```

In Docker entrypoint or deploy script:

```bash
#!/bin/sh
set -e
gofreight migrate --force
exec ./bin/myapp
```

---

## Process management

### systemd

```ini
[Unit]
Description=Gofreight App
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/myapp
EnvironmentFile=/var/www/myapp/.env
ExecStart=/var/www/myapp/bin/myapp
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Background jobs

Separate worker service:

```ini
[Service]
ExecStart=/usr/local/bin/gofreight queue:work
WorkingDirectory=/var/www/myapp
EnvironmentFile=/var/www/myapp/.env
```

### Scheduler

```cron
* * * * * cd /var/www/myapp && gofreight schedule:run >> storage/logs/scheduler.log 2>&1
```

---

## Reverse proxy (nginx)

```nginx
server {
    listen 443 ssl http2;
    server_name myapp.example.com;

    ssl_certificate     /etc/ssl/certs/myapp.crt;
    ssl_certificate_key /etc/ssl/private/myapp.key;

    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    location /socket {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

---

## Docker

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /myapp .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /myapp .
COPY .env.example .env
COPY app/views app/views
COPY public public
COPY db db
COPY config config
EXPOSE 8080
CMD ["./myapp"]
```

`docker-compose.yml` with Postgres + Redis:

```yaml
services:
  app:
    build: .
    ports: ["8080:8080"]
    env_file: .env
    depends_on: [db, redis]

  worker:
    build: .
    command: gofreight queue:work
    env_file: .env
    depends_on: [redis, db]

  db:
    image: postgres:16
    environment:
      POSTGRES_DB: myapp
      POSTGRES_PASSWORD: secret

  redis:
    image: redis:7-alpine
```

---

## Health checks

Gofreight mounts `GET /health` automatically:

```bash
curl https://myapp.example.com/health
```

Response:

```json
{
  "status": "ok",
  "checks": {
    "database": "ok"
  }
}
```

Degraded (503) when database is unreachable.

### Custom health checks

```go
checker := health.New()
checker.Checks["redis"] = func(ctx context.Context) error {
    return redisClient.Ping(ctx).Err()
}
app.Router.Get("/health", checker.Handler())
```

---

## Asset compilation

Production assets:

```bash
# Vite / frontend build
npm run build

# Load manifest
app.LoadAssetManifest("public/manifest.json")
```

Route cache (optional):

```bash
gofreight route:cache
```

---

## Scaling

| Component | Scale approach |
|-----------|----------------|
| HTTP | Multiple app instances behind load balancer |
| Sessions | `SESSION_DRIVER=redis` |
| Cache | `CACHE_STORE=redis` |
| WebSockets | `UseRedisBroadcast(REDIS_URL)` |
| Jobs | Multiple `queue:work` processes |

---

## WireDefaults for production

Minimal production bootstrap:

```go
app := application.New()
app.WireDefaults()
app.UseExceptionHandler()
app.UseCSRF()
app.UseRateLimit(200, time.Minute)
_ = app.LoadAssetManifest("public/manifest.json")
```

See [Application wiring](application-wiring.md).

---

## Pre-deploy checklist

- [ ] `GOFREIGHT_ENV=production`, `APP_DEBUG=false`
- [ ] `APP_KEY` generated and stored securely
- [ ] PostgreSQL/MySQL configured
- [ ] Redis for sessions, cache, queue (if multi-process)
- [ ] Migrations run (`gofreight migrate --force`)
- [ ] HTTPS configured
- [ ] Health check monitored
- [ ] Worker and scheduler processes running
- [ ] Logs aggregated (stdout → journald / CloudWatch / etc.)
- [ ] Backups configured for database

---

## Related

- [Configuration](configuration.md)
- [Security](security.md)
- [Application wiring](application-wiring.md)
- [Jobs & Queues](jobs.md)
- [Scheduling](scheduling.md)
