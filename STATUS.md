# STATUS

**Current level:** L02 — auth-tenancy (in-progress, started 2026-04-20)
**Last updated:** 2026-05-10

## This-week focus

L02 warm-ups: warm-up #3 handler-side complete; integration ring queued.

**Done (session arc 04-21 → 05-10):**

1. ✅ ADR 0003 (`malformed-uuid-translation`) accepted — 22P02 → `ErrNotFound` (404). Storage-boundary translation; ADR 0001-style presence-disclosure posture pinned for the level.
2. ✅ ADR 0004 (`http-response-utilities`) accepted — shared `internal/httpx` package + Shape 2 JSON envelope (`{"error": "..."}`) + Ownership 1 (handler picks status). Compile-time invariant for byte-identical 404 across packages.
3. ✅ `internal/httpx` package shipped (`WriteJSON`, `WriteError`, `DecodeJSON`). Handler refactored — buffer-encode dance + `http.Error` calls collapsed.
4. ✅ Request body cap at handler edge — `http.MaxBytesReader(w, r.Body, 1*MiB)` + `*http.MaxBytesError` branch returning 413 (option B placement).
5. ✅ ADR 0005 (`workflow-lifecycle-ops`) accepted — PUT strict / DELETE idempotent 204 / soft delete via `deleted_at`. Pairing B chosen: extends ADRs 0003 + 0004 opacity stance to a third surface; three ADRs now share the posture.
6. ✅ PUT/DELETE handlers shipped. Schema: `deleted_at timestamptz` inlined. Storage: `Update`/`Delete` + soft-delete filter on `GetByID`/`List`. `translatePgError` helper consolidates ADR 0003's 22P02 → ErrNotFound rule across all storage methods.
7. ✅ Smoke: 13 cases verified live. Pairing B opacity contract holds — DELETE on existing/just-deleted/never-existed/malformed-UUID all return byte-identical 204. PUT 404 collapses missing/soft-deleted/malformed-UUID into one response.
8. ✅ ADR 0006 (`test-strategy`) accepted — Path 1C + 2C + 3A: handler-level fakeStore primary + integration ring (`//go:build integration`) + envelope-decode helper centralizes ADR 0004's Shape 2 contract + minimal map-backed fakeStore (no fault injection at L02). Cashes the L01 `storage`-interface design payoff; locks ADR 0005 opacity invariant as a regression test.
9. ✅ Test suite scaffolded — `internal/workflows/handler_test.go` (fakeStore + helpers: `errorMessage`, `requireStatus`, `decodeWorkflow`, `decodeWorkflowList`, `postJSON`, `putJSON`, `seedWorkflow`, `newTestServer`, `readBody`) + `internal/workflows/storage_integration_test.go` (build-tagged, 3 named tests still `t.Skip`).
10. ✅ `TestDelete_OpacityInvariant` shipped — promotes the four-DELETE-paths byte-identical wire assertion from one-time smoke to compile-time regression test. Encodes ADRs 0003 + 0004 + 0005 in one assertion.
11. ✅ Handler branch coverage: 22 tests green covering 200 / 201 / 204 / 400 / 404 / 413. Create / Update / List / GetByID / Delete each with success + error paths. Soft-delete invisibility asserted at every read site (List + GetByID + Update). 500 path deferred per ADR 0006 cons-accepted (no fault injection at L02).

**Queued (in order):**

1. **Integration ring (warm-up #3 tail).** Provision `DATABASE_URL_TEST` (separate test DB or accept dev-DB clobber via TRUNCATE), then implement the three named integration tests already scaffolded in `storage_integration_test.go`: `TestStorage_Update_RETURNING`, `TestStorage_GetByID_22P02_to_NotFound` (the only end-to-end assertion for ADR 0003's translation rule against the real driver), `TestStorage_Delete_SoftDeleteFilter`. Run via `go test -tags=integration ./internal/workflows/`.
2. **`project/docs/architecture.md` stub (warm-up #4).** Document the L01 baseline + L02 additions (httpx, ADR 0003 boundary translation, soft-delete pattern, test pyramid shape) before auth/tenancy land on top.

After warm-ups: L02 proper — auth middleware + tenancy column + ADRs for auth model, tenancy isolation.

## Next-session target

1. **Decide test-DB convention.** Option A: separate `flux_test` database via `psql -c 'CREATE DATABASE flux_test'` + new env var. Option B: reuse dev DB with `DATABASE_URL_TEST=$DATABASE_URL`, accept TRUNCATE clobbering smoke-test rows. Option A is production-CI-shaped; Option B unblocks the tests in 30 seconds.
2. **Implement the three integration tests.** Suggested shapes already in `storage_integration_test.go` doc comments. The 22P02-translation test is the architecturally distinct one — only real driver round-trip produces a real `*pgconn.PgError` with code 22P02; fakeStore can't simulate it.
3. **Verify the build-tag discipline holds.** Default `go test ./...` skips integration ring; `go test -tags=integration ./...` runs everything. Document the run command in warm-up #4's architecture stub.

## Open blockers

- None. Branch is ahead of `origin/main` by N local commits (this session added: ADR 0006 + scaffold + 4 test commits = 6 commits). Push when ready.

## In-flight ADRs

- ADR 0003 ✅ accepted (`adrs/0003-malformed-uuid-translation.md`).
- ADR 0004 ✅ accepted (`adrs/0004-http-response-utilities.md`).
- ADR 0005 ✅ accepted (`adrs/0005-workflow-lifecycle-ops.md`).
- ADR 0006 ✅ accepted (`adrs/0006-test-strategy.md`).
- ADR queue: `auth-model`, `tenancy-isolation` (both for L02 proper, after warm-ups #3 tail + #4).

## Known gaps carried from L01

- ~~**Malformed UUID → 500.**~~ ✅ Fixed by ADR 0003 implementation.
- ~~**`PUT /workflows/{id}` and `DELETE /workflows/{id}` not implemented.**~~ ✅ Shipped via ADR 0005.
- ~~**No automated tests yet.**~~ ✅ Handler-level suite lives at `internal/workflows/handler_test.go` (22 tests, 200/201/204/400/404/413 covered). Integration ring scaffolded but not implemented.
- **`project/docs/architecture.md` does not exist.** Queued as warm-up #4.
