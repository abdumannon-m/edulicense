#!/usr/bin/env bash
#
# Copy the whole database from one Neon project to another.
#
# Neon cannot move an existing project between regions, so switching region means
# creating a fresh project in the target region and copying the data across.
# See docs/region-migration.md for the full runbook.
#
# Usage:
#   SOURCE_URL='postgresql://...ap-southeast-1...' \
#   TARGET_URL='postgresql://...eu-central-1...' \
#   ./scripts/migrate-region.sh
#
# Both URLs must be the DIRECT endpoint, not the -pooler one. The pooled endpoint
# runs through pgbouncer, which does not support the session-level operations
# pg_dump and pg_restore rely on.
#
# Requires Docker. The postgres:17 image supplies pg_dump/psql so the local
# machine needs no Postgres client, and a newer client can always dump an older
# server.

set -euo pipefail

PG_IMAGE="postgres:17"
TABLES=(users sessions certificates sales_deals test_center_applications application_documents reminders activity_logs)

: "${SOURCE_URL:?set SOURCE_URL to the direct connection string of the current database}"
: "${TARGET_URL:?set TARGET_URL to the direct connection string of the new database}"

if [[ "$SOURCE_URL" == *"-pooler."* ]]; then
  echo "error: SOURCE_URL is the pooled endpoint. Use the direct one (drop '-pooler' from the host)." >&2
  exit 1
fi
if [[ "$TARGET_URL" == *"-pooler."* ]]; then
  echo "error: TARGET_URL is the pooled endpoint. Use the direct one (drop '-pooler' from the host)." >&2
  exit 1
fi
if [[ "$SOURCE_URL" == "$TARGET_URL" ]]; then
  echo "error: SOURCE_URL and TARGET_URL are the same database." >&2
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "error: docker is required but was not found on PATH." >&2
  exit 1
fi

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT
dump_file="$workdir/dump.sql"

# Run a psql query against a database and print the bare result.
# The URL and query go in as environment variables rather than argv so the
# credentials stay out of the container's command line. docker run does not
# start a shell, so an explicit sh -c is needed to expand them.
psql_value() {
  local url="$1" query="$2"
  docker run --rm -e PGURL="$url" -e PGQUERY="$query" "$PG_IMAGE" \
    sh -c 'psql "$PGURL" -tAX -c "$PGQUERY"'
}

echo "==> Checking connectivity"
source_version="$(psql_value "$SOURCE_URL" 'SHOW server_version')"
target_version="$(psql_value "$TARGET_URL" 'SHOW server_version')"
echo "    source server: Postgres $source_version"
echo "    target server: Postgres $target_version"

echo "==> Verifying the target is empty"
existing="$(psql_value "$TARGET_URL" \
  "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'")"
if [[ "$existing" != "0" ]]; then
  echo "error: target already has $existing table(s) in the public schema." >&2
  echo "       Refusing to overwrite. Create a fresh Neon project, or drop the schema first." >&2
  exit 1
fi

echo "==> Dumping source"
# --no-owner / --no-privileges: the role names differ between Neon projects, so
# ownership and GRANT statements from the source would fail on the target.
docker run --rm -e PGURL="$SOURCE_URL" "$PG_IMAGE" \
  sh -c 'pg_dump "$PGURL" --no-owner --no-privileges --format=plain --quote-all-identifiers' \
  > "$dump_file"
echo "    wrote $(wc -c < "$dump_file" | tr -d ' ') bytes"

echo "==> Restoring into target"
# ON_ERROR_STOP so a partial restore fails loudly instead of leaving a half-copied
# database that looks fine until someone hits the missing rows.
docker run --rm -i -e PGURL="$TARGET_URL" "$PG_IMAGE" \
  sh -c 'psql "$PGURL" -v ON_ERROR_STOP=1 -X --quiet' \
  < "$dump_file"

echo "==> Comparing row counts"
failed=0
for table in "${TABLES[@]}"; do
  src="$(psql_value "$SOURCE_URL" "SELECT count(*) FROM \"$table\"")"
  dst="$(psql_value "$TARGET_URL" "SELECT count(*) FROM \"$table\"")"
  if [[ "$src" == "$dst" ]]; then
    printf '    %-28s %6s = %-6s ok\n' "$table" "$src" "$dst"
  else
    printf '    %-28s %6s != %-6s MISMATCH\n' "$table" "$src" "$dst"
    failed=1
  fi
done

# goose tracks applied migrations here. If it did not come across, the next
# `migrate` run would try to re-apply 001_init and fail on existing tables.
goose_src="$(psql_value "$SOURCE_URL" \
  "SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='goose_db_version'")"
if [[ "$goose_src" == "1" ]]; then
  src="$(psql_value "$SOURCE_URL" 'SELECT max(version_id) FROM goose_db_version')"
  dst="$(psql_value "$TARGET_URL" 'SELECT max(version_id) FROM goose_db_version')"
  if [[ "$src" == "$dst" ]]; then
    printf '    %-28s %6s = %-6s ok\n' "goose version" "$src" "$dst"
  else
    printf '    %-28s %6s != %-6s MISMATCH\n' "goose version" "$src" "$dst"
    failed=1
  fi
fi

if [[ "$failed" != "0" ]]; then
  echo
  echo "Row counts do not match. Do NOT switch DATABASE_URL yet." >&2
  exit 1
fi

echo
echo "Copy verified. Next steps:"
echo "  1. Point DATABASE_URL at the new project's POOLED endpoint in Vercel."
echo "  2. Set \"regions\": [\"fra1\"] in vercel.json."
echo "  3. Redeploy, then confirm with:"
echo "       curl -sI https://www.edulicense.uz/admin/login | grep x-vercel-id"
echo "  4. Keep the old project until the new one has served real traffic."
