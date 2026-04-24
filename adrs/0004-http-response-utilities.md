# 0004. HTTP response utilities

**Status:** Accepted
**Date:** 2026-04-25
**Level:** L02 (auth-tenancy)

## Context

L01 handlers inline two patterns:

- **Success:** 5-line buffer encode (`bytes.Buffer` → `json.NewEncoder` → `Content-Type` → `WriteHeader` → `Write`). 3× duplication.
- **Error:** `http.Error(w, msg, status)` + `log.Printf`. ~7× duplication.

Fine at L01. L02 forces extraction:

- **Auth middleware** → 401 body must match handler 4xx/5xx body. Two paths, one shape.
- **ADR 0003** → malformed UUID, missing UUID, cross-tenant UUID = byte-identical 404. Shape drift leaks disclosure-opacity.
- **Test suite** → asserting inline bytes locks tests to impl. Structured envelope forward-compatible.
- **L03 `runs`** → envelope becomes cross-package contract. Pin shape pre-second-consumer = cheap. Post = coordinated rollout.

Decision scope = **boundary** + two guarantees:

1. **Envelope shape** — body structure.
2. **Status ownership** — handler picks or error carries.

Out of scope: middleware concerns (auth, logging, request-id).

## Options Considered

### Option A: Private helpers, same package

Unexported `writeJSON` / `writeError` in `project/internal/workflows/response.go`. Promote at L03 if pressure arrives.

- **Pros:**
  - Zero new boundary.
  - Stdlib "grow organically, extract under pressure" posture.
  - ~40 LOC relocated, handlers −15 lines each.
  - Per-package envelope divergence stays possible (e.g., future gRPC-gateway).
- **Cons:**
  - Defers envelope decision. L03 `runs` can silently diverge.
  - Invariant = convention, not compile-time.
  - ADR 0003 byte-identical-404 = review-time concern, not build-time.
  - Relitigate extraction at L03 + coordinate any drift.

### Option B: Shared `internal/httpx` package

`project/internal/httpx` exports `WriteJSON`, `WriteError`, `DecodeJSON`. ~60 LOC hand-rolled. Workflows now, runs at L03.

- **Pros:**
  - Pins envelope once, pre-drift. Convention → compile-time invariant.
  - ADR 0003 opacity = build-time guarantee (single impl → byte-identical bodies).
  - Tests cross-package, no same-package constraint.
  - Forces sub-decisions now when cost low.
  - No external dep.
- **Cons:**
  - New boundary pre-second-consumer. "Just in case" tax.
  - Scope-creep risk: `httpx` → middleware + CORS + parsing dumping ground. Discipline needed.
  - Domain → utility coupling axis. One-way.

### Option C: Responder / encoder framework

Handlers return `(any, error)`. Middleware encodes, inspects `HTTPStatus()` on error types. Errors HTTP-aware. Precedents: huma, chi+render, echo.

- **Pros:**
  - Handlers = pure business logic. No `ResponseWriter`.
  - Central seam: tracing, request-id, logging, CORS, RBAC.
  - Type-system envelope enforcement. Handlers cannot write wrong shape.
- **Cons:**
  - Big abstraction tax. Signature change + error HTTP-awareness + middleware ordering.
  - Premature at L02 (one package, ~100 LOC handlers, no cross-cuts). Payoff at 10+ packages.
  - Domain errors couple to HTTP. Non-HTTP consumers (gRPC, CLI, workers) carry baggage.
  - Hard to back out. C → A/B = handler-by-handler rewrite.
  - Framework lock-in risk if library adopted.

## Sub-decisions

Decision must name answers. Independent of boundary choice.

### 1. Envelope shape

- **Shape 1 — `text/plain` (status quo):** `"workflow not found\n"`.
  - Pro: stdlib default, zero design.
  - Con: not machine-parseable. Adding codes/trace_id = breaking change.
- **Shape 2 — JSON simple:** `{"error": "workflow not found"}`.
  - Pro: machine-parseable, matches success content-type.
  - Con: no room for structured fields without later break.
- **Shape 3 — JSON structured:** `{"error": {"code": "NOT_FOUND", "message": "..."}}`.
  - Pro: additive — `trace_id`, `details`, `fields` add later, clients ignore unknowns.
  - Con: over-engineered if never grown. Must design `code` semantics now (case, stability).

### 2. Status ownership

- **Ownership 1 — Handler picks:** `writeError(w, http.StatusNotFound, msg)`.
  - Pro: domain errors transport-agnostic. Reusable in gRPC, CLI, workers.
  - Con: consistency = convention at every call site.
- **Ownership 2 — Error carries:** `ErrNotFound` implements `HTTPStatus() int { return 404 }`.
  - Pro: mapping defined once at error type. No accidental miscategorization.
  - Con: domain errors HTTP-aware. Baggage for non-HTTP consumers.

