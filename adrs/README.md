# Architecture Decision Records

Each file in this directory captures one load-bearing decision: what was decided, why, what was considered and rejected, and what consequences follow. ADRs are append-only — superseding rather than rewriting.

The cross-ADR synthesis (postures and invariants that span multiple ADRs) lives in `project/docs/architecture.md`. Tactical state lives in `STATUS.md`. Module layout lives in the project tree.

## Catalog

| # | Title | Status | Date | One-line |
|---|---|---|---|---|
| [0001](0001-router-choice.md) | HTTP router choice | Accepted | 2026-04-18 | `net/http` + `http.ServeMux` (Go 1.22 method+pattern). No chi / gin / echo. |
| [0002](0002-postgres-driver.md) | Postgres driver choice | Accepted | 2026-04-18 | `pgx/v5` over `database/sql`. Postgres-first vs SQL-generic posture. |
| [0003](0003-malformed-uuid-translation.md) | Malformed UUID translation | Accepted | 2026-04-21 | SQLSTATE 22P02 → `ErrNotFound` at storage boundary; surfaces as 404. |
| [0004](0004-http-response-utilities.md) | HTTP response utilities | Accepted | 2026-04-25 | Shared `internal/httpx`; Shape 2 envelope `{"error": "..."}`; handler owns status. |
| [0005](0005-workflow-lifecycle-ops.md) | Workflow lifecycle ops | Accepted | 2026-04-30 | PUT strict / DELETE idempotent 204; soft-delete via `deleted_at`. |
| [0006](0006-test-strategy.md) | Test strategy | Accepted | 2026-05-01 | Path 1C + 2C + 3A — fakeStore primary + integration ring + envelope helper. |

Queued (not yet drafted): `auth-model`, `tenancy-isolation` — both for L02 proper.

## When to write an ADR

From `.claude/rules/collaboration.md`:

- Adding a new load-bearing dependency (framework, major library, new DB).
- Architectural splits (extracting a service, splitting a module).
- DB / schema changes with migration implications.
- Auth model changes.
- Anything touching security posture.
- Major refactors.
- Cloud resource decisions.

Each ADR: Context → Options Considered → Decision → Consequences. Keep them tight; the goal is durability, not exhaustive prose.

## Numbering

Sequential, zero-padded to four digits. New ADRs take the next number, period. Superseded ADRs keep their number and add a `**Status:** Superseded by NNNN` line.
