# L01 — Minimum Viable Service

**Status:** not started
**Tier:** 1 — Backend Engineer
**Started:** —
**Completed:** —

## Goal

Build the first working version of the keystone Go service: an HTTP server with Postgres persistence for **workflow definitions**. CRUD endpoints only — no execution, no auth, no tests yet.

A workflow is the core domain entity. At L01 it's just a record with a name, a trigger type, and a list of steps stored as JSON. Execution comes at L05.

## Why this level exists

L01 is the "can you ship anything in Go at all?" level. By the end you have a process that speaks HTTP and reads/writes a database. It's deliberately simple — the goal is to expose Go fundamentals (modules, error handling, package layout, `net/http`, Postgres interaction) without distraction.

Future levels layer on top (auth, tests, Docker, async work). Focus here on making the structure sane, not clever.

## Prerequisites

- Go installed (`infrastructure/setup-guides/01-go-toolchain.md`)
- Postgres running locally with `flux_dev` database (`infrastructure/setup-guides/02-postgres-local.md`)
- Basic Go syntax familiarity — if you've written zero Go, skim [A Tour of Go](https://go.dev/tour/) first (~1-2 hours)

## Exit criteria

All checkable. All must be true before claiming L01 done.

- [ ] Go module initialized in `project/` (`go mod init github.com/<user>/learning/project` or similar)
- [ ] `workflows` table exists in `flux_dev` database (schema: id uuid, name text, trigger_type text, steps jsonb, timestamps)
- [ ] HTTP server runs via `go run ./cmd/server` and listens on `:8080`
- [ ] `POST /workflows` accepts JSON, validates minimally, creates a row, returns 201 + created resource
- [ ] `GET /workflows` returns JSON array of all workflows (200)
- [ ] `GET /workflows/{id}` returns a single workflow (200) or 404
- [ ] `PUT /workflows/{id}` updates a workflow (200) or 404
- [ ] `DELETE /workflows/{id}` deletes a workflow (204) or 404
- [ ] Database connection string read from `DATABASE_URL` environment variable (not hardcoded)
- [ ] Manual end-to-end test with `curl` documented in `notes/l01-curl-tests.md` (covers all 5 endpoints)
- [ ] At least one ADR written — the router-choice ADR is almost certain (see below)
- [ ] Architecture doc stub created at `project/docs/architecture.md` describing the L01 shape

## Scope

### In-scope

- **Package layout:** `cmd/server/main.go` (entry point) + `internal/workflows/` (handlers + storage). The `cmd/` + `internal/` pattern is idiomatic Go and scales.
- **HTTP routing:** `net/http` stdlib OR a minimal router like `chi`. (ADR decision — see below.)
- **Postgres driver:** `pgx` OR `database/sql` + `lib/pq`. (ADR decision — see below.)
- **Schema:** a single hand-written SQL file for L01 (`project/schema.sql`). Migration tools come at L03.
- **JSON:** stdlib `encoding/json`. No extra library needed.
- **Error handling:** return appropriate HTTP status codes. 400 for malformed JSON, 404 for not found, 500 for DB errors, 201/200/204 on success.
- **Logging:** `log.Printf` or `log.Println` is fine. Structured logging (`log/slog`) comes at L03.

### Out of scope (deliberately deferred)

- Authentication / tenancy → L02
- Automated tests → L02
- Docker / containers → L03
- Migration tool (goose/golang-migrate) → L03
- Structured logging (`log/slog`) → L03
- Background execution of workflows → L05
- Queues, caches, retries → Tier 2

Resist the temptation to add these early. Each one arrives at the level that teaches it.

## ADR-worthy decisions likely to come up

Pause and invoke `/write-adr <slug>` before implementing each:

1. **Router choice** (`router-choice`) — `net/http` vs `chi` vs `gorilla/mux` vs `gin`/`echo`. Trade-offs: stdlib-only simplicity vs ergonomic routing vs ecosystem weight. My bias: `chi` for L01 — small, idiomatic, stdlib-compatible.
2. **Postgres driver** (`postgres-driver`) — `pgx` (modern, native) vs `database/sql` + `pq` (more portable, older). My bias: `pgx` — better performance, better types, active maintenance.
3. **Package layout** (`package-layout`) — `cmd/` + `internal/` vs flat. Recommendation: `cmd/` + `internal/` from day one. Not really an ADR — more a standard — but document the choice.
4. **`steps` storage** (`workflow-steps-storage`) — JSONB column (queryable), plain JSON text, or normalized step tables. My bias: JSONB for L01 simplicity; normalize when querying-into-steps becomes needed.

## Reading triggered by this level

- ★ *Learning Go* by Jon Bodner (2nd ed.) — Chapters 1-5 (basics), then 9 (errors), 11 (testing) as needed. Skim 12 (context) — context will come back at L05.
- [Go docs: `net/http`](https://pkg.go.dev/net/http) — HTTP server section.
- [Effective Go](https://go.dev/doc/effective_go) — read "Formatting", "Commentary", "Names", "Errors", "Packages" sections.
- If choosing `pgx`: [pgx v5 docs](https://pkg.go.dev/github.com/jackc/pgx/v5).
- If choosing `chi`: [chi GitHub README](https://github.com/go-chi/chi).

## Stretch (only if flying)

- `/healthz` endpoint returning `{"status":"ok"}` — trivial but good habit.
- Simple request logging middleware (log method + path + duration).
- Use [REST Client for VS Code](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) with a `.http` file instead of `curl` for reproducible manual tests. Save as `project/testdata/http-requests.http`.

## Anti-scope (do NOT do)

- Don't add tests. That's L02. You'll be tempted. Resist.
- Don't add Docker. That's L03.
- Don't refactor into "clean architecture" layers. L01 is flat-ish on purpose.
- Don't add more features than the 5 CRUD endpoints.

## When you think you're done

Run this checklist:

1. Walk through each exit criterion and check it off honestly.
2. Run the curl tests end-to-end. Document them in `notes/l01-curl-tests.md`.
3. Write or update `project/docs/architecture.md` with the L01 shape (package layout, routes, schema).
4. Commit + push.
5. Invoke `/start-level 02 auth-tenancy` to scaffold the next level.

If any criterion isn't met, don't transition. Finish it.
