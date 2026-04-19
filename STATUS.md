# STATUS

**Current level:** L02 — auth-tenancy (in-progress, started 2026-04-20)
**Last updated:** 2026-04-20

## This-week focus

L02 opens. The first session bundles the L01 carry-overs as warm-ups before the auth model itself:

1. Draft ADR `malformed-uuid-translation` — 404 vs 400 posture. Recommendation is 404 (aligns with the tenancy presence-disclosure stance arriving later this level).
2. Implement `PUT /workflows/{id}` + `DELETE /workflows/{id}` — mechanical extension of L01 patterns (storage interface, handler method, route registration).
3. Translate Postgres `SQLSTATE 22P02` at the storage boundary per the ADR.
4. Write the first Go test suite using `httptest.NewServer` + a `fakeStore` satisfying the unexported `storage` interface. Target every error-ladder branch.
5. Write `project/docs/architecture.md` stub documenting the L01 layering.

Once warm-ups land, start L02 proper: auth middleware + tenancy column + ADRs for auth model, tenancy isolation, and test strategy.

## Next-session target

1. Read `levels/L02-auth-tenancy.md` end-to-end; refine exit criteria if anything feels under- or over-scoped for the level.
2. Invoke `/write-adr malformed-uuid-translation` and commit the decision before touching code. The 404-vs-400 choice sets a presence-disclosure posture that the tenancy work will inherit — deciding this first avoids retrofitting later.
3. Start warm-up #1 (PUT/DELETE). Daniel writes the code; Claude coaches via Learn-by-Doing + staff-engineer review.

## Open blockers

- None. L01 shipped clean; L02 brief scaffolded and ready for Daniel's review.

## In-flight ADRs

- None yet. L02 queues four: `malformed-uuid-translation`, `auth-model`, `tenancy-isolation`, `test-strategy`. First one lands before any L02 code does.

## Known gaps carried from L01

- **Malformed UUID → 500** (currently). Postgres raises `SQLSTATE 22P02` when `{id}` isn't a valid UUID; the handler's error ladder only translates `ErrNotFound`, so everything else falls to the catch-all 500. Fix lives at the storage boundary — ADR `malformed-uuid-translation` chooses between reusing `ErrNotFound` (404) or introducing `ErrInvalidID` (400).

- **`PUT /workflows/{id}` and `DELETE /workflows/{id}` not implemented.** Originally L01 exit criteria; scope-narrowed mid-session once it was clear the architectural lessons (boundary translation, error ladder, consumer-defined interfaces, buffer-first encode, lifecycle orchestration) were covered by `Create` / `GetByID` / `List`. Both are mechanical extensions of existing patterns with no new learning payload — pulled into L02 as quick warm-ups.

- **No automated tests yet.** L01 brief deferred tests to L02. The unexported consumer-defined `storage` interface in `handler.go` is the affordance that makes handler-level tests trivial — a `fakeStore` is ~20 lines and reaches every 400/404/500 branch.

- **`project/docs/architecture.md` does not exist.** Last deferred L01 exit criterion. Drafted at L02 open so that L02's auth + tenancy additions land on top of a documented L01 baseline.
