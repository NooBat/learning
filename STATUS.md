# STATUS

**Current level:** L01 — mvs (not started yet)
**Last updated:** 2026-04-17

## This-week focus

Environment setup + first Go handler.

## Next-session target

1. Follow `infrastructure/setup-guides/01-go-toolchain.md` — install Go, verify with `go version`.
2. Follow `infrastructure/setup-guides/02-postgres-local.md` — install Postgres, create `flux_dev` database.
3. Initialize the Go module in `project/` (`go mod init <your-github-path>/project`).
4. Write the first HTTP handler: `POST /workflows` that accepts JSON and writes to Postgres.

Read `levels/L01-mvs.md` first for full scope and exit criteria.

## Open blockers

- **Private GitHub repo not yet created** — needs to exist before first `git push`. Daniel: create `learning` (or another name) as a private repo at github.com, then follow the "push" commands Claude will output once git is initialized.

## In-flight ADRs

- None yet. Expect the first one at L01: **router choice** (`net/http` vs `chi` vs `gin`/`echo`). Use `/write-adr` when the decision comes up.
