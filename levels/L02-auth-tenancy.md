# L02 — Auth & Tenancy

**Status:** in-progress
**Tier:** 1 — Backend Engineer
**Started:** 2026-04-20

## Goal

Layer identity and multi-tenancy onto the L01 workflows service, and plant the first Go test suite. Every workflow is owned by a tenant; every request is authenticated; unauthenticated or cross-tenant reads are rejected with deliberate presence-disclosure discipline.

## Why this level exists

L02 is the "who are you, and what are you allowed to see?" level. It introduces the first real **security boundary** (authentication — who can talk to the service) and the first real **data boundary** (tenancy scoping — what a caller can see within it). These are two architectural concepts most backend systems eventually need, and getting them wrong early is expensive to fix later — both tend to ossify into load-bearing assumptions within weeks.

L02 also establishes the **testing posture**. L01 shipped without tests intentionally; the consumer-defined `storage` interface in `handler.go` was the seed. L02 plants the tree: `httptest.NewServer` + a `fakeStore` satisfying that interface, covering every error-ladder branch. Integration tests arrive alongside, but narrowly — golden paths only, real DB, transactional cleanup.

**The architectural lesson:** *isolation is a layered posture, not a feature.* Who can talk to the service (auth), what they can see within it (tenancy scope), what shape of error surfaces on violation — each is a separate decision with its own ADR, and the choices compose into the service's security stance.

## Prerequisites

- L01 shipped and closed (`levels/L01-mvs.md`)
- Basic familiarity with HTTP auth semantics (Authorization header, 401 vs 403)
- A working model of SQL-level isolation options — row-level filtering vs Postgres RLS vs schema-per-tenant
- The L01 carry-overs below are natural warm-ups; they open the session before auth work begins

## Warm-up tasks (L01 carry-overs)

Quick wins before diving into the auth model itself. Each also exercises a specific architectural muscle.

1. **Implement `PUT /workflows/{id}` and `DELETE /workflows/{id}`.** Extend the unexported `storage` interface in `handler.go`, add two storage methods, two handler methods, two route registrations. Mechanical application of L01 patterns — but worth noticing that the error ladder for `PUT` is identical-shape to `Create` plus a 404 branch from `GetByID`. That compositional structure is the point.
2. **Translate Postgres `SQLSTATE 22P02` (malformed UUID) at the storage boundary.** `GET /workflows/<not-a-uuid>` currently returns 500; clean fix lives in `storage.go`, not in handlers. **ADR-worthy:** map to `ErrNotFound` (404 posture — presence-opaque) or introduce `ErrInvalidID` (400 posture — shape-precise)? Both are defensible. **My bias: 404.** Reason: tenancy is about to land, and cross-tenant workflow lookups will also want to return 404 regardless of whether the row exists. Choosing 404 here lets one posture rule the whole service. Write the ADR before the code.
3. **Write the first Go test suite.** Handler-level tests using `httptest.NewServer` + a `fakeStore` implementing the unexported `storage` interface. Target every error-ladder branch across `Create` / `GetByID` / `List` / `PUT` / `DELETE`. No real database. The interface is already the right shape; the fake is ~20 lines.
4. **Write `project/docs/architecture.md`** — capture the L01 layering (handler ↔ storage boundary, consumer-defined interfaces, validation-at-the-edge, buffer-first encode, lifecycle orchestration). Last deferred L01 exit criterion.

## Exit criteria

All checkable. All must be true before claiming L02 done. Daniel edits this list before starting — scope can tighten but should not quietly expand.

### Carry-overs from L01

- [ ] `PUT /workflows/{id}` updates a workflow (200) or 404
- [ ] `DELETE /workflows/{id}` deletes a workflow (204) or 404
- [ ] Malformed UUID returns a non-500 status (400 or 404 per ADR decision)
- [ ] `project/docs/architecture.md` exists and describes the L01 layering

### Testing (the first real test suite)

- [ ] `go test ./...` runs green
- [ ] Handler-level tests using `httptest.NewServer` + `fakeStore` cover every branch of `Create` / `GetByID` / `List` / `PUT` / `DELETE`
- [ ] Test coverage is measured (`go test -cover`) — the number isn't the target; visibility is
- [ ] At least one integration test hits real Postgres via pgx (golden path: create → get → list → update → delete) with transactional fixture cleanup
- [ ] ADR written for the test strategy (where each kind of confidence lives)

### Auth (identity layer)

- [ ] Requests without valid credentials return 401
- [ ] Auth model chosen and documented in an ADR (opaque bearer token / JWT / session / OIDC delegation)
- [ ] Credentials verified at a middleware layer, not inside handlers (boundary discipline — handlers never touch raw credentials)
- [ ] Auth errors disclose nothing about why credentials were rejected (no "user not found" vs "wrong password" leakage — both return the same opaque 401)
- [ ] Caller identity attached to `r.Context()` and retrieved via a typed accessor, not ad-hoc context keys scattered across handlers

### Tenancy (data boundary)

- [ ] `workflows.tenant_id` column exists (NOT NULL, foreign key to `tenants.id`)
- [ ] Minimal `tenants` table exists (id, name, timestamps) and `users` table associates users with tenants
- [ ] Tenancy isolation model chosen and documented in an ADR (application-level filtering / Postgres RLS / schema-per-tenant)
- [ ] Cross-tenant reads return 404 (not 403) — presence-disclosure discipline, aligned with the 22P02 decision above
- [ ] `GET /workflows` returns only the caller's tenant's workflows
- [ ] `GET /workflows/{id}` returns 404 if the workflow exists but belongs to a different tenant
- [ ] `POST /workflows` stamps `tenant_id` from the authenticated identity, not from the request body (never trust client-supplied tenancy)

