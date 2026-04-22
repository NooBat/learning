# 0003. Malformed UUID translation

**Status:** Accepted
**Date:** 2026-04-21
**Level:** L02 (auth-tenancy)

## Context

L01 shipped `GET /workflows/{id}` where the handler passes `{id}` straight to `storage.GetByID(ctx, id string)`, which binds it as `$1` in `SELECT ... FROM workflows WHERE id = $1`. The `workflows.id` column is typed `uuid`.

When `{id}` is not a valid UUID (e.g., `GET /workflows/not-a-uuid`, `GET /workflows/abc123`), Postgres rejects the query at the wire protocol level with `SQLSTATE 22P02` (`invalid_text_representation`) — the bind parameter can't be coerced to `uuid`. The storage method returns this error unwrapped; the handler's error ladder only translates `ErrNotFound`, so `22P02` falls through to the catch-all `500 internal server error`.

L01 closed with this as a documented architectural gap. The fix lives at the storage boundary — same layer that already translates `pgx.ErrNoRows → ErrNotFound`. Handlers should not know Postgres SQLSTATE codes; the storage boundary is the single place where driver concerns are laundered into domain concerns.

**Why this decision is load-bearing for L02:** tenancy (arriving later this level) also wants cross-tenant reads to return 404 — a single presence-disclosure posture where the client cannot distinguish *"doesn't exist"* from *"not yours"*. The malformed-UUID decision either harmonizes with that posture (one 404-for-everything-unreadable stance) or introduces a second client-observable contract the tenancy work has to coexist with. Choosing here first lets the tenancy ADR inherit the stance rather than negotiate it.

The choice also touches what gets logged server-side: regardless of the client-observable status, the server-side log should always distinguish *malformed input* from *legitimately not-found* for debugging. The ADR is only about the client contract.

## Options Considered

### Option A: Map `22P02 → ErrNotFound` (404 posture)

Storage layer catches `22P02` via `pgconn.PgError.Code` and returns `ErrNotFound`. Handler error ladder stays two-branch (`ErrNotFound` + catch-all). Client sees 404 for malformed-id, valid-id-not-found, and (later) valid-id-cross-tenant — indistinguishable.

- **Pros:**
  - Single presence-disclosure posture across the whole service. Malformed id, valid id that doesn't exist, valid id owned by another tenant — all 404. Client cannot enumerate, probe, or distinguish.
  - Harmonizes with the tenancy 404-over-403 stance arriving later in L02. One ADR sets the service-wide posture; the tenancy ADR inherits it.
  - No new domain-error sentinel needed. Handler error ladder stays at two branches (ErrNotFound + catch-all 500). Minimum surface-area growth.
  - Attackers cannot use status-code differences to probe the ID space — malformed and shape-valid-but-absent look identical.
- **Cons:**
  - Hides a distinction a well-behaved client *could* use (e.g., a CLI wanting to print "that's not a valid UUID" vs "no workflow with that ID"). In practice rarely used; server-side log can still distinguish for debugging.
  - Some REST-purist intuitions want `400 Bad Request` for shape errors and `404 Not Found` reserved for presence errors. This option deliberately steps away from that orthodoxy in favor of disclosure-opacity.

### Option B: Introduce `ErrInvalidID` → 400 posture

Storage layer catches `22P02` and returns a new sentinel `ErrInvalidID`. Handler error ladder gains a third branch that maps it to `400 Bad Request`. Clients see 400 for malformed id, 404 for shape-valid-but-absent, and (later) 404 for cross-tenant.

- **Pros:**
  - Explicit, two-contract semantics: `400` means "shape wrong", `404` means "shape right, nothing there". Debuggers and typed clients can distinguish.
  - Aligns with REST orthodoxy (shape errors at 400; presence errors at 404).
  - No ambiguity in server-side audit trail — the client's status code matches the log's cause.
- **Cons:**
  - Adds a second presence-disclosure posture the service exposes. Tenancy still wants 404 for cross-tenant; malformed wants 400 — so the service's stance becomes "400 for malformed, 404 for everything-else-unreadable". Two postures where one could suffice.
  - Enables a low-value enumeration signal: an attacker who gets `400` knows their id was syntactically wrong (and thus knows that shape-valid probes from then on will produce the more informative `404` vs `200`). Tiny information leak; in combination with other timing or shape differences, contributes to service fingerprinting.
  - Handler error ladder grows by one branch. Every future mutation handler (PUT, DELETE) will carry the same third branch.
  - Introduces a new domain sentinel (`ErrInvalidID`) whose only job is to translate a specific Postgres SQLSTATE. The architectural value of introducing a domain error should be *"it represents a domain concept"*; here the concept is really just *"the driver parsed your input wrong"* — arguably a tell that the abstraction is off.

### Option C: Validate UUID at the handler edge

