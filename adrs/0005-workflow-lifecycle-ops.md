# 0005. Workflow lifecycle ops (PUT/DELETE)

**Status:** Accepted
**Date:** 2026-04-30
**Level:** L02 (auth-tenancy)

## Context

L02 warm-up #2: implement `PUT /workflows/{id}` + `DELETE /workflows/{id}`. Three architectural decisions stack, none mechanical.

Forces converging:

- **ADR 0003 + ADR 0004 locked opacity-on-wire** — 404 byte-identical at storage, JSON envelope at transport. Same posture should extend to lifecycle ops, or it breaks at the third surface.
- **Tenancy isolation arrives next this level.** Every read query gains `WHERE tenant_id = $1`. A soft-delete filter (`AND deleted_at IS NULL`) joins the same WHERE clause — combined edit cheaper than two passes.
- **L05+ runs reference workflow rows.** Hard delete creates orphan-ref class (run history points at vanished workflows) or forces cascade-delete complexity. Soft delete keeps run history stable.
- **Client retries.** Network blip mid-DELETE → client retry. Strict 404 on second call gives wrong signal ("delete failed"). Idempotent 204 stays safe.
- **PUT semantics constrained by prior schema.** Workflow IDs server-generated via `gen_random_uuid()` (L01). Client cannot PUT to an unborn ID — they don't know it. Upsert PUT requires client-supplied IDs; not architecturally free.

Decision scope: three independent axes.

1. **PUT semantics** — strict (must exist) vs upsert (creates on miss).
2. **DELETE existence** — strict 404 vs idempotent 204.
3. **Storage shape** — hard delete (row gone) vs soft delete (`deleted_at` column).

Out of scope: workflow run history (L03), retention policy (L05+ when data grows), partial unique indexes (deferred until a uniqueness constraint motivates one), undelete API (L02 has no recovery surface — manual SQL only).

## Options Considered

Two internally-consistent postures.

### Pairing A — event-shaped, errors loud

PUT strict / DELETE strict 404 / hard delete.

- **Pros:**
  - Schema unchanged. No migration. Every existing read query untouched.
  - ~30 LOC per handler. Mechanical extension of L01 patterns (storage interface method, handler method, route registration).
  - Simplest mental model: "delete = gone."
- **Cons:**
  - DELETE not idempotent. Retry-after-success returns 404. Client thinks delete failed.
  - Information leak on retries: 404-after-204 observable from outside.
  - Breaks ADR 0003 + 0004 opacity stance at the third surface (storage / transport / lifecycle).
  - No audit trail. No accidental-delete recovery.
  - L05+ runs reference vanished workflows → orphan-ref cleanup or cascade-delete required.

### Pairing B — state-shaped, retries safe

PUT strict / DELETE idempotent 204 / soft delete (`deleted_at timestamptz NULL`).

- **Pros:**
  - Idempotent DELETE. Retries safe by construction.
  - Audit trail. `deleted_at IS NOT NULL` is queryable history of what got removed and when.
  - Extends ADR 0003 + 0004 opacity to deletes — DELETE on never-existed, on existing, and on already-soft-deleted all return 204. Fourth layer of "wire reveals nothing about presence."
  - Tenancy filter + soft-delete filter compose into one WHERE clause edit.
  - L05+ runs reference workflows that don't disappear under them.
- **Cons:**
  - Schema migration. `ALTER TABLE workflows ADD COLUMN deleted_at timestamptz`.
  - **Every** read query gains `WHERE deleted_at IS NULL`. Easy to forget; rows leak through any query that misses the filter. Durable mental tax.
  - Partial unique indexes needed if name uniqueness becomes a constraint (`CREATE UNIQUE INDEX ... WHERE deleted_at IS NULL`).
  - Retention policy eventually needed (deferrable to L05+).

### Why no upsert PUT

Upsert PUT means "client supplies ID; server creates if missing." This requires client-known IDs. Workflow IDs are server-generated (`gen_random_uuid()` per L01 schema). A client cannot PUT to an ID they don't know — which means upsert PUT only fires for IDs the client *invented*, contradicting "server owns ID generation." Forced into strict PUT by the prior schema decision, not actually a free choice. Documenting the cascade so future-reader sees why this isn't on the table.

## Decision

**Pairing B: PUT strict / DELETE idempotent 204 / soft delete.**

**Tipping factor:** Pairing B extends the opacity-on-wire posture established by ADR 0003 (storage 404) and ADR 0004 (transport envelope) into a third surface — the lifecycle DELETE response. Idempotent 204 + soft delete makes DELETE on never-existed, on existing, and on already-soft-deleted return identical responses; three ADRs now share the stance, turning opacity from a tactical per-decision choice into a service-wide invariant future contributors inherit by reading the codebase. Pairing A breaks the invariant on the second DELETE retry — a 404-after-204 transition leaks "this ID existed once" via observable response shift, undoing what ADRs 0003 and 0004 were chosen to prevent. Cost accepted: every read query gains `WHERE deleted_at IS NULL`, a discipline-by-convention tax that's manageable at L02's two read sites but becomes the natural pressure point for moving the filter into RLS or a query helper once the count grows past ~6-8.

**Cons accepted:**

- **Schema migration.** Upfront cost. Single `ALTER TABLE workflows ADD COLUMN deleted_at timestamptz`. Mitigated by being a one-time edit at L02 scale.
- **"Every read query needs the filter" tax.** L02 has only two read sites (`GetByID`, `List`). Manageable now; revisit when query count grows past ~6-8. That's the natural pressure point for either RLS (push filter into DB) or query-builder discipline.
- **Retention deferred.** Soft-deleted rows accumulate. Documented as known L05+ task; not a blocker at L02 with low row counts.
- **No undelete API.** Recovery via manual SQL only. Acceptable until a real audit/recovery story lands at L05+.