### Architecture & documentation

- [ ] ADRs written for: malformed-UUID translation, auth model, tenancy isolation model, test strategy
- [ ] `project/docs/architecture.md` updated with the L02 shape (auth middleware, tenancy column, test pyramid posture)

## Scope

### In-scope

- **Auth middleware:** a single HTTP middleware that extracts credentials, validates them, attaches a caller identity to `r.Context()`, short-circuits with 401 on invalid.
- **Tenancy column:** `workflows.tenant_id NOT NULL`, foreign-keyed to a minimal `tenants` table (id + name + timestamps).
- **Identity model:** a minimal `users` table with `tenant_id`. Users authenticate; their tenant is derived from the identity, never the request body.
- **Test harness:** `httptest` + `fakeStore` for handlers; a separate integration test using the real Postgres pool with transactional cleanup.

### Out of scope (deliberately deferred)

- OAuth / external identity providers (Google, GitHub SSO) → L04+ (when real deployment motivates it)
- Role-based access control within a tenant (admin vs member vs viewer) → L03+
- API keys / service-to-service auth → L05 (arrives naturally with background workers)
- Tenant self-service signup UI → never at this tier (backend only)
- Rate limiting per tenant → L07 (resilience)
- Multi-factor auth → out-of-scope for this tier entirely
- Password reset flows → manual DB seed is fine for L02
- Docker / containers → L03
- Migration tool (goose/golang-migrate) → L03
- Structured logging (`log/slog`) → L03

Resist adding these. Each arrives at the level that teaches it.

## ADR-worthy decisions likely to come up

Pause and invoke `/write-adr <slug>` before implementing each:

1. **Malformed-UUID error posture** (`malformed-uuid-translation`) — map Postgres `22P02` to `ErrNotFound` (404) or introduce `ErrInvalidID` (400)? Trade-off framing: presence-disclosure discipline vs input-shape precision. **My bias: 404** — aligns with the tenancy presence-disclosure posture arriving in the same level.
2. **Auth model** (`auth-model`) — opaque bearer token (server-side table, explicit revocation, no JWT footguns) vs JWT (stateless, self-contained, more ecosystem) vs session cookie (stateful, server-tracked, browser-native) vs OIDC delegation (outsourced identity). **My bias: opaque bearer token with a server-side `tokens` table.** Simpler primitives, explicit revocation, and defers the JWT footgun landscape (algorithm confusion, clock skew, key rotation) to a later level when service-to-service need actually motivates it. JWT arrives naturally around L05.
3. **Tenancy isolation model** (`tenancy-isolation`) — application-level filtering (`WHERE tenant_id = $1` threaded through every query) vs Postgres RLS policies (DB enforces) vs schema-per-tenant (DDL-level separation) vs database-per-tenant (fully isolated). **My bias: application-level filtering** at L02, with RLS documented as the future migration path if cross-cutting auditability or defense-in-depth concerns emerge. Cheap to start, explicit in code, trivially testable.
4. **Test strategy / pyramid shape** (`test-strategy`) — where does each kind of confidence live? Handler-unit (fake storage, fast, branch-coverage) / integration (real DB, transactional, golden-path) / smoke (real HTTP client, build-time only). Document the pyramid so L03+ doesn't drift into the wrong layer (common failure: everything becomes a slow integration test because they "feel more real").

## Reading triggered by this level

- ★ *Learning Go* by Jon Bodner — Chapter 11 (Testing) thoroughly; Chapter 12 (Context) revisit now that real middleware is landing.
- [Go testing docs](https://pkg.go.dev/testing) — `*testing.T`, table-driven tests, `httptest` subpackage.
- [OWASP ASVS v4](https://owasp.org/www-project-application-security-verification-standard/) — skim authentication section (V2); pick 3-5 requirements to verify against the implementation.
- *Release It!* by Michael Nygard — Chapter 5 (Stability Patterns); starting to orient toward failure modes.
- If choosing JWT: [RFC 7519](https://datatracker.ietf.org/doc/html/rfc7519) + [JWT best practices RFC 8725](https://datatracker.ietf.org/doc/html/rfc8725). Read before implementing — the footguns are plural and non-obvious.
- If choosing opaque tokens: think through rotation and revocation before writing the code; both are trivial design decisions if made up-front, gnarly retrofits if not.

## Stretch (only if flying)

- Add `X-Request-ID` middleware (propagate if set, generate UUID if not) — sets up observability work at L08.
- Write a Go benchmark (`testing.B`) for the `List` endpoint — introduces benchmarking before L06 needs it.
- Add a second tenant in the integration test and verify isolation end-to-end (a "tenancy smoke test") rather than relying on unit coverage alone.

## Anti-scope (do NOT do)

- Don't add RBAC. Later.
- Don't add OAuth providers. Later.
- Don't add Docker. L03.
- Don't introduce JWT just because "everyone uses JWT" — that's a decision, not a default. Write the ADR first.
- Don't route every assertion through an integration test. Handler-level tests with `fakeStore` catch most branches faster and cheaper.
- Don't let the test suite drift into the "only integration tests" failure mode. The pyramid exists for a reason.

## When you think you're done

1. Walk through each exit criterion. Honest checkoff.
2. Run `go test ./...` — all green.
3. Run `go test -cover ./...` — note the number even if unimpressed.
4. Smoke test: create tenant A, create tenant B, authenticate as A, confirm A can't see B's workflows (404, not 403, not 500).
5. Update `project/docs/architecture.md` with the L02 shape (auth middleware layer, tenancy column + scoping posture, test pyramid).
6. Commit + push.
7. Invoke `/start-level 03 production-local` to scaffold the next level.

If any criterion isn't met, don't transition. Finish it.
