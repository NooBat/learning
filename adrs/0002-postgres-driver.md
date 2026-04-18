# 0002. Postgres driver choice

**Status:** Accepted
**Date:** 2026-04-18
**Level:** L01 (mvs)

## Context

L01 connects to Postgres (`flux_dev`, via `DATABASE_URL`) and performs CRUD on `workflows` rows:

- `INSERT` for `POST /workflows`
- `SELECT` for `GET /workflows`, `GET /workflows/{id}`
- `UPDATE` for `PUT /workflows/{id}`
- `DELETE` for `DELETE /workflows/{id}`

The `workflows.steps` column is `jsonb`, which makes driver support for Postgres-specific types (vs generic text) material.

The driver choice shapes:

- **API style** — canonical `database/sql` interface (Query/Exec/Scan with `any`-typed targets) vs pgx native (typed methods, generics helpers).
- **Scan ergonomics** — how much ceremony to read a row into a struct.
- **Access to Postgres-specific features** — `LISTEN/NOTIFY`, `COPY`, batch queries, native JSONB handling.
- **Connection pool semantics** — `sql.DB` pool vs `pgxpool.Pool`.
- **Future portability** — would we ever want to point at a different SQL database? (Realistically: no, for this project. But the decision still has weight.)

Unlike the router choice, the Go ecosystem does *not* have a settled "stdlib is back" answer here. There's real debate.

## Options Considered

### Option A: `pgx` native (v5)

- **Pros:**
  - Native Postgres protocol implementation — no `database/sql` indirection.
  - First-class support for Postgres types: JSONB, arrays, intervals, `timestamptz`, UUIDs — typed, not string-encoded.
  - `pgx.CollectRows[T]()` + `pgx.RowToStructByName[T]()` give clean generic struct scanning (Go 1.18+).
  - Direct access to Postgres-specific features: `LISTEN/NOTIFY` (useful around L05), `COPY`, batched queries.
  - Actively maintained; considered the modern default for new Go+Postgres projects.
  - `pgxpool.Pool` has clearer pool semantics than `sql.DB` in many writeups.
- **Cons:**
  - Postgres-only. If we ever wanted to swap (unlikely), the SQL-execution layer is a full rewrite.
  - Teaches a pgx-specific mental model instead of the canonical `database/sql` pattern that every classic Go tutorial uses.
  - `pgxpool` vs `sql.DB` naming divergence can confuse when reading mixed-era Go examples.
- **Notes:** Recommended by the pgx authors themselves over the stdlib-adapter mode for new projects.

### Option B: `database/sql` + `jackc/pgx/v5/stdlib` (hybrid)

- **Pros:**
  - Uses the canonical `database/sql` API — same patterns as the Go standard library tutorials and *Learning Go*.
  - Under the hood, pgx handles the wire protocol (so you still get modern performance).
  - Stable contract: if we ever swap to `lib/pq` or up to pgx native, only the driver import changes, not query code.
  - You learn the baseline `database/sql` patterns (Scan, Exec, QueryRow) that transfer to any driver.
- **Cons:**
  - pgx-specific features (`LISTEN/NOTIFY`, `COPY`, batch) are *inaccessible* through the `database/sql` interface — you'd have to drop to pgx native to use them.
  - `database/sql` Scan uses variadic `any` pointers — type safety is at runtime, not compile-time.
  - `sql.DB` pool has its own set of knobs (`MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`) that need tuning separately.
- **Notes:** A defensible "middle path" that defers the pgx-native commitment.

### Option C: `database/sql` + `lib/pq`

- **Pros:**
  - Canonical Go SQL style. Every 2015-2022 tutorial you find online uses this.
  - Pure Go, no cgo.
- **Cons:**
  - Widely considered maintenance-mode; the `lib/pq` README itself points at pgx for new projects.
  - Protocol implementation lags behind pgx on newer Postgres features.
- **Notes:** Included for comparison. Not a serious contender for new projects in 2026.

### Option D: `database/sql` + `pgx/stdlib` + `jmoiron/sqlx`

- **Pros:**
  - `sqlx` adds `Get()`, `Select()`, and `StructScan()` — removes most of the Scan ceremony from Option B.
  - Keeps the `database/sql` interface stable while giving more ergonomic struct loading.
- **Cons:**
  - Another dependency to track on top of pgx.
  - Doesn't restore access to pgx-specific features — still hidden behind the stdlib interface.
  - Two libraries to learn (the `database/sql` contract and sqlx extensions).
- **Notes:** Popular with teams that want stdlib conventions + ergonomic scan. Arguably overkill for 5 CRUD handlers at L01.

## Decision

**Chosen:** A — `pgx` native (v5) with `pgxpool` for the connection pool.

The architectural posture: **this is a Postgres-first system, and the data access layer should speak Postgres natively** rather than hide behind a lowest-common-denominator interface designed to also work with MySQL and SQLite. The `workflows.steps` column is JSONB — a first-class part of the domain model, not an implementation detail — and the abstraction tax of round-tripping it through `[]byte` + `json.Unmarshal` at every query site is noise we'd rather not pay. Looking forward: several upcoming capabilities (LISTEN/NOTIFY for async workers at L05, batched queries at L06) are Postgres-specific, so sitting behind the `database/sql` facade would mean hopping the fence anyway once we get there — better to start on the side we'll end up on. The coupling we're explicitly accepting: the data access layer is no longer driver-portable (swapping DBs means rewriting query sites), and future contributors need to internalize the pgx mental model instead of the canonical `database/sql` one. That trade is defensible because portability to "some other SQL database" is imaginary for this project, while fidelity to the Postgres we actually run is real and compounds.

## Consequences

**Positive (any choice):**

- L01 exit criteria (CRUD over curl, env-driven `DATABASE_URL`) are reachable with any option.
- Pairs with ADR 0001 to lock in the full "first dependency footprint" for the project.

**Negative (option-conditional; to finalize on Decision):**

- If A (pgx native): committed to Postgres-only; learning a pgx-specific API instead of the canonical `database/sql` pattern.
- If B (hybrid): pgx-specific features unreachable without dropping to native; `any`-typed Scan is runtime-checked only.
- If C (lib/pq): adopting a maintenance-mode driver in 2026; some newer Postgres features lag.
- If D (hybrid + sqlx): extra library; still no access to pgx-native features.

**Neutral:**

- Any option requires hand-written SQL at L01 (one `project/schema.sql` + query strings in handler code). Migration tools arrive at L03.

## Revisit triggers

- **L03** — introducing migrations (goose / golang-migrate) and structured logging around queries. If Scan ceremony is burning us, reconsider.
- **L05** — async execution / background jobs. If we need `LISTEN/NOTIFY`, `COPY`, or batch queries, the native-pgx side becomes materially more attractive (and staying on B/C/D means dropping to pgx native just for those paths).
- **L06** — performance / caching. If pool behavior becomes load-bearing, `pgxpool` vs `sql.DB` semantics matter.

## References

- [pgx v5 — native API](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [pgx v5 — `stdlib` driver](https://pkg.go.dev/github.com/jackc/pgx/v5/stdlib)
- [lib/pq](https://github.com/lib/pq)
- [jmoiron/sqlx](https://github.com/jmoiron/sqlx)
- [`database/sql` overview](https://pkg.go.dev/database/sql)
- ADR 0001 — router choice (this ADR completes the "first dependencies" pair)
- L01 brief: [`levels/L01-mvs.md`](../levels/L01-mvs.md)
