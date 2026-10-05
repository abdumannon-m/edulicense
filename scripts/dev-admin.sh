#!/usr/bin/env bash
# Runs the Go admin locally in Docker against a local Postgres. A fresh
# database is seeded with a test admin and the SAT 1111 finance history;
# scripts/dev-db-from-prod.sh replaces it with a copy of production. Rebuilds when .go or template files change;
# CSS under web/static is served from disk, so a browser refresh shows it.
#
#   scripts/dev-admin.sh            # http://localhost:8090/admin
#
# The login page signs you in automatically as the first active super admin
# (DEV_AUTO_LOGIN, honoured only for a localhost APP_BASE_URL).
set -euo pipefail
cd "$(dirname "$0")/.."

PORT="${PORT:-8090}"
NET=edu-dev-net
PG=edu-dev-pg
APP=edu-dev-app

cleanup() { docker rm -f "$APP" >/dev/null 2>&1 || true; }
trap cleanup EXIT INT TERM

docker network create "$NET" >/dev/null 2>&1 || true
if ! docker ps --format '{{.Names}}' | grep -qx "$PG"; then
	docker rm -f "$PG" >/dev/null 2>&1 || true
	docker run -d --name "$PG" --network "$NET" \
		-e POSTGRES_PASSWORD=local -e POSTGRES_DB=edu_license \
		postgres:17-alpine >/dev/null
fi
until docker exec "$PG" pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done

# Seed only an empty database, never a production copy. The auto-login
# signs in as the seed admin, or as the first active super admin of a copy.
SEED=1
DEV_LOGIN_EMAIL=owner@edulicense.uz
EXISTING=$(docker exec "$PG" psql -U postgres -d edu_license -At -c \
	"SELECT email FROM users WHERE role = 'super_admin' AND active ORDER BY created_at LIMIT 1" 2>/dev/null || true)
if [ -n "$EXISTING" ]; then
	SEED=0
	DEV_LOGIN_EMAIL=$EXISTING
fi

docker rm -f "$APP" >/dev/null 2>&1 || true
exec docker run --rm --name "$APP" --network "$NET" -p "$PORT:8080" \
	-v "$PWD:/src" -w /src \
	-v edu-dev-gomod:/go/pkg/mod -v edu-dev-gobuild:/root/.cache/go-build \
	-e DATABASE_URL="postgres://postgres:local@$PG:5432/edu_license?sslmode=disable" \
	-e SESSION_SECRET=local-dev-secret-0123456789abcdef0123 \
	-e COOKIE_SECURE=false -e APP_BASE_URL="http://localhost:$PORT" \
	-e APP_TIMEZONE=Asia/Tashkent -e ADDR=:8080 -e DEV_AUTO_LOGIN=1 -e SEED="$SEED" -e DEV_LOGIN_EMAIL="$DEV_LOGIN_EMAIL" \
	-e SEED_ADMIN_NAME="Edu License Owner" \
	-e SEED_ADMIN_EMAIL="owner@edulicense.uz" \
	-e SEED_ADMIN_PASSWORD="change-this-password" \
	golang:1.26-alpine sh -c '
		set -u
		stamp() { find cmd pkg -name "*.go" -o -name "*.html" | xargs stat -c %Y | sort -n | tail -1; }
		go build -o /tmp/edu ./cmd/server || exit 1
		/tmp/edu migrate
		if [ "$SEED" = 1 ]; then
			/tmp/edu seed-admin 2>/dev/null || true
			/tmp/edu finance-import 2>/dev/null || true
		fi
		while true; do
			last=$(stamp)
			/tmp/edu serve & pid=$!
			while [ "$(stamp)" = "$last" ]; do sleep 1; done
			echo "change detected, rebuilding"
			until go build -o /tmp/edu.next ./cmd/server; do sleep 2; [ "$(stamp)" != "$last" ] && last=$(stamp); done
			kill $pid; wait $pid 2>/dev/null; mv /tmp/edu.next /tmp/edu
		done'