Handler parses `{id}` via `uuid.Parse` (from `github.com/google/uuid` or similar) before calling storage. On parse failure: return `400 Bad Request`. Storage never sees malformed input; no boundary translation needed.

- **Pros:**
  - Fails fast at the handler; no DB roundtrip for obviously-invalid input.
  - Storage layer's error surface stays at its current minimum (`ErrNotFound`, `ErrInvalidInput`).
  - Feels tidy: validation-at-the-edge is already an established L01 pattern (see `validate(*Workflow)`).
- **Cons:**
  - Duplicates validation logic. DB already rejects malformed UUIDs via `22P02`; now the handler does too. **Two sources of truth for UUID validity** — if the handler parser diverges from Postgres's UUID parser (e.g., case sensitivity, hyphen tolerance, microsoft-braces `{uuid}` form), you get inconsistent behavior between "direct handler call" and "handler call that somehow reached storage anyway".
  - Wrong layer for this concern. The L01 validate-at-edge pattern exists for **domain invariants** (trim whitespace on name, check trigger_type enumeration) — facts about the business model. UUID syntax is a **framework concern** already handled by the driver + DB type system. Pushing it into the handler inflates validation-at-edge into a catch-all.
  - Adds a new dependency (`google/uuid` or equivalent) for what is effectively a type-coercion question.
  - Still doesn't resolve the architectural decision — now you have to decide what the handler returns on parse failure (400 or 404?). Options A and B collapse back in anyway, just at a different layer.
  - In a future world where `{id}` could be a different type (ULID, KSUID, a migration period supporting both), moving validation into the handler turns a storage-boundary change into a handler-layer change with more blast radius.

## Decision

**Option A — map `22P02 → ErrNotFound` (404 posture).**

**Tipping architectural factor:** unifies presence-disclosure across the whole service. Malformed UUID, valid-but-absent UUID, and (landing later this level) valid UUID belonging to another tenant all produce the identical `404 Not Found` — the client can neither enumerate the ID space nor distinguish *"doesn't exist"* from *"not yours"*. One service-wide stance beats two coexisting contracts.

**Accepted con:** loses the client-observable distinction between *"you sent garbage"* (shape-wrong) and *"you sent a plausible ID that isn't present"* (shape-right). In practice this distinction is rarely load-bearing for real clients; debugging is preserved by structured server-side logging that distinguishes the two cases even though the HTTP status does not. This is the trade: client opacity, server transparency.

**Composes with tenancy:** this decision *sets* the stance the tenancy-isolation ADR inherits. Cross-tenant reads returning 404 is no longer its own architectural question — it is the same posture applied to a different layer. Writing this ADR first lets the tenancy ADR reference the posture rather than re-litigate it.

## Consequences

**New code.** `storage.GetByID` gains a second error-translation branch. Alongside the existing `errors.Is(err, pgx.ErrNoRows) → ErrNotFound`, translate `pgconn.PgError` with `Code == "22P02"` to the same `ErrNotFound`. No new sentinel is introduced — `ErrNotFound` gains a second trigger path. Requires a new import: `github.com/jackc/pgx/v5/pgconn`. Handler error ladder is unchanged (stays two-branch: `ErrNotFound` + catch-all 500). Same translation applies to any future read method that accepts a UUID parameter (`GetByID` and `DELETE`'s storage method when it lands; `UPDATE`/`PUT` inherits via its `GetByID`-shaped branch).

**Server-side log.** Because the status code collapses two distinct causes into one response, the storage boundary must log the *distinguishing* cause before returning `ErrNotFound`. Minimum shape: one log line that records the SQLSTATE code (so `22P02` vs "no rows" is visible in ops triage) and the ID that was attempted. Without this, the loss of client-observable distinction leaks into the server's own debuggability — exactly what this option is supposed to *preserve*.

**Reversibility.** Local edit at the storage boundary if the decision is overturned later — introducing `ErrInvalidID` and wiring it through the handler error ladder is roughly an hour of mechanical work. **Caveat:** if this 404 stance becomes a *public API contract* that external clients depend on (reading 404 as "not present" specifically, not "malformed OR not present"), switching to a 400-for-malformed stance later becomes a *breaking change* for those clients. At L02 scale — pre-deployment, no external clients — this is free. Worth re-confirming before cutting a v1 release at L04.

**Locks in for future ADRs.** (a) `tenancy-isolation` inherits the 404 posture for cross-tenant reads — does not reopen the presence-disclosure question. (b) `auth-model` is orthogonal — 401 sits on a different axis (authentication vs resource presence) and is unaffected. (c) `test-strategy` gains a concrete branch to cover: the handler test for `GET /workflows/<malformed-uuid>` must verify status 404 *and* a distinguishing server-side log line, not 400 or 500.
