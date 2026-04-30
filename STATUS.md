# STATUS

**Current level:** L02 — auth-tenancy (in-progress, started 2026-04-20)
**Last updated:** 2026-04-30

## This-week focus

L02 warm-ups in progress. Each one earns a structural lesson before the auth model proper.

**Done (this session arc, 04-21 → 04-30):**

1. ✅ ADR 0003 (`malformed-uuid-translation`) accepted — 22P02 → `ErrNotFound` (404). Storage-boundary translation; ADR 0001-style presence-disclosure posture pinned for the level.
2. ✅ ADR 0004 (`http-response-utilities`) accepted — shared `internal/httpx` package + Shape 2 JSON envelope (`{"error": "..."}`) + Ownership 1 (handler picks status). Compile-time invariant for byte-identical 404 across packages, ahead of L03 `runs`.
3. ✅ `internal/httpx` package shipped (`WriteJSON`, `WriteError`, `DecodeJSON`). Handler refactored — buffer-encode dance + `http.Error` calls collapsed.
4. ✅ Request body cap at handler edge — `http.MaxBytesReader(w, r.Body, 1*MiB)` + `*http.MaxBytesError` branch returning 413 (option B placement: per-endpoint, helper stays JSON-only).

**Queued (in order):**

1. **ADR 0005 (`workflow-lifecycle-ops`) + PUT/DELETE handlers.** Three stacked decisions: PUT semantics (strict vs upsert), DELETE existence (strict 404 vs idempotent 204), hard vs soft delete. Two natural pairings: (hard, strict) for L02 simplicity, (soft, idempotent) to extend ADR 0003 opacity + retention/audit. ADR drives the `storage` interface signatures, so write the ADR first.
2. **First Go test suite.** `httptest.NewServer` + `fakeStore` satisfying the unexported `storage` interface. Cover every 400/404/413/500 branch; one real-DB integration test for the golden path.
3. **`project/docs/architecture.md` stub.** Document the L01 baseline + the L02 additions (httpx, ADR 0003 boundary translation) before auth/tenancy land on top.

After warm-ups: L02 proper — auth middleware + tenancy column + ADRs for auth model, tenancy isolation, test strategy.

## Next-session target

1. Decide: draft ADR 0005 covering all three lifecycle decisions in one doc, or split into smaller ADRs. Recommendation: one ADR (decisions compose, splitting fragments the trade-off space).
2. If one ADR: invoke `/write-adr workflow-lifecycle-ops`. Frame the three decisions and the two natural pairings; pick before touching `storage.go`.
3. After ADR: extend `storage` interface with `Update` + `Delete` (signatures driven by the ADR), then handler methods + route registration.

## Open blockers

- None. Branch is ahead of `origin/main` by 4 local commits (ADR 0003 + ADR 0004 + httpx integration + body cap). Push when ready.

## In-flight ADRs

- ADR 0003 ✅ accepted (`adrs/0003-malformed-uuid-translation.md`).
- ADR 0004 ✅ accepted (`adrs/0004-http-response-utilities.md`).
- ADR 0005 (`workflow-lifecycle-ops`) — queued, drafting next.
- ADR queue beyond 0005: `auth-model`, `tenancy-isolation`, `test-strategy`.

## Known gaps carried from L01

- ~~**Malformed UUID → 500.**~~ ✅ Fixed by ADR 0003 implementation.

- **`PUT /workflows/{id}` and `DELETE /workflows/{id}` not implemented.** Next up. ADR 0005 first.

- **No automated tests yet.** Still queued. The unexported consumer-defined `storage` interface in `handler.go` is the seed affordance.

- **`project/docs/architecture.md` does not exist.** Still queued. Will document L01 + httpx + ADR 0003 baseline before L02 auth/tenancy land on top.