**Composition with prior ADRs:**

- **ADR 0003 (22P02 → 404):** unchanged. Storage continues to translate malformed UUIDs into `ErrNotFound`. Delete handler implements idempotency by *ignoring* `ErrNotFound` (treats it as no-op success), keeping the storage translation rule uniform.
- **ADR 0004 (httpx envelope):** PUT/DELETE error responses use the same `httpx.WriteError` envelope. DELETE success uses 204 No Content with no body — `httpx` doesn't grow a "no-body" helper for one site (handler writes status + nothing).

## Consequences

**Schema migration (new file `project/schema/0002_soft_delete.sql` or inline in current `schema.sql`):**

```sql
ALTER TABLE workflows ADD COLUMN deleted_at timestamptz;
```

Single nullable column. No default, no constraint at L02. No index — soft-delete filter joins existing WHERE clauses; index decisions wait for query-pressure data.

**Storage interface (extended):**

```go
type storage interface {
    Create(ctx context.Context, w *Workflow) error
    GetByID(ctx context.Context, id string) (*Workflow, error)
    List(ctx context.Context) ([]*Workflow, error)
    Update(ctx context.Context, id string, w *Workflow) error
    Delete(ctx context.Context, id string) error
}
```

**Storage method semantics:**

- `Update`: `UPDATE workflows SET name = $2, trigger_type = $3, steps = $4, updated_at = now() WHERE id = $1 AND deleted_at IS NULL RETURNING ...`. Zero rows updated → `ErrNotFound`. 22P02 (malformed UUID) → `ErrNotFound` per ADR 0003. Mutable fields: `name`, `trigger_type`, `steps`. Immutable: `id`, `created_at`. `updated_at` server-managed.
- `Delete`: `UPDATE workflows SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`. Returns nil regardless of rows-affected count. 22P02 still translated to `ErrNotFound` (uniform translation per ADR 0003); handler ignores `ErrNotFound` for idempotency.
- `GetByID` (existing, modified): add `AND deleted_at IS NULL` to WHERE. Soft-deleted rows return `ErrNotFound`.
- `List` (existing, modified): add `WHERE deleted_at IS NULL`. Soft-deleted rows excluded.

**Handler:**

- `PUT /workflows/{id}` — `update` method. Reuses Create's body-decode + `MaxBytesReader` + `*http.MaxBytesError` 413 branch + `validate`. Calls `store.Update(ctx, id, &workflow)`. Error ladder: `errors.Is(err, ErrNotFound)` → 404 `"workflow not found"`; `errors.Is(err, ErrInvalidInput)` → 400; catch-all 500. Success: 200 with updated workflow body.
- `DELETE /workflows/{id}` — `delete` method. No body decode, no validation. Calls `store.Delete(ctx, id)`. Error ladder: `errors.Is(err, ErrNotFound)` → swallow (idempotency); other errors → 500. Success: 204 No Content with no body. `w.WriteHeader(http.StatusNoContent)` direct (no `httpx.WriteJSON` — empty-body case).
- Routes: `mux.HandleFunc("PUT /workflows/{id}", h.update)`, `mux.HandleFunc("DELETE /workflows/{id}", h.delete)`.

**Tests (warm-up #3):**

- `fakeStore` extends to track soft-deleted state per row (map of id → soft-deleted bool, or sentinel time).
- PUT coverage: 200 (success), 400 (invalid body, body-too-large, validation), 404 (missing or soft-deleted), 413 (oversized body), 500 (storage error).
- DELETE coverage: 204 on existing, 204 on missing, 204 on already-soft-deleted — assert all three return identical responses (opacity invariant).

**Reversibility:**

- **Pairing B → A (soft → hard):** moderate. `DELETE FROM workflows WHERE deleted_at IS NOT NULL` to clear soft-deleted rows; `ALTER TABLE workflows DROP COLUMN deleted_at`; strip `WHERE deleted_at IS NULL` from every query; flip handler DELETE to strict 404 path. ~1-2 hours mechanical at L02 scale; sticky once external clients depend on idempotent DELETE.
- **DELETE idempotent → strict:** mechanical handler change but breaks any client that relied on idempotency (retry semantics shift). Sticky once clients ship.
- **PUT strict → upsert:** requires moving ID generation client-side (architectural rework, not a flag flip). Separate ADR if ever motivated.

**Future ADR locks/opens:**

- **Tenancy isolation (next L02 ADR):** WHERE clause becomes `tenant_id = $1 AND deleted_at IS NULL` everywhere. Two filters, one edit per query. Soft-delete-by-tenant works for free.
- **Workflow runs (L03):** runs reference workflow rows that don't vanish — orphan-ref class avoided by construction.
- **Retention policy (L05+):** scheduled `DELETE FROM workflows WHERE deleted_at < now() - interval '<retention>'` or moved to archival table. Out of scope L02; documented as known future task.
- **Partial unique indexes:** if a uniqueness constraint on `name` (or `name + tenant_id` after tenancy) arrives, `CREATE UNIQUE INDEX ... WHERE deleted_at IS NULL` keeps soft-deleted rows from blocking new-row creates.
- **Undelete API:** if accidental-delete recovery becomes a product requirement, `PATCH /workflows/{id}` with `deleted_at = NULL` is the natural endpoint. Out of scope L02.
- **API versioning (L04+):** DELETE 204 idempotency becomes part of the v1 contract once a client ships. Switching to strict 404 = versioned breaking change.
