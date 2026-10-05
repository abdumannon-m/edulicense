#!/usr/bin/env bash
# Copies the production database into the local dev Postgres (edu-dev-pg) so
# the dev server shows real data. Production is only read (pg_dump); every
# change you make afterwards stays on your machine.
#
#   scripts/dev-db-from-prod.sh     # then restart scripts/dev-admin.sh
#
# Reads DATABASE_URL_UNPOOLED (or DATABASE_URL) from .env.production.local.
set -euo pipefail
cd "$(dirname "$0")/.."

PG=edu-dev-pg
NET=edu-dev-net
IMAGE=postgres:17-alpine

SRC=$(set -a; . ./.env.production.local >/dev/null 2>&1; echo "${DATABASE_URL_UNPOOLED:-${DATABASE_URL:-}}")
if [ -z "$SRC" ]; then
	echo "No DATABASE_URL_UNPOOLED or DATABASE_URL in .env.production.local" >&2
	exit 1
fi

docker network create "$NET" >/dev/null 2>&1 || true
if ! docker ps --format '{{.Names}}' | grep -qx "$PG"; then
	docker rm -f "$PG" >/dev/null 2>&1 || true
	docker run -d --name "$PG" --network "$NET" \
		-e POSTGRES_PASSWORD=local -e POSTGRES_DB=edu_license "$IMAGE" >/dev/null
fi
until docker exec "$PG" pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done

DUMP=$(mktemp)
trap 'rm -f "$DUMP"' EXIT
echo "Dumping production (read-only)…"
docker run --rm -e SRC="$SRC" "$IMAGE" sh -c 'pg_dump --no-owner --no-privileges "$SRC"' >"$DUMP"

echo "Replacing the local edu_license database…"
docker exec "$PG" psql -U postgres -q -c "DROP DATABASE IF EXISTS edu_license WITH (FORCE)" -c "CREATE DATABASE edu_license"
docker exec -i "$PG" psql -U postgres -q -v ON_ERROR_STOP=1 -d edu_license <"$DUMP" >/dev/null

docker exec "$PG" psql -U postgres -d edu_license -At -c "
	SELECT 'users ' || count(*) FROM users
	UNION ALL SELECT 'applications ' || count(*) FROM test_center_applications
	UNION ALL SELECT 'deals ' || count(*) FROM sales_deals
	UNION ALL SELECT 'finance transactions ' || count(*) FROM fin_transactions"
echo "Done. Restart the dev server to sign in as the first super admin."
