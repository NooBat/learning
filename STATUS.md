# STATUS

**Current level:** L01 — mvs (complete 2026-04-20)
**Last updated:** 2026-04-20

## This-week focus

L01 shipped. Ready to transition to L02 — auth-tenancy.

## Next-session target

1. Invoke `/start-level 02 auth-tenancy` to scaffold the brief and flip `LEVELS.md`.
2. First L02 design conversation should bundle the carry-over items from L01 (see "Known gaps" below) — they are natural openers because L02 is already touching error translation (auth domain errors) and JSON shapes (tenant-scoped payloads).
3. Candidate opening tasks for L02:
   - Implement `PUT /workflows/{id}` and `DELETE /workflows/{id}` (deferred L01 exit criteria — mechanical extensions of the patterns already in place).
   - Translate Postgres SQLSTATE `22P02` (malformed UUID) at the storage boundary — choose between returning `ErrNotFound` (404) or a new `ErrInvalidID` (400); either is architecturally sound.
   - Write the first Go test suite — handler-level tests using a `fakeStore` that satisfies the consumer-defined `storage` interface (no real DB needed). ~20 lines of fake buys coverage of every 400/404/500 branch already written.

## Open blockers

- None. L01 code shipped and smoke-tested end-to-end; server binary builds clean; all positive + negative HTTP paths verified.

## In-flight ADRs

- None. ADR 0001 (router-choice → `net/http` stdlib) and ADR 0002 (postgres-driver → `pgx` native) both accepted during L01. Next ADR-worthy decision likely arrives when L02 chooses an auth model (JWT vs session vs OIDC delegation).

## Known gaps documented at L01 close

- **Malformed UUID → 500 (should be 400 or 404).** When `{id}` in `GET /workflows/{id}` is not a valid UUID, Postgres raises `SQLSTATE 22P02` (`invalid_text_representation`). The handler's error ladder only translates `ErrNotFound`; everything else falls to the catch-all 500. Architecturally clean fix lives at the **storage boundary** — translate 22P02 into a domain error (either reusing `ErrNotFound` for the 404 posture, or introducing `ErrInvalidID` for the 400 posture). Deferred to L02.

- **`PUT /workflows/{id}` and `DELETE /workflows/{id}` not implemented.** Originally in L01 exit criteria (`levels/L01-mvs.md`); scope-narrowed mid-session once it was clear the remaining architectural lessons (boundary translation, error ladder, consumer-defined interfaces, buffer-first encode, lifecycle orchestration) were all covered by `Create` / `GetByID` / `List`. Both would be mechanical extensions of existing patterns with no new learning payload. Pulled into L02 as a quick warm-up.

- **No automated tests yet.** L01 brief already deferred tests to L02. The consumer-defined `storage` interface (unexported, in `handler.go`) is the affordance that makes handler-level tests trivial — a fake implementation is ~20 lines and reaches every error branch.
