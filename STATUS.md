# STATUS

**Current level:** L01 — mvs (in progress)
**Last updated:** 2026-04-18

## This-week focus

Environment setup + first Go handler.

## Next-session target

Setup done (Go + Postgres installed). Remaining L01 opening steps:

1. Sanity-check the toolchain: `go version`, `psql -U flux -d flux_dev -c "SELECT 1;"`.
2. Write **ADR 0001 — router choice** (`net/http` vs `chi` vs others). Gate on this before coding.
3. Write **ADR 0002 — postgres driver** (`pgx` vs `database/sql` + `lib/pq`). Gate on this before coding.
4. `go mod init github.com/NooBat/learning/project` inside `project/`.
5. First HTTP handler: `POST /workflows` that accepts JSON and writes to Postgres.

Read `levels/L01-mvs.md` for full scope and exit criteria.

## Open blockers

- None. Remote repo `NooBat/learning` exists and initial scaffold pushed (`d516944`).

## In-flight ADRs

- **0001 router-choice** — to be written before first handler. Claude's bias: `chi` for stdlib-compatibility + minimal ergonomics.
- **0002 postgres-driver** — to be written before first handler. Claude's bias: `pgx` for modern Go idioms + active maintenance.
