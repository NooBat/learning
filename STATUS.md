# STATUS

**Current level:** L02 — auth-tenancy (in-progress, started 2026-04-20)
**Last updated:** 2026-05-01

## This-week focus

L02 warm-ups: 2 of 3 done.

**Done (session arc 04-21 → 05-01):**

1. ✅ ADR 0003 (`malformed-uuid-translation`) accepted — 22P02 → `ErrNotFound` (404). Storage-boundary translation; ADR 0001-style presence-disclosure posture pinned for the level.
2. ✅ ADR 0004 (`http-response-utilities`) accepted — shared `internal/httpx` package + Shape 2 JSON envelope (`{"error": "..."}`) + Ownership 1 (handler picks status). Compile-time invariant for byte-identical 404 across packages.
3. ✅ `internal/httpx` package shipped (`WriteJSON`, `WriteError`, `DecodeJSON`). Handler refactored — buffer-encode dance + `http.Error` calls collapsed.
4. ✅ Request body cap at handler edge — `http.MaxBytesReader(w, r.Body, 1*MiB)` + `*http.MaxBytesError` branch returning 413 (option B placement).
5. ✅ ADR 0005 (`workflow-lifecycle-ops`) accepted — PUT strict / DELETE idempotent 204 / soft delete via `deleted_at`. Pairing B chosen: extends ADRs 0003 + 0004 opacity stance to a third surface; three ADRs now share the posture.
6. ✅ PUT/DELETE handlers shipped. Schema: `deleted_at timestamptz` inlined. Storage: `Update`/`Delete` + soft-delete filter on `GetByID`/`List`. `translatePgError` helper consolidates ADR 0003's 22P02 → ErrNotFound rule across all storage methods.
7. ✅ Smoke: 13 cases verified live. Pairing B opacity contract holds — DELETE on existing/just-deleted/never-existed/malformed-UUID all return byte-identical 204. PUT 404 collapses missing/soft-deleted/malformed-UUID into one response.

**Queued (in order):**

1. **First Go test suite (warm-up #3).** `httptest.NewServer` + `fakeStore` satisfying the unexported `storage` interface. Cover every 200/201/204/400/404/413/500 branch at the handler level. One real-DB integration test for the golden path. The pairing B opacity invariants from ADR 0005 are the highest-value test targets — the four DELETE paths must produce identical responses by automated assertion, not just live smoke.
2. **`project/docs/architecture.md` stub (warm-up #4).** Document the L01 baseline + L02 additions (httpx, ADR 0003 boundary translation, soft-delete pattern) before auth/tenancy land on top.

After warm-ups: L02 proper — auth middleware + tenancy column + ADRs for auth model, tenancy isolation, test strategy.

## Next-session target

1. Test suite skeleton: `internal/workflows/handler_test.go` with `fakeStore` implementation. Goal: every error-ladder branch covered + opacity-invariant assertion on the four DELETE paths.
2. Optional: `internal/workflows/storage_integration_test.go` — one golden-path integration test against real Postgres for the soft-delete behavior (live smoke verified it; need a regression test).
3. Choose: write a generic test helper for parsing `httpx` envelope responses (forward-compatible to ADR 0004's noted Shape 2 → Shape 3 migration), or assert on inline JSON bytes initially.

## Open blockers

- None. Branch is ahead of `origin/main` by 7 local commits. Push when ready.

## In-flight ADRs

- ADR 0003 ✅ accepted (`adrs/0003-malformed-uuid-translation.md`).
- ADR 0004 ✅ accepted (`adrs/0004-http-response-utilities.md`).
- ADR 0005 ✅ accepted (`adrs/0005-workflow-lifecycle-ops.md`).
- ADR queue: `test-strategy` (warm-up #3 motivates writing this — testing pyramid + envelope-decode helper). Then `auth-model`, `tenancy-isolation`.

## Known gaps carried from L01

- ~~**Malformed UUID → 500.**~~ ✅ Fixed by ADR 0003 implementation.
- ~~**`PUT /workflows/{id}` and `DELETE /workflows/{id}` not implemented.**~~ ✅ Shipped via ADR 0005.
- **No automated tests yet.** Next up — warm-up #3.
- **`project/docs/architecture.md` does not exist.** After tests — warm-up #4.
