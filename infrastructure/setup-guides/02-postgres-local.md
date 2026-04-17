# 02 — Run Postgres locally

**Goal:** a Postgres 14+ database running on `localhost:5432` with a `flux` user and `flux_dev` database ready to connect to from Go code.
**Who runs this:** Daniel. Claude does not execute install or DB commands.

## Prerequisites

- macOS + Homebrew (see `01-go-toolchain.md`)
- Nothing else bound to port 5432 (check with `lsof -i :5432`; if it returns something, stop or reconfigure it first)

## Why local Postgres (not Docker yet)

At L01 we want minimum moving parts — learn Go + SQL first, containers later. Docker arrives at L03. If you prefer Docker *now*, that's fine — you'll just do L03's work earlier. Either way, get Postgres running somehow.

## Step 1: Install Postgres

```bash
brew install postgresql@16
```

(Any recent major version — 14, 15, 16 — is fine. 16 is current stable.)

## Step 2: Start Postgres as a background service

```bash
brew services start postgresql@16
```

This keeps it running and auto-starts on login. Management commands:

- Stop: `brew services stop postgresql@16`
- Status: `brew services list`
- Restart: `brew services restart postgresql@16`

## Step 3: Verify it's running

```bash
psql postgres -c "SELECT version();"
```

Expected: a row printing the Postgres version.

**If this fails** (e.g., *"could not connect to server"*): the service didn't start. Check `brew services list` — state should be `started`. If not, read the Postgres logs:

```bash
brew services info postgresql@16
# follow the log path shown, e.g., tail -f /opt/homebrew/var/log/postgresql@16.log
```

## Step 4: Create the dev database + user

We want a dedicated app user (`flux`) distinct from the default superuser. Good habit — production never uses a superuser.

```bash
# Create a user with password 'flux' (dev-only; don't reuse for anything real)
psql postgres -c "CREATE USER flux WITH PASSWORD 'flux';"

# Create the dev database owned by flux
psql postgres -c "CREATE DATABASE flux_dev OWNER flux;"

# Explicit privileges grant (belt + suspenders)
psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE flux_dev TO flux;"
```

## Step 5: Connect as the flux user

```bash
psql postgres://flux:flux@localhost:5432/flux_dev -c "SELECT current_user, current_database();"
```

Expected output:
```
 current_user | current_database
--------------+------------------
 flux         | flux_dev
```

## Step 6: Save the connection string

L01's Go code reads the Postgres URL from the `DATABASE_URL` environment variable. Set it in your shell now — either in `~/.zshrc` (permanent, global):

```bash
export DATABASE_URL="postgres://flux:flux@localhost:5432/flux_dev?sslmode=disable"
```

Or in a local `.env` file at the repo root (gitignored — `.env` is already in `.gitignore`):

```bash
DATABASE_URL="postgres://flux:flux@localhost:5432/flux_dev?sslmode=disable"
```

> `sslmode=disable` is fine for local dev. In production that'd be `require` or `verify-full`.

**Prefer the `.env` approach** — keeps repo-specific config out of your global shell. You'll need a tool to load it (`direnv`, or load manually with `source .env`). For L01, `source .env` before running the server is sufficient.

## Step 7 (L01 prep): Enable `pgcrypto` for UUID generation

The `workflows` table uses `gen_random_uuid()` for primary keys, which needs the `pgcrypto` extension (built into Postgres 13+, just needs enabling):

```bash
psql $DATABASE_URL -c "CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";"
```

**Step 7b (optional, deferred):** the actual `CREATE TABLE` for `workflows` is part of L01 itself — you'll write it in `project/schema.sql` when setting up the module. Don't create the table yet; that's the first L01 commit.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `could not connect to server` | Service not running | `brew services start postgresql@16` |
| `role "postgres" does not exist` | Homebrew uses your system user as superuser | Use `psql -U $USER postgres` instead, or `createuser -s postgres` |
| `port 5432 already in use` | Another Postgres / Docker bound | `lsof -i :5432` to find it; stop or pick different port |
| `authentication failed for user "flux"` | Wrong password / missing user | Re-run Step 4; default pg_hba.conf should accept password auth on localhost |

## What Claude does NOT do here

- Does not run `brew install postgresql@16`.
- Does not run any `psql` commands.
- Does not create the database or user.
- Does not edit your `~/.zshrc` or any `.env` file.

Claude's role in this guide is to explain *why* each step matters and help you debug error output — not to execute.

## Verification

Before returning to L01, you should be able to run:

```bash
psql $DATABASE_URL -c "\dt"
```

And see either an empty result (no tables yet — correct for now) or the message `Did not find any relations.` — both mean the connection works.

Then:

```bash
psql $DATABASE_URL -c "SELECT gen_random_uuid();"
```

Should print a UUID. Confirms `pgcrypto` is enabled.

## Next

Postgres + Go both verified? Return to `levels/L01-mvs.md` and begin the Go module setup.
