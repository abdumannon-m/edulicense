# Edu License Go platform

Go web app for the public Edu License site and the private `/admin` dashboard.

- Public routes: `/`, `/uz`, `/privacy`, `/uz/privacy`, `/verify/{slug}`, `/uz/verify/{slug}`
- Private routes: `/admin/*`, with only `/admin/login` public
- Stack: Go `net/http`, Chi, server-rendered templates, HTMX, Postgres, Goose migrations, bcrypt sessions, S3/R2 uploads, Telegram reminder job

## Local setup

Install Go 1.23+ and Postgres, then configure env:

```sh
cp .env.example .env
```

Required values:

- `DATABASE_URL`
- `SESSION_SECRET`
- `APP_BASE_URL`
- S3/R2 vars for document uploads
- Telegram vars for reminder notifications

Run migrations and create the first super admin:

```sh
go run ./cmd/server migrate
SEED_ADMIN_NAME="Owner" \
SEED_ADMIN_EMAIL="owner@edulicense.uz" \
SEED_ADMIN_PASSWORD="change-me" \
go run ./cmd/server seed-admin
```

Start the app:

```sh
go run ./cmd/server serve
```

Open `http://localhost:8080` and `http://localhost:8080/admin/login`.

## Operations

Send due Telegram reminders from a managed scheduler or cron:

```sh
go run ./cmd/server reminders
```

The job scans reminders due in `APP_TIMEZONE`, posts to `TELEGRAM_OPERATIONS_CHAT_ID`, and marks each sent.

## Deployment

Use the included `Dockerfile` on a managed PaaS with a managed Postgres database. Typical release steps:

```sh
./edu-license migrate
./edu-license seed-admin   # first deploy only
./edu-license serve
```

Set `COOKIE_SECURE=true` in production. `/robots.txt` disallows `/admin`, but security depends on session auth and role checks.

### Vercel

The repository also includes `api/app.go` and `vercel.json` so Vercel can route `/admin` and `/admin/*` into a Go Function while the existing public Astro pages continue to serve normally. Configure these Vercel environment variables before redeploying:

- `DATABASE_URL`
- `SESSION_SECRET`
- `APP_BASE_URL=https://edulicense.uz`
- `COOKIE_SECURE=true`
- S3/R2 variables for uploads
- Telegram variables for reminders

If those variables are missing, `/admin/login` will stop returning 404 but will show a runtime configuration error until the environment is completed and migrations are run.

## Checks

```sh
go test ./...
go build ./cmd/server
```

Postgres integration tests are skipped unless `TEST_DATABASE_URL` is set.

The previous Astro source is still present as historical reference, but the Go app is the primary runtime.
