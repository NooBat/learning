# STATUS

**Current level:** L02 — auth-tenancy (in-progress, started 2026-04-20)
**Last updated:** 2026-05-10

## This-week focus

L02 warm-ups arc **closed**: integration ring shipped + architecture posture-doc + ADR catalog. L02 proper next — `auth-model` and `tenancy-isolation` ADRs to draft, then implementation.

**Done (session arc 04-21 → 05-10):**

1. ✅ ADR 0003 (`malformed-uuid-translation`) accepted — 22P02 → `ErrNotFound` (404). Storage-boundary translation; ADR 0001-style presence-disclosure posture pinned for the level.
2. ✅ ADR 0004 (`http-response-utilities`) accepted — shared `internal/httpx` package + Shape 2 JSON envelope (`{"error": "..."}`) + Ownership 1 (handler picks status). Compile-time invariant for byte-identical 404 across packages.
3. ✅ `internal/httpx` package shipped (`WriteJSON`, `WriteError`, `DecodeJSON`). Handler refactored — buffer-encode dance + `http.Error` calls collapsed.
4. ✅ Request body cap at handler edge — `http.MaxBytesReader(w, r.Body, 1*MiB)` + `*http.MaxBytesError` branch returning 413 (option B placement).
5. ✅ ADR 0005 (`workflow-lifecycle-ops`) accepted — PUT strict / DELETE idempotent 204 / soft delete via `deleted_at`. Pairing B chosen: extends ADRs 0003 + 0004 opacity stance to a third surface; three ADRs now share the posture.
6. ✅ PUT/DELETE handlers shipped. Schema: `deleted_at timestamptz` inlined. Storage: `Update`/`Delete` + soft-delete filter on `GetByID`/`List`. `translatePgError` helper consolidates ADR 0003's 22P02 → ErrNotFound rule across all storage methods.
7. ✅ Smoke: 13 cases verified live. Pairing B opacity contract holds — DELETE on existing/just-deleted/never-existed/malformed-UUID all return byte-identical 204. PUT 404 collapses missing/soft-deleted/malformed-UUID into one response.
8. ✅ ADR 0006 (`test-strategy`) accepted — Path 1C + 2C + 3A: handler-level fakeStore primary + integration ring (`//go:build integration`) + envelope-decode helper centralizes ADR 0004's Shape 2 contract + minimal map-backed fakeStore (no fault injection at L02). Cashes the L01 `storage`-interface design payoff; locks ADR 0005 opacity invariant as a regression test.
9. ✅ Test suite scaffolded — `internal/workflows/handler_test.go` (fakeStore + helpers: `errorMessage`, `requireStatus`, `decodeWorkflow`, `decodeWorkflowList`, `postJSON`, `putJSON`, `seedWorkflow`, `newTestServer`, `readBody`) + `internal/workflows/storage_integration_test.go` (build-tagged).
10. ✅ `TestDelete_OpacityInvariant` shipped — promotes the four-DELETE-paths byte-identical wire assertion from one-time smoke to compile-time regression test. Encodes ADRs 0003 + 0004 + 0005 in one assertion.
11. ✅ Handler branch coverage: 22 tests green covering 200 / 201 / 204 / 400 / 404 / 413. Create / Update / List / GetByID / Delete each with success + error paths. Soft-delete invisibility asserted at every read site (List + GetByID + Update). 500 path deferred per ADR 0006 cons-accepted (no fault injection at L02).
12. ✅ **Integration ring shipped** (warm-up #3 tail). `flux_test` DB provisioned (Option A — separate physical DB, mirrors CI shape; setup-guide step 8 added). 3 tests in `storage_integration_test.go` (build-tag `//go:build integration`):
    - `TestStorage_Update_RETURNING` — id + created_at immutable; updated_at server-bumped via `now()`. 2ms sleep clears Postgres microsecond-resolution race.
    - `TestStorage_GetByID_22P02_to_NotFound` — ADR 0003 end-to-end against live driver. Contract-only via `errors.Is(err, ErrNotFound)` (the only viable assertion — `translatePgError` severs the wrap chain by design, making `errors.As` against `*pgconn.PgError` non-viable).
    - `TestStorage_Delete_SoftDeleteFilter` — soft-deleted row excluded from both GetByID + List.
    - Build-tag discipline verified: default `go test ./...` skips ring (22 handler tests); `-tags=integration` runs all 25.
13. ✅ **Architecture posture-doc + ADR catalog shipped** (warm-up #4). `project/docs/architecture.md` ~33 lines: opacity-on-wire (ADRs 0003+0004+0005) + stdlib-purism (ADRs 0001+0002+0006) — the only artifact in the repo capturing how those ADR clusters share one design assertion. Trimmed hard from a 7-section draft after honest scrutiny: 6 sections duplicated STATUS / project-tree / ADRs / would-be-better-as-`adrs/README.md`. Invariants section considered + dropped (each candidate already lives in source ADR's Decision section). `adrs/README.md` created as the conventional ADR catalog location.

**Queued (L02 proper):**

1. **Draft `auth-model` ADR.** Authentication mechanism choice — JWT (stateless, scaling) vs session-cookie (revokable, simpler) vs API-key (machine-to-machine) vs basic-auth (dev-only). Compatibility constraint: presence-disclosure (404 opacity) must hold post-auth — the 401 challenge needs to compose with existing 404 behavior. Stdlib-purism posture in play (no auth library yet; ADRs 0001/0002/0006 baseline says frameworks earn their slot).
2. **Draft `tenancy-isolation` ADR.** Schema column (`tenant_id` on workflows) + WHERE-clause discipline at every query site + cross-tenant access returns 404 (not 403, opacity preserved). Composes with auth ADR — auth identifies the tenant; tenancy enforces isolation.
3. **Implementation after both ADRs accepted.** Order shaped by ADRs: middleware likely first (puts tenant_id in context), then column migration + WHERE sweep across storage methods. Test pyramid extends — fakeStore tracks tenant state; integration ring gets one tenancy-isolation test.

## Next-session target

1. **Open `auth-model` ADR.** Forces converging: mechanism choice across ~4 candidates; auth-vs-authz boundary; dependency posture (stdlib-only vs introduce a library); deployment model (Fly.io at L04 implies HTTPS-required posture and cookie domain decisions). Decision space is large — the ADR's Options Considered section will be the load-bearing one.
2. **Then `tenancy-isolation`.** Decisions stack: auth ADR's tenant-id-on-context output is the input to tenancy WHERE-clause discipline.
3. **Implementation after both ADRs accepted.**

## Open blockers

- None. Branch ahead of `origin/main` by 2 commits this session: `3e5d125` (warm-up #3 tail — integration ring) + `a1cf233` (warm-up #4 — architecture posture doc + ADR catalog). Push when ready.

## In-flight ADRs

- ADR 0001-0006 all ✅ accepted. Catalog at `adrs/README.md`.
- ADR queue (L02 proper): `auth-model`, `tenancy-isolation`.

## Known gaps carried from L01

- ~~**Malformed UUID → 500.**~~ ✅ Fixed by ADR 0003 implementation.
- ~~**`PUT /workflows/{id}` and `DELETE /workflows/{id}` not implemented.**~~ ✅ Shipped via ADR 0005.
- ~~**No automated tests yet.**~~ ✅ Handler suite (22 tests) + integration ring (3 tests) green; runs gated by build tag.
- ~~**`project/docs/architecture.md` does not exist.**~~ ✅ Shipped (warm-up #4) — posture-only synthesis.

(No remaining gaps from L01.)
