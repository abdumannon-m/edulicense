# Moving the app and database to Frankfurt

## Why

The Vercel function and the Neon database must sit in the same region. Every admin
page issues several database queries one after another — `/admin/overview` alone
makes nine round trips before rendering — so any distance between function and
database gets multiplied by nine.

The original deployment had the function in `iad1` (Washington DC) and the database
in `ap-southeast-1` (Singapore): roughly 15,000 km, ~230 ms per round trip, about
two seconds of pure waiting per page on tables holding almost no data.

Frankfurt is the closest region where **both** Vercel and Neon operate. Neon runs in
only eight AWS regions, and Dubai and Mumbai — physically nearer to Tashkent — are
not among them.

| Pair | Vercel | Neon | Distance from Tashkent |
|---|---|---|---|
| **Frankfurt** | `fra1` | `aws-eu-central-1` | ~4,700 km |
| London | `lhr1` | `aws-eu-west-2` | ~5,600 km |
| Singapore | `sin1` | `aws-ap-southeast-1` | ~5,700 km |

## Order of operations

The two halves must move together. Setting `fra1` while the database is still in
Singapore recreates the original problem in a new direction, so the region in
`vercel.json` changes **last**.

### 1. Create the new Neon project

Neon cannot move an existing project: *"You cannot change the region for an existing
project."* Create a new one in the Neon console, region **AWS Europe (Frankfurt)**.

Copy both connection strings from the new project — the direct endpoint and the
pooled one (`-pooler` in the host). Both are needed, for different steps.

### 2. Copy the data

```bash
SOURCE_URL='<current direct URL, ap-southeast-1>' \
TARGET_URL='<new direct URL, eu-central-1>' \
./scripts/migrate-region.sh
```

Use the **direct** endpoints here, not the pooled ones — the pooler runs through
pgbouncer, which does not support the session-level operations `pg_dump` and `psql`
need for a restore.

The script refuses to run if the target already has tables, compares row counts for
all eight tables afterwards, and checks that goose's migration state came across. It
exits non-zero on any mismatch. Docker is the only prerequisite; it pulls the
Postgres client itself.

### 3. Repoint the application

In the Vercel project's environment variables, set `DATABASE_URL` to the new
project's **pooled** endpoint. Serverless functions open a lot of short-lived
connections, which is what the pooler is for.

### 4. Switch the function region

In `vercel.json`:

```json
"regions": ["fra1"]
```

On Hobby, `regions` in `vercel.json` is ignored — set the single region in the Vercel
project settings instead.

### 5. Redeploy and verify

```bash
curl -sI https://www.edulicense.uz/admin/login | grep x-vercel-id
```

Expect `fra1::fra1::…`. Two different codes, such as `hkg1::iad1::…`, mean the edge
PoP and the function region differ — the second is the one that matters, and it must
read `fra1`.

Then time a real page:

```bash
curl -s -o /dev/null -w 'ttfb=%{time_starttransfer}s\n' https://www.edulicense.uz/admin/login
```

Before the move this measured 0.57–0.91 s from Tashkent, against 0.35 s for the
static home page.

### 6. Keep the old project

Leave the Singapore project running until the new one has served real traffic for a
day or so. Delete it only once you are confident, since restoring means repeating
this whole process in reverse.

## Also worth doing

**Check autosuspend.** Neon scales the database to zero after about five minutes
idle. For an admin tool used a few times a day, most visits pay a wake-up cost of
roughly 0.5–2 s — which can matter more day to day than the region move. Raise the
timeout or disable it in the Neon console under the project's compute settings.

**Leave the query count alone.** Collapsing `DashboardStats`' seven queries into one
looks like the obvious fix, but once the function sits beside the database each round
trip costs about a millisecond, so the whole change saves perhaps 15 ms. Not worth
the churn unless the admin still feels slow after the move.