Compose independently: any option × any shape × any ownership. Some combinations fight their own intent (e.g., Option C + Ownership 1 undoes C's point).

## Decision

**Option B (shared `internal/httpx`) + Shape 2 (JSON simple) + Ownership 1 (handler picks).**

**Tipping factor (B):** turns the envelope invariant from convention into a compile-time property before L03 `runs` ships. Single `httpx.WriteError` impl → byte-identical 404 bodies across packages by construction, so ADR 0003's opacity guarantee becomes a build-time property, not a review-time one.

**Shape 2 rationale:** JSON envelope (`{"error": "..."}`) unblocks two concrete L02+ needs — frontend error rendering (parse once, no content-type sniffing) and log-aggregator filtering (structured JSON is machine-matchable; `text/plain` is grep-only). Cheaper than Shape 3's `code`/`message` structure — deferred until real need appears (client code-branching on `code`, trace injection, structured field validation).

**Ownership 1 rationale:** domain errors stay transport-agnostic. `ErrNotFound` / `ErrInvalidInput` have zero HTTP knowledge. When the service grows non-HTTP surfaces (CLI, background worker, eventual gRPC), errors don't drag HTTP baggage into contexts where status codes are meaningless.

**Accepted cons:**

- **B:** new package boundary pre-second-consumer — "just in case" abstraction tax. Scope-creep risk (`httpx` as a dumping ground for CORS, middleware, pagination). Mitigated by keeping scope explicit in code comments: encode / decode / error envelope only. Anything else = separate package.
- **Shape 2:** no room for `code` / `trace_id` / `details` without a later migration. Shape 2 → Shape 3 is breaking if clients read `error` as a string. Free at L02 pre-deploy; worth re-confirming before L04 first deploy.
- **Ownership 1:** consistency by convention at every call site — every `httpx.WriteError` call must pair the right status with the right error. Enforced at review + test time, not by types.

**Composition with ADR 0003:** single `httpx.WriteError` implementation guarantees byte-identical 404 body across malformed-UUID (translated at storage), missing-UUID, and (later) cross-tenant-UUID paths. ADR 0003's server-transparency requirement is unaffected — logging stays at handler/storage boundary; `httpx` does not own logging.

## Consequences

**New package.** `project/internal/httpx` with three exports:

- `WriteJSON(w http.ResponseWriter, status int, body any) error` — buffer-first encode then flush (header + body written atomically).
- `WriteError(w http.ResponseWriter, status int, msg string)` — writes `{"error": msg}` with `Content-Type: application/json`.
- `DecodeJSON(r *http.Request, dst any) error` — wraps `json.NewDecoder(r.Body).Decode(dst)`. Handler decides the 400 message on failure.

No new external dependency. Package stays ~60 LOC. Scope explicitly bounded in package doc comment: encode / decode / error envelope only. Middleware, CORS, pagination, request parsing are **out of scope** for `httpx` — separate packages if/when they arrive.

**Handler changes.** Every handler (`create`, `getByID`, `list`) collapses:

- Buffer-encode dance (5 lines × 3 sites) → `httpx.WriteJSON(w, status, body)` (1 line per site).
- `http.Error(...)` calls (text/plain) → `httpx.WriteError(w, status, msg)` (JSON envelope). ~7 call sites across the three handlers.
- `json.NewDecoder(r.Body).Decode(&workflow)` → `httpx.DecodeJSON(r, &workflow)`.

Error ladder structure unchanged — same `errors.Is` branches, same status choices per branch. Handlers shrink by ~15 lines each.

**Server-side logging.** Stays at handler/storage boundary as today. `httpx` does not own logging — keeping logging where causes are visible preserves ADR 0003's server-transparency (storage logs SQLSTATE before returning `ErrNotFound`; handler logs its own decode/encode errors). Side benefit: JSON envelope makes error responses machine-filterable by log aggregators (`jq '.error'` on response bodies; `grep` was the only prior option).

**Reversibility per choice.**

- **Boundary (B → A):** cheap. Collapse `httpx` back into per-package private helpers, update imports. ~30 minutes mechanical work.
- **Envelope (Shape 2 → Shape 3):** **sticky** once external clients exist. Additive *only* if clients read `error` as an arbitrary JSON value; most clients read it as a string → migration becomes a breaking change. Free at L02 pre-deploy. Re-confirm before L04 first deploy.
- **Ownership (1 → 2):** moderate. Requires teaching every domain error type `HTTPStatus()` + updating every `httpx.WriteError` call site. ~1–2 hours at L02 scale; linearly more per call site after L03+.

**Future ADR locks/opens.**

- **`auth-model`:** 401 body inherits Shape 2 by construction. Middleware calls `httpx.WriteError(w, http.StatusUnauthorized, "invalid token")` — same envelope as handler 4xx. Satisfies the "two paths, one envelope" requirement named in Context.
- **`tenancy-isolation`:** inherits ADR 0003 opacity + this ADR's byte-identical guarantee. Cross-tenant 404 body is `{"error": "workflow not found"}`, indistinguishable from missing/malformed. No new architectural question — same posture applied to a new layer.
- **`test-strategy`:** tests assert against parsed JSON, not byte-compared plaintext. Establishes a test helper that JSON-decodes response bodies from day one — forward-compatible to a future Shape 3 migration. **Recommendation:** write that helper alongside the first handler test.
- **API versioning (L04+):** Shape 2 becomes part of the de facto v1 contract as soon as a client ships. A future Shape 3 migration = versioned breaking change (content-type negotiation `application/vnd.flux.v2+json` or path-versioned endpoint). Worth naming explicitly at L04's first-deploy ADR.
