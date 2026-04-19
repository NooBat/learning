# L01 — Minimum Viable Service

**Status:** complete (with documented deferrals — see "What actually shipped")
**Tier:** 1 — Backend Engineer
**Started:** 2026-04-18
**Completed:** 2026-04-20

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

- [x] Go module initialized in `project/` (`go mod init github.com/NooBat/learning/project`)
- [x] `workflows` table exists in `flux_dev` database (schema: id uuid, name text, trigger_type text CHECK-constrained, steps jsonb, timestamps)
- [x] HTTP server runs via `go run ./cmd/server` and listens on `:8080`
- [x] `POST /workflows` accepts JSON, validates at the handler edge, creates a row, returns 201 + created resource (hydrated with server-assigned `id` / `created_at` / `updated_at` via `INSERT ... RETURNING`)
- [x] `GET /workflows` returns JSON array of all workflows (200), newest-first by `created_at`
- [x] `GET /workflows/{id}` returns a single workflow (200) or 404
- [ ] ~~`PUT /workflows/{id}` updates a workflow (200) or 404~~ — **deferred to L02** (see "What actually shipped" below)
- [ ] ~~`DELETE /workflows/{id}` deletes a workflow (204) or 404~~ — **deferred to L02**
- [x] Database connection string read from `DATABASE_URL` environment variable (not hardcoded)
- [ ] ~~Manual end-to-end test with `curl` documented in `notes/l01-curl-tests.md`~~ — **deferred** (B-choice at closeout: a checked-in smoke script was deemed lower value than the `fakeStore` test-suite investment planned at L02 opener; smoke test was still run interactively at closeout)
- [x] Two ADRs written — `0001-router-choice` (→ `net/http` stdlib) and `0002-postgres-driver` (→ `pgx` native)
- [ ] ~~Architecture doc stub created at `project/docs/architecture.md`~~ — **deferred to L02** (the layering the doc would describe — consumer-defined `storage` interface, handler↔storage boundary translation, validation-at-the-edge posture — is captured in the 2026-04-20 LOG entry; a doc stub at `project/docs/architecture.md` will land alongside the L02 opening changes)

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

## What actually shipped (close recap — 2026-04-20)

L01 closed with **CRUD-minus-U-D**: Create, Read-one, Read-list — plus all supporting architecture. The mutation pair (PUT/DELETE) was scope-narrowed mid-session because the remaining architectural lessons L01 was structured to teach were all covered:

- **Boundary translation** (pgx driver errors ↔ domain sentinels) — exercised by `GetByID` (`pgx.ErrNoRows` → `ErrNotFound`) and `Create` (catch-all path).
- **Consumer-defined interfaces** — `handler.go`'s unexported `storage` interface, not a producer-exported `IStorage` on `storage.go`. Pays off as a test affordance at L02.
- **Validation at the edge** — `validate(*Workflow)` mutates-in-place (TrimSpace) and wraps `ErrInvalidInput`; DB CHECK constraint is the invariant source of truth.
- **Buffer-first JSON encode posture** — materialize full response before any byte is written, so mid-encode failures still return 500 instead of a corrupted 200. Trade memory for recoverability.
- **Method-prefixed routing** (Go 1.22+ stdlib `ServeMux`) — `POST /workflows`, `GET /workflows/{id}`, `GET /workflows`; auto-405 on wrong method.
- **Lifecycle orchestration** — signal-driven ctx via `signal.NotifyContext`, buffered-size-1 error channel, goroutine-wrapped `ListenAndServe`, select on `ctx.Done() vs errCh`, shutdown context rooted at `context.Background()` (not the already-cancelled parent).

PUT/DELETE would be straightforward extensions of the patterns already in place — no new architectural lessons, ~30 additional lines across storage + handler + one more interface method + one route. Moved to L02 as a warm-up.

### Deferred to L02 (tracked in `STATUS.md` and the 2026-04-20 LOG entry)

1. Implement `PUT` and `DELETE` for workflows. Extend the consumer-defined `storage` interface, add two storage methods, two handler methods, two route registrations.
2. Translate Postgres `SQLSTATE 22P02` (malformed UUID) at the storage boundary. Today it surfaces as a 500. Map to `ErrNotFound` (404 posture) or a new `ErrInvalidID` sentinel (400 posture) — the architectural call is which client contract is cleaner.
3. Write the first Go test suite using `httptest.NewServer` + a `fakeStore` satisfying the consumer-defined `storage` interface. The interface is already the right shape; the fake is ~20 lines and covers every 400/404/500 branch.
4. Write `project/docs/architecture.md` stub describing the layering.
5. Optional: checked-in smoke script (`scripts/smoke.sh`) once the test suite lands — lower priority than Go tests but cheap.

### Known architectural gap

`GET /workflows/<malformed-uuid>` returns 500. Root cause: Postgres raises `SQLSTATE 22P02 (invalid_text_representation)` when `$1` isn't a valid UUID; this error falls through the handler's error ladder to the catch-all 500 branch. Clean fix lives at the storage boundary (see deferred item 2). L02 will close this.
