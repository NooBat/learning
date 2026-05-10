# Session Log

Append-only. Newest entries at the bottom. Use the session-end skill (`/session-end`) to append structured entries; manual entries are OK but keep the format consistent so the tail is scannable.

Entry types:
- **Session summary** (default, every session)
- `[LEVEL-TRANSITION]` — when moving from one level to the next
- `[RETRO]` — every 3-4 levels or end of tier
- `[ADR-PROPOSED]` / `[ADR-ACCEPTED]` — decision record created or approved
- `[RULE-CHANGE-PROPOSED]` — when a `.claude/rules/` file should be amended
- `[PIVOT]` — when the roadmap changes direction

---

## [2026-04-17] Repo initialized

**Did:**
- Scaffolded learning repo from approved plan (`~/.claude/plans/i-am-currently-a-shiny-crab.md`).
- Created `CLAUDE.md` (auto-loaded entry point) + `ABOUT-ME.md` + `ROADMAP.md` + `STATUS.md` + `LEVELS.md` + this file.
- Created `.claude/rules/`: `collaboration.md`, `session-protocol.md`, `conflict-resolution.md`, `verify-claims.md`.
- Created `.claude/skills/`: `bootstrap`, `start-level`, `write-adr`, `session-end`, `retro`.
- Created `.claude/settings.json` with `"outputStyle": "Learning"` — per-repo learning mode.
- Wrote `levels/L01-mvs.md` (first level brief).
- Wrote `infrastructure/setup-guides/01-go-toolchain.md` and `02-postgres-local.md`.
- Wrote `.gitignore`, ran `git init`.

**Decided:**
- Keystone project: Developer Automation Platform (working name `flux`) — covers scheduling, notifications, feature flags as one system.
- Backend: Go (not TypeScript).
- Tracking: level-based, not calendar-based.
- Source of truth: this repo pushed to private GitHub. Claude-side memory treated as disposable cache.
- Rules format: `.claude/rules/*.md` — native Claude Code convention. Always-apply rules have no frontmatter per docs.
- Collaboration model: Claude writes guides + scaffolds, Daniel runs commands + writes business logic.

**Blocked:**
- Private GitHub repo not yet created. Daniel needs to create `learning` (or another name) as a private repo on github.com, then push.

**Next:**
- Daniel creates the remote repo and pushes.
- Install Go per `infrastructure/setup-guides/01-go-toolchain.md`.
- Install Postgres per `infrastructure/setup-guides/02-postgres-local.md`.
- Begin L01: init Go module in `project/`, write first handler.

---

## [2026-04-18] [ADR-PROPOSED] 0001 — HTTP router choice

Pause on implementation. Decision doc drafted at `adrs/0001-router-choice.md`. Daniel to review four options (`net/http` stdlib, `chi`, `gin`, `gorilla/mux`) and fill in the Decision section + set Status=Accepted before the first handler is written.

## [2026-04-18] [ADR-ACCEPTED] 0001 — HTTP router choice → `net/http` stdlib

Decision locked: Option A (`net/http` stdlib, Go 1.22+ ServeMux). Reasoning: cheap reversibility via shared `http.Handler` interface; stdlib teaches the canonical HTTP model; explicit accept of verbose middleware composition until L03.

## [2026-04-18] [ADR-PROPOSED] 0002 — Postgres driver choice

Pause on implementation. Decision doc drafted at `adrs/0002-postgres-driver.md`. Daniel to review four options (`pgx` native, `database/sql` + `pgx/stdlib`, `database/sql` + `lib/pq`, `database/sql` + `pgx/stdlib` + `sqlx`) and fill in the Decision section + set Status=Accepted before the first query is written. This is the second of the L01 "first dependencies" pair.

## [2026-04-18] [ADR-ACCEPTED] 0002 — Postgres driver choice → `pgx` native

Decision locked: Option A (`pgx` native v5 + `pgxpool`). Architectural posture — Postgres-first, not SQL-generic. Tipping factors: JSONB fidelity for `workflows.steps` domain field; forward compatibility with Postgres-specific features (LISTEN/NOTIFY, batched queries) at L05/L06. Explicitly accepting: driver lock-in (no swap to MySQL/SQLite without rewriting query sites); pgx mental model for future contributors.

## [2026-04-19 → 2026-04-20] [LEVEL-TRANSITION] L01 closed — CRUD-minus-U-D shipped

Long session spanning two calendar days. L01 closed with Create / Read-one / List plus all supporting architecture; PUT/DELETE deliberately scope-narrowed into L02 as mechanical extensions of patterns already in place. Pre-closeout smoke test passed end-to-end against real Postgres.

**Did (code — all in `project/`):**
- `schema.sql`: single `workflows` table. `uuid` PK via `gen_random_uuid()` (pgcrypto extension). `trigger_type text` with CHECK constraint (`'schedule' | 'webhook' | 'manual'`) — chose CHECK over ENUM because the taxonomy is an evolving business concept; DROP/ADD CHECK constraint beats the ENUM rename/migrate/drop rebuild cost. `steps jsonb NOT NULL DEFAULT '[]'`. `created_at`/`updated_at timestamptz DEFAULT now()`.
- `internal/workflows/types.go`: `Workflow`, `Step`, `TriggerType` (typed string) + `ValidTriggerTypes` slice. `Step` is intentionally opaque at L01 (`Kind string` + `Config map[string]any`) — step execution arrives at L05.
- `internal/workflows/storage.go`: concrete `Storage` type backed by `*pgxpool.Pool`. `Create` uses `INSERT ... RETURNING id, created_at, updated_at` + single-roundtrip `QueryRow(...).Scan(...)` for server-generated-field hydration. `GetByID` scans into a fresh stack-allocated `var w Workflow`, translates `pgx.ErrNoRows` → `ErrNotFound` via `errors.Is` (survives wrapping), bubbles any other error raw. `List` iterates `pgx.Rows` with `defer rows.Close()` right after the error check (NOT before — would nil-deref on Query error), fresh `var w Workflow` *inside* the loop (outside would alias all slice entries to the last row — classic Go footgun), `rows.Err()` after the loop (catches errors `rows.Next()` hides behind its `false` return), `ORDER BY created_at DESC` for stable newest-first ordering, `[]*Workflow{}` empty-slice return to marshal as `[]` not `null`.
- `internal/workflows/handler.go`: `Handler` + unexported consumer-defined `storage` interface. Methods: `create` / `getByID` / `list` + `Register(mux)`. `validate(*Workflow)` trims name, checks empty + 500-rune cap via `utf8.RuneCountInString` (not `len` — bytes ≠ runes), checks `slices.Contains(ValidTriggerTypes, ...)`; returns `wrapInvalid(...)` which wraps `ErrInvalidInput` with `%w` so `errors.Is` works at the handler. Error ladder shape: specific domain branches (`ErrNotFound`, `ErrInvalidInput`) first, catch-all `if err != nil` 500 last — enforces exhaustiveness that Go's type system doesn't. Buffer-first encode (`bytes.Buffer` materialized before any header flushed) — memory trade for mid-encode failure recoverability. Server-side detailed `log.Printf`, client-side opaque `"internal error"` — info-disclosure posture. Method-prefixed routes: `POST /workflows`, `GET /workflows/{id}` (brace syntax — Go 1.22+ stdlib, not `:id` which is Express/chi), `GET /workflows`.
- `cmd/server/main.go`: `main` is a shim; all work in `run(ctx) error`. Signal-driven ctx via `signal.NotifyContext(bg, SIGINT, SIGTERM)`. `run` reads `DATABASE_URL`, opens `pgxpool`, runs a 5s-bounded `Ping`, wires Storage + Handler + mux + `/healthz` (registered in `main` NOT inside the workflows handler — app-level vs domain-level layering), builds `http.Server`. **Lifecycle orchestration:** buffered-size-1 `errCh`, goroutine wrapping `ListenAndServe`, select on `<-ctx.Done()` vs `<-errCh`, shutdown context rooted at `context.Background()` with 10s timeout (NOT derived from parent ctx — parent is already cancelled at this point, so a derived ctx would be pre-expired and `Shutdown` would skip draining).

**Did (bookkeeping at close):**
- Ran `go build ./cmd/server` — clean compile.
- End-to-end smoke test against live Postgres: 201 on Create with full hydrated response body, 200 on GetByID and List, 400 on empty-name + invalid-trigger + malformed JSON, 404 on well-formed-but-unknown UUID, 500 on malformed UUID (documented gap), 405 on wrong method (auto from stdlib mux).
- Updated `STATUS.md` (L01 → complete, next-session target listing the L02 warm-ups).
- Updated `LEVELS.md` (L01 → `[x]`).
- Updated `levels/L01-mvs.md` (status → complete, exit criteria tick state, added "What actually shipped" + "Deferred to L02" + "Known architectural gap" sections).
- Updated `.gitignore` to exclude `/project/server` (build artifact) and `.serena/` (Serena MCP machine-local cache).

**Decided (L01-wide design choices, each with explicit rejected alternative):**
- **Go enum pattern: typed `string`, not `iota int`.** `TriggerType` crosses three serialization boundaries (JSON wire, Postgres `text` column, Go internal code). A typed string is transparent on all three — no custom `MarshalJSON`/`Scanner` plumbing. An `iota int` would have forced all three boundaries to carry custom marshalers, with breakage modes invisible at compile time.
- **Consumer-defined, unexported interface — NOT producer-exported `IStorage`.** The `storage` interface is declared in `handler.go` (the consumer), contains only the three methods the handler actually needs, and is unexported. Follows the "accept interfaces, return concrete types" Go idiom; avoids the "preemptive interface" anti-pattern where a package exports an interface before a second implementation exists. Pays off immediately as a test affordance (a `fakeStore` satisfying the interface is the L02 test entry point).
- **Validation at the handler edge, mutating-in-place — pattern A of five considered.** Options weighed: A (edge validate in handler), B (separate Request DTO → Domain mapping — "ports and adapters" tactical variant), C (parse-don't-validate / smart constructors), D (struct-tag validator library), E (trust DB CHECK only + map constraint errors). Chose A: the DB CHECK constraint is the real invariant source of truth; `validate()` is the UX push that turns a CHECK violation (500) into an early 400 with a human-readable body. Will likely evolve toward B as auth-specific invariants (tenant ID, permissions) layer in at L02.
- **Buffer-first encode posture for every success response.** All three handler methods encode into `bytes.Buffer` before flushing headers + status + body. Trade: memory (full response in RAM) for mid-encode failure recoverability (can still return 500 instead of a corrupted 200). Revisit trigger: pagination (L03-ish) or response-size growth.
- **Shutdown context rooted at `context.Background()`.** The parent ctx is *already cancelled* when shutdown begins — that's *why* shutdown is running. Deriving from it would give a pre-expired context; `Shutdown` would abort in-flight requests instantly instead of draining them. Single most common Go-service lifecycle bug.
- **PUT/DELETE scope-narrowed out of L01.** The architectural lessons L01 was structured to teach — boundary translation, consumer interfaces, error ladder exhaustiveness, buffer-first encode, lifecycle orchestration — were all covered by Create + GetByID + List. PUT/DELETE would be ~30 lines of mechanical extension with no new lesson payload. Pulled into L02 as a warm-up. This is a deliberate scope decision, not a slip — documented in `levels/L01-mvs.md` "What actually shipped".
- **Testing deferred to L02 opener, not wedged into L01 closeout.** The consumer-defined `storage` interface is the single biggest test affordance in this codebase; writing the first test suite using that affordance deserves its own dedicated learning moment. Diluting it into "and also close L01" would under-invest in both.
- **No checked-in smoke script (`scripts/smoke.sh`).** Considered and rejected at closeout (option B). Rationale: interactive smoke test already ran clean end-to-end; a bash script would be lower-leverage than the `fakeStore` Go test suite planned for L02 opener. If useful later, reconsider alongside L03 CI work.

**Collaboration-pattern correction mid-session (IMPORTANT — future sessions watch for this):**
Daniel explicitly flagged drift — "Weird, why you don't let me code anything?" and "Why didn't you let me setup anything during this?" — calling out that the session had slipped into Claude-writes-implementation / Daniel-accepts, instead of Daniel-writes / Claude-reviews-designs. Invoked "Rewind": Claude-authored method bodies across `storage.go`, `handler.go`, and `cmd/server/main.go` were stripped back to `panic("TODO: …")` stubs. Daniel then re-implemented each method himself, with Claude restricted to coaching the architectural questions (error translation direction, context propagation, channel buffering size, shutdown-ctx rooting, route syntax).

This is a concrete worked example of `.claude/rules/collaboration.md` operating as intended. The failure mode has a clear smell: *Claude writes a 20-line method body; Daniel says "ok, next"*. The correct shape is: *Claude sets up a TODO(human) marker + Learn-by-Doing request; Daniel writes; Claude reviews as a staff engineer*. Post-rewind the session produced better learning (every storage + handler method written by Daniel with iterative review) and shipped the same feature set. Future sessions should self-audit on this boundary — drift usually starts with "let me just scaffold this one quick thing".

**Did (ADR work — all on 2026-04-18 but noted here for L01 summary completeness):**
- `adrs/0001-router-choice.md` — `net/http` stdlib (over chi / gorilla / gin+echo). Tipping factor: Go 1.22+ method patterns eliminate the ergonomic gap that historically motivated third-party routers; reversibility is cheap via the shared `http.Handler` interface.
- `adrs/0002-postgres-driver.md` — `pgx` native v5 + `pgxpool` (over `database/sql` + stdlib/lib-pq + sqlx). Tipping factor: JSONB fidelity for `workflows.steps` + forward compatibility with Postgres-specific features (LISTEN/NOTIFY, batched queries) at L05/L06. Explicitly accepted: driver lock-in.

**Blocked:**
- None.

**Next:**
- `/start-level 02 auth-tenancy` — write the L02 brief.
- L02 openers bundle the L01 carry-overs: PUT/DELETE, 22P02 error translation, first Go test suite using the `fakeStore` affordance, `project/docs/architecture.md` stub. L02's core goal (auth + tenancy) stacks on top of these.

## [2026-04-20] [LEVEL-TRANSITION] L01 → L02

- **Completed:** L01 — Go HTTP + Postgres CRUD workflows service shipped at commits `6e75813` (code) and `624150c` (bookkeeping). Exit criteria met for Create / Read-one / Read-list; PUT/DELETE, malformed-UUID translation, architecture.md stub, and the first test suite all deferred to L02 as warm-ups.
- **Started:** L02 (auth-tenancy) — layer identity and multi-tenancy onto the workflows service, and plant the first Go test suite using the consumer-defined `storage` interface as the seed affordance. Brief scaffolded at `levels/L02-auth-tenancy.md`; ready for Daniel's review + exit-criteria refinement before any code moves.
- **Why L02 now:** all L01 exit criteria either shipped or were explicitly re-homed to L02 with architectural rationale (see L01 "What actually shipped" recap). No open blockers; working tree clean post-`624150c`.
- **Key artifacts from L01:**
  - `adrs/0001-router-choice.md` — accepted: `net/http` stdlib (Go 1.22+ method patterns close the historical ergonomic gap that motivated chi/gin; reversibility cheap via `http.Handler`).
  - `adrs/0002-postgres-driver.md` — accepted: `pgx` native v5 + `pgxpool` (JSONB fidelity for `workflows.steps`; forward compatibility with Postgres-specific features at L05/L06; driver lock-in explicitly accepted).
  - Six architectural patterns in the codebase: (1) boundary translation (`pgx.ErrNoRows` → `ErrNotFound`), (2) consumer-defined unexported `storage` interface in `handler.go`, (3) validation-at-the-edge with in-place mutation, (4) buffer-first JSON encode, (5) Go 1.22+ method-prefixed ServeMux with `{id}` brace syntax, (6) lifecycle orchestration rooted at `context.Background()` for shutdown (not the already-cancelled parent).
- **Carry-over from L01 (bundled as L02 warm-up tasks, in order):**
  1. ADR `malformed-uuid-translation` — choose 404 (reuse `ErrNotFound`) or 400 (introduce `ErrInvalidID`) for Postgres `SQLSTATE 22P02`. Recommended bias: 404, because tenancy's presence-disclosure posture arriving this same level wants cross-tenant reads to return 404 too. Picking 404 here lets one posture rule the whole service.
  2. Implement `PUT /workflows/{id}` + `DELETE /workflows/{id}`. Mechanical extension — but a good place to notice `PUT`'s error ladder is `Create` + a 404 branch from `GetByID` (compositional pattern).
  3. Translate `22P02` at the storage boundary per the ADR.
  4. First Go test suite: `httptest.NewServer` + `fakeStore` satisfying the unexported `storage` interface; cover every 400/404/500 branch at the handler level. Add one real-DB integration test for the golden path.
  5. Write `project/docs/architecture.md` documenting the L01 baseline before L02 extends it.
- **L02 ADRs queued:** `malformed-uuid-translation` (warm-up #1), `auth-model`, `tenancy-isolation`, `test-strategy`. The first three are architecturally load-bearing; the fourth sets the testing pyramid shape every later level inherits.
- **Opening decision biases (to be challenged via ADR — not accepted as default):**
  - **Auth model:** opaque bearer token with a server-side `tokens` table, over JWT / session / OIDC. Simpler primitives, explicit revocation, defers the JWT footgun landscape (algorithm confusion, clock skew, key rotation) to a later level when service-to-service need actually motivates it.
  - **Tenancy isolation:** application-level filtering (`WHERE tenant_id = $1` threaded through every query) over Postgres RLS / schema-per-tenant. Cheap to start, explicit in code, trivially testable; RLS documented as the future migration path if defense-in-depth concerns emerge.
  - **Presence disclosure:** cross-tenant reads return 404 (not 403). Consistent with the 22P02 decision above. A 403 leaks existence; a 404 doesn't.
- **Blocked:** None.
- **Next:**
  - Daniel reads `levels/L02-auth-tenancy.md` and refines exit criteria.
  - Commit the scaffold: `git add -A && git commit -m "scaffold L02: auth-tenancy"`.
  - Invoke `/write-adr malformed-uuid-translation` before touching any code — the 404-vs-400 decision sets the presence-disclosure posture the rest of L02 inherits.

## [2026-04-21] [ADR-ACCEPTED] 0003 — Malformed UUID translation → 404

- **Did:** Closed warm-up #1 ADR. Decision: Postgres `SQLSTATE 22P02` translates to `ErrNotFound` at the storage boundary (option A — reuse existing sentinel, no new `ErrInvalidID`). One presence-disclosure posture for the whole service: malformed, missing, and (later) cross-tenant UUIDs all return byte-identical 404 bodies.
- **Decided:** Translation lives in storage's `GetByID` via `*pgconn.PgError` `errors.As` + `Code == "22P02"` check. Server-side `log.Printf` retains the raw SQLSTATE for debuggability — opacity-on-wire, transparency-in-log. Handler error ladder unchanged (still keys on `errors.Is(err, ErrNotFound)`).
- **Why option A over `ErrInvalidID`:** the malformed UUID is a transport-parsing failure, not a domain concept. Inventing a domain sentinel for it would couple the domain to URL-encoding choices. Tenancy isolation (later this level) needs the same 404 anyway — picking 404 here lets one rule cover both layers.
- **Blocked:** None.
- **Next:** Begin ADR 0004 (HTTP response utilities) — extracting the buffer-encode + `http.Error` patterns duplicated across handlers, ahead of L03 `runs` adding a second consumer.

## [2026-04-23] [ADR-PROPOSED] 0004 — HTTP response utilities

- **Did:** Drafted ADR 0004 covering three options (private same-package helpers / shared `internal/httpx` package / responder framework) + two sub-decisions (envelope shape, status ownership). Ran caveman-compression pass on the prose to keep it scannable.
- **Decided:** Status `Proposed`. No implementation yet — the boundary choice is the architectural pivot, and the envelope decision is sticky once external clients exist (Shape 2 → Shape 3 = breaking for clients reading `error` as a string). Both worth pinning before code.
- **Blocked:** Decision pending Daniel's review.
- **Next:** Daniel reviews options and locks the decision. If B chosen, scaffold `project/internal/httpx`.

## [2026-04-25] [ADR-ACCEPTED] 0004 — HTTP response utilities → Option B + Shape 2 + Ownership 1

- **Did:**
  - Accepted ADR 0004 (`adrs/0004-http-response-utilities.md`). Decision: shared `internal/httpx` package (option B) + JSON-simple envelope `{"error": "<msg>"}` (Shape 2) + handler picks HTTP status (Ownership 1).
  - Created `project/internal/httpx/httpx.go` (~50 LOC). Three exports: `WriteJSON` (buffer-first encode), `WriteError` (envelope wrapper delegating to `WriteJSON` for shape consistency), `DecodeJSON` (wraps `json.NewDecoder(r.Body).Decode(dst)` with `DisallowUnknownFields()` — Daniel's implementation per Learn-by-Doing).
  - Refactored `internal/workflows/handler.go` to consume httpx helpers. Removed `bytes` + `encoding/json` imports; added httpx import. Buffer-encode dance × 3 collapsed; `http.Error` calls × 7 swapped to `httpx.WriteError`. Error ladder shape unchanged.
  - Committed as `ddba176` (ADR) + `80c0b50` (httpx package + handler integration).
- **Decided (the three architectural pivots, recap):**
  - **Option B over A:** turns the envelope invariant from convention into compile-time property before L03 `runs` adds a second consumer. ADR 0003's byte-identical-404 guarantee becomes a build-time property.
  - **Shape 2 over Shape 1 (text/plain) and Shape 3 (structured):** machine-parseable for log aggregators and frontend rendering, without designing `code` semantics that may never be needed. Shape 2 → 3 sticky once clients ship; cheap pre-deploy.
  - **Ownership 1 over Ownership 2:** domain errors stay transport-agnostic. `ErrNotFound` / `ErrInvalidInput` carry zero HTTP knowledge — reusable in CLI / worker / future gRPC surfaces.
- **Did (also):** WriteJSON encode-failure path writes the canonical Shape 2 body as a literal (`{"error":"internal server error"}`) rather than recursing through `WriteError` — avoids infinite-recursion class on a marshal failure.
- **Blocked:** None.
- **Next:** Validation refactor in handler — body size cap + decode error handling decisions.

## [2026-04-29] handler validation refactor + opacity decisions

- **Did:**
  - `httpx.DecodeJSON` settles as `decoder.DisallowUnknownFields()` + `Decode(dst)` — extra fields rejected at the helper level, no malformed-vs-extra distinction.
  - Handler keeps decode-error handling: opaque "invalid json" 400 to the client, raw `err` logged server-side. Same opacity-on-wire / transparency-in-log posture as ADR 0003 — generalized from storage boundary to request boundary.
  - Added `KiB` / `MiB` package-level constants (`1 << 10`, `1 << 20`) using bit shifts.
- **Decided:**
  - **Opacity at the handler, not the helper.** httpx is a transport encoder, not an error policy layer — handler picks the user-facing message and the log level. Avoids a new sentinel error type that would couple httpx to caller semantics.
  - **Body size cap location: handler-side (option B), not helper-side.** Per-endpoint flexibility (different limits per route possible later); helper stays JSON-only; future migration to middleware is mechanical when auth / upload routes need different caps.
- **Blocked:** Variable naming for the `*http.MaxBytesError` distinguish — `isSizeExceeded` was awkward; settled on idiomatic `mbe, ok := errors.AsType[*http.MaxBytesError](err)` next session.
- **Next:** Implement MaxBytesReader at handler edge and ship.

## [2026-04-30] L02 warm-up #1 closed + option 2 framing

- **Did:**
  - `internal/workflows/handler.go` `create` now caps body at `1 * MiB` via `http.MaxBytesReader(w, r.Body, maxBodySize)`. `*http.MaxBytesError` branch returns 413 (`http.StatusRequestEntityTooLarge`) with `mbe.Limit` logged server-side. Non-size decode failures continue to 400 "invalid json".
  - Build + vet clean. Committed as `26b302b`.
  - Updated `STATUS.md` to reflect ADR 0003 + 0004 acceptance, httpx shipped, body cap landed; queued ADR 0005 + tests + architecture.md as remaining warm-ups.
- **Decided:**
  - **413 distinguish over collapsing into 400.** Two reasons: client-actionable (413 tells them to send less data; 400 doesn't), and observable (separate metric bucket for "client too noisy" vs "client malformed" once metrics arrive at L05+). Cost: one extra branch in the error ladder.
  - **`errors.AsType` over `errors.As`.** Go 1.26 stdlib generic — no zero-value placeholder var, returns `(T, bool)` directly. Verified by build (initially assumed third-party, was wrong; lesson: verify by doing, not by memory).
- **Blocked:** None.
- **Next session target:**
  - **Option 2 (PUT/DELETE) is ADR-worthy first.** Three stacked decisions: PUT semantics (strict vs upsert), DELETE existence (strict 404 vs idempotent 204), hard vs soft delete. ADR drives `storage` interface signatures.
  - Two natural pairings: (hard, strict) for L02 simplicity, or (soft, idempotent) to extend ADR 0003 opacity to deletes + add retention/audit posture.
  - Recommendation: single ADR `0005-workflow-lifecycle-ops.md` covering all three (decisions compose; splitting fragments the trade-off space). After ADR: extend storage interface, then handler methods + routes.

## [2026-05-01] [ADR-ACCEPTED] 0005 — Workflow lifecycle ops → Pairing B + warm-up #2 shipped

- **Did (ADR):** Drafted, accepted, and committed ADR 0005 (`adrs/0005-workflow-lifecycle-ops.md`, commit `8a13ecc`). Decision: PUT strict / DELETE idempotent 204 / soft delete via `deleted_at`. Tipping factor framed as opacity-on-wire through-line: ADRs 0003 (storage 404), 0004 (transport envelope), 0005 (lifecycle DELETE) compose into a service-wide stance — three ADRs sharing the posture turn opacity from per-decision tactic into inherited invariant.
- **Did (implementation):**
  - `project/schema.sql`: `deleted_at timestamptz` (nullable, no default) inlined into `CREATE TABLE workflows`. Idempotent on re-run since `CREATE TABLE IF NOT EXISTS` covers the whole shape.
  - `project/internal/workflows/storage.go`:
    - New `Update(ctx, id, *Workflow)` — `UPDATE ... RETURNING id, created_at, updated_at`. Mutable: `name`/`trigger_type`/`steps`. Server-managed: `id`/`created_at`/`updated_at`.
    - New `Delete(ctx, id)` — `UPDATE workflows SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`. Uses `Exec` (returns `CommandTag`), not `QueryRow`.
    - `GetByID` and `List` gain `WHERE deleted_at IS NULL`.
    - New `translatePgError(err) error` helper — maps `pgx.ErrNoRows` and `*pgconn.PgError`/22P02 to `ErrNotFound`. Single source of truth for ADR 0003's rule; called by `GetByID`/`Update`/`Delete`. Future SQLSTATE additions (23505, 23503, …) land in this one function.
  - `project/internal/workflows/handler.go`:
    - New `update` (PUT): reuses `MaxBytesReader` + `httpx.DecodeJSON` + `validate` from create. Error ladder: `ErrNotFound` → 404; `ErrInvalidInput` → 400; catch-all → 500. Success: 200 with hydrated body via `httpx.WriteJSON(w, http.StatusOK, workflow)`.
    - New `delete` (DELETE): no body decode, no validation. Calls `store.Delete`; swallows `ErrNotFound` (idempotency); other errors → 500. Success: `w.WriteHeader(http.StatusNoContent)` direct (no body — RFC 9110 §15.3.5).
    - `storage` interface extended with `Update`/`Delete`.
    - Routes registered: `PUT /workflows/{id}`, `DELETE /workflows/{id}`.
- **Did (review iterations — 2 rounds):**
  - **Round 1 issues caught & fixed:** `Delete` used `QueryRow` (silent-error class — pgx defers errors to `Scan`, which never ran); handler `delete` had `_ = h.store.Delete(...)` swallowing all errors including connection-loss; `Update` only RETURNed `updated_at` (response body had empty `id`/zero `created_at`); 22P02 translation missing from `Update`/`Delete`; PUT returned 201 not 200; DELETE wrote `null` body via `httpx.WriteJSON(w, 204, nil)`; `DeletedAt` exposed in `types.Workflow` JSON (storage concept leaking to API); `ALTER TABLE` appended (non-idempotent on re-run).
  - **Round 2:** dead `pgx.ErrNoRows` branch in `Delete` (Exec never returns it); SQL style consistency; DRY threshold — same 22P02 check in three storage methods → extracted as `translatePgError` helper with named contract (the durable artifact future-readers will cite).
- **Did (smoke — 13 cases live against real Postgres on `:8080`):**
  - All passed. The big architectural assertion: cases 8/9/10/11 (DELETE on existing / just-deleted / never-existed / malformed-UUID) returned **byte-identical** 204 No Content with empty body. Pairing B's opacity contract holds end-to-end at runtime.
  - PUT 404 collapses three not-addressable paths (missing UUID / malformed UUID / soft-deleted ID) into one response: `{"error":"workflow not found"}`. Same architectural shape as DELETE 204, different status.
  - PUT 413 fired on a 2_000_046-byte body (cap 1_048_576) — the error-class distinguish from option 1 holds in real conditions, not just unit-test theory.
  - List excluded just-soft-deleted row but kept L01-era rows with `deleted_at IS NULL`. Soft-delete filter discriminates correctly.
- **Decided:**
  - **`translatePgError` extracted at three call sites, not earlier.** Two would have been YAGNI; three is the conventional DRY threshold. ADR 0005 added two methods, hit the threshold, refactor justified itself by the extraction's structural payoff (named contract, single point for future SQLSTATE additions, unit-testable in isolation once tests land).
  - **`DeletedAt` not in `types.Workflow`.** Soft delete is a storage concept the API hides — exposing the field via JSON would contradict pairing B's opacity intent. Field stays out of the wire contract; storage internals don't need it on the struct since all reads filter `WHERE deleted_at IS NULL`.
  - **`w.WriteHeader(http.StatusNoContent)` direct, no `httpx` helper.** RFC 9110 §15.3.5 forbids 204 bodies; growing `httpx` for one no-body case would expand scope unhelpfully. Single-call-site direct write keeps the helper JSON-only per ADR 0004's scope note.
- **Commits:**
  - `8a13ecc` — ADR 0005 accepted.
  - `a175f45` — PUT/DELETE handlers + soft delete + `translatePgError`.
- **Blocked:** None.
- **Next session target:**
  - **Warm-up #3: first Go test suite.** `internal/workflows/handler_test.go` using `httptest.NewServer` + `fakeStore` satisfying the unexported `storage` interface. The four DELETE paths' opacity invariant from this session's smoke is the highest-value test target — assert byte-identical responses by automated comparison.
  - Decision en route to writing the suite: response-body assertion strategy (parse JSON via shared helper vs assert raw bytes). The choice ties into ADR 0004's noted Shape 2 → Shape 3 migration concern — a parsing helper is forward-compatible; raw-byte asserts lock tests to current shape.
  - Possibly motivates writing the ADR for `test-strategy` first if the testing-pyramid shape needs pinning before the suite scaffolds.

## [2026-05-10] [ADR-ACCEPTED] 0006 — Test strategy + warm-up #3 handler-side shipped

- **Did (ADR):** Drafted, accepted, and committed ADR 0006 (`adrs/0006-test-strategy.md`, commit `b353660`). Decision: Path 1C + 2C + 3A — handler-level fakeStore primary + integration ring (`//go:build integration`) + envelope-decode helper centralizing ADR 0004's Shape 2 contract + minimal map-backed fakeStore (no fault injection at L02). Tipping factor framed as the L01 design payoff cashed: the consumer-defined unexported `storage` interface in `handler.go` (chosen over a producer-exported `IStorage`) doubles as test boundary — a `fakeStore` satisfying the interface is exactly what 1C requires, and 3A keeps the fake minimal at L02 scale without pre-engineering for needs that haven't surfaced.
- **Did (scaffold, commit `39103d8`):**
  - `project/internal/workflows/handler_test.go`: fakeStore (`Create`/`GetByID`/`List`/`Update`/`Delete` satisfying the unexported `storage` interface, map-backed, sequential-only) + helpers (`errorMessage`, `requireStatus`, `decodeWorkflow`, `newTestServer`, `readBody`, `newFakeUUID`) + one passing pattern-reference test (`TestGetByID_NotFound`) + `TestDelete_OpacityInvariant` left as `t.Skip` with `TODO(human)` and structured suggested-shape comments.
  - `project/internal/workflows/storage_integration_test.go`: `//go:build integration` build tag + `integrationPool`/`truncateWorkflows` helpers (`DATABASE_URL_TEST` env var) + three named tests with `t.Skip` placeholders and suggested shapes: `TestStorage_Update_RETURNING`, `TestStorage_GetByID_22P02_to_NotFound`, `TestStorage_Delete_SoftDeleteFilter`.
- **Did (TestDelete_OpacityInvariant — Learn-by-Doing, commit `b7577b8`):** Daniel implemented the test in two passes. Round-1 review caught: `store.Create` error ignored (couples test to fake-implementation-detail) and case-2 message ambiguity (cases[0] and cases[1] share `w.ID`, so failure messages couldn't distinguish "first DELETE failed" from "second DELETE failed"). Round-2 fix landed the named-cases struct (`label`/`id` per case) and the seed-error check. The four-DELETE-paths byte-identical assertion is now a regression test instead of one-time smoke.
- **Did (handler branch coverage, commits `0381e4c` + `aac4ac0` + `b7ce027`):**
  - **Create branch (`0381e4c`):** `TestCreate_Success` (201, populated id + server-stamped timestamps), `TestCreate_InvalidJSON` (400 "invalid json"), `TestCreate_ValidationFailures` (table-driven: empty name / overlong name / unknown trigger), `TestCreate_BodyTooLarge` (413 with 2 MiB body vs 1 MiB cap). Helpers added: `postJSON`, `putJSON`.
  - **Update branch (`aac4ac0`):** `TestUpdate_Success` (200 + id/created_at immutable), `TestUpdate_InvalidJSON`, `TestUpdate_NotFound` (table-driven: never-existed + soft-deleted both collapse to one 404 — the table-shape itself is the assertion that both paths must behave identically per ADR 0005), `TestUpdate_BodyTooLarge`, `TestUpdate_ValidationFailures`. Helper added: `seedWorkflow`.
  - **List + GetByID (`b7ce027`):** `TestList_Empty`, `TestList_ReturnsLiveOnly` (soft-deleted excluded — read-site mirror of the opacity invariant), `TestGetByID_Success`, `TestGetByID_SoftDeleted` (404 mirror). Helper added: `decodeWorkflowList`.
- **Did (final state):** 22 tests green across handler_test.go covering 200 / 201 / 204 / 400 / 404 / 413. Soft-delete invisibility asserted at every read site (List, GetByID, Update). Integration ring still 3 `t.Skip` stubs pending DB provisioning.
- **Decided:**
  - **`len == 0` over `bytes.Equal` for opacity-test body assertion.** Discussion surfaced that `len == 0` is strictly stronger for the current contract — it catches uniform RFC violations (e.g., all four paths return `204 + "{}"`) that `bytes.Equal` misses (uniform = passes equality). `bytes.Equal` future-couples the test to opacity-as-principle (decoupled from current spec); `len == 0` couples to RFC 9110 §15.3.5 + ADR 0005. Rule of thumb: test the invariant the ADR names, not the mechanism that currently realizes it. When mechanism and invariant collapse to the same assertion, prefer the simpler one. Chose `len == 0`.
  - **`fakeStore.Delete` mirrors real `*Storage.Delete` (returns nil for any 0-rows-affected case), not "ErrNotFound on miss".** Rationale: real Storage's `UPDATE ... WHERE id = $1` returns nil for never-existed (UPDATE 0 rows is not an error) and ErrNotFound only for malformed UUID (via `translatePgError` on real 22P02). fakeStore can't easily detect malformed without UUID parsing; mirroring the dominant case (return nil) keeps the fake honest. Drift class: handler's `errors.Is(err, ErrNotFound)` swallow branch is unreachable from fakeStore — covered by integration ring's 22P02 test, not by the unit suite. Documented in the fakeStore doc comment.
  - **500 path deferred per ADR 0006 cons-accepted.** Storage-internal failures (connection-loss, lock-timeout, deadlock) are unreachable from fakeStore at Option 3A's no-fault-injection discipline. If 500 coverage becomes needed, the path is a separate `failingStore` test double (preserves Option 3A by adding a *different* fake, not extending fakeStore). Acceptable at L02 scale: those failure modes are ops concerns, not contract concerns.
  - **Integration ring deferred to next session.** Three named tests scaffolded with suggested shapes; needs `DATABASE_URL_TEST` provisioning decision before implementation. Mixing test-DB infra setup + writing tests + verifying real-driver behavior into the tail of a session that already shipped 22 tests pressure-loads three things at once.
- **Architectural notes worth preserving:**
  - **fakeStore is the executable spec of what handler expects from storage**, not a mirror of real Storage internals. Drift between fake/real is acceptable up to the point integration ring backstops it. Update fake when contract changes (interface signature, new domain concept, new domain error sentinel); don't update for SQL/driver bugs or implementation refactors that preserve observable behavior. Naming the rule explicitly because it's the durable mental model future-readers need.
  - **Table-driven tests with identical expected output = "these paths MUST collapse" assertion.** `TestUpdate_NotFound` and `TestDelete_OpacityInvariant` both encode this — splitting them into separate per-case tests would lose the implicit equivalence claim. The table shape *is* the architectural assertion.
  - **ADR 0006 `errorMessage` helper survives a Shape 2 → Shape 3 migration as a one-function edit.** Every assertion call site uses the helper; migrating the envelope means changing one struct + one decode line, not sweeping N tests. The `bytes.Equal` discussion above applies the same principle to `TestDelete_OpacityInvariant`'s body assertion — chose to lock to the current spec instead of the principle, but the helper-vs-call-site design lets the rest of the suite migrate cheaply.
- **Commits:**
  - `b353660` — ADR 0006 accepted.
  - `39103d8` — test scaffold (fakeStore + helpers + opacity-invariant TODO + integration stubs).
  - `b7577b8` — `TestDelete_OpacityInvariant` shipped (Daniel-written).
  - `0381e4c` — Create branch coverage.
  - `aac4ac0` — Update branch coverage.
  - `b7ce027` — List + GetByID coverage.
- **Blocked:** None. Branch ahead of origin.
- **Next session target:**
  - **Warm-up #3 tail: integration ring.** Provision `DATABASE_URL_TEST` (option A: separate `flux_test` DB; option B: reuse dev DB and accept TRUNCATE clobbering the 13 smoke-test rows). Implement the three integration tests already scaffolded — `TestStorage_GetByID_22P02_to_NotFound` is the architecturally distinct one (only real driver round-trip produces a real `*pgconn.PgError`; fakeStore can't simulate it).
  - Run via `go test -tags=integration ./internal/workflows/`. Default `go test ./...` continues to skip integration tests.
  - **Then warm-up #4:** `project/docs/architecture.md` stub. Document L01 baseline + L02 additions (httpx, ADR 0003 boundary translation, soft-delete pattern, test pyramid shape). The build-tag run command lands here.

## [2026-05-10] Session: warm-up #3 tail (integration ring) + warm-up #4 (architecture posture-doc + ADR catalog). L02 warm-ups arc closes.

- **Did (warm-up #3 tail — integration ring):**
  - Provisioned `flux_test` DB. Option A (separate physical DB) over Option B (reuse dev DB with TRUNCATE clobber). Posture: physical isolation, mirrors CI shape, matches what L03's CI ADR will codify. The "30s unblock" of Option B trades against every future smoke-test re-seed; friction moves, doesn't disappear.
  - Step 8 added to `infrastructure/setup-guides/02-postgres-local.md` — `CREATE DATABASE flux_test OWNER flux` + schema apply via `psql -f project/schema.sql` + `DATABASE_URL_TEST` DSN appended to `.env.local`.
  - Implemented 3 named tests in `project/internal/workflows/storage_integration_test.go` (build tag `//go:build integration`):
    - `TestStorage_Update_RETURNING` — id + created_at immutable across Update; updated_at server-bumped via `now()`. 2ms sleep before Update clears Postgres microsecond-resolution race risk (without it, fast hardware can produce identical timestamps for back-to-back transactions and break `.After`).
    - `TestStorage_GetByID_22P02_to_NotFound` — ADR 0003 end-to-end against the live driver. Daniel implemented the assertion as a Learn-by-Doing. Picked contract-only via `errors.Is(err, ErrNotFound)`. Pathway-lock via `errors.As` against `*pgconn.PgError` was the second LbD option but ruled out — `translatePgError` returns `ErrNotFound` raw (no wrap chain). Daniel surfaced this himself by reading `translatePgError` before writing the assertion: *"What wrapped error, we are not wrapping anything?"* The choice itself surfaces an architectural property: opacity posture extends to the Go error chain, not just the wire.
    - `TestStorage_Delete_SoftDeleteFilter` — soft-deleted row excluded from both `GetByID` (returns `ErrNotFound`) and `List` (filtered out). Catches a future query helper that forgets to compose the `deleted_at IS NULL` filter on one of the read sites.
  - Build-tag discipline verified end-to-end: default `go test ./...` runs 22 handler tests + skips ring; `go test -tags=integration ./...` runs all 25.
  - Commit `3e5d125` (integration ring + flux_test setup).

- **Did (warm-up #4 — architecture doc + ADR catalog):**
  - Drafted `project/docs/architecture.md` with 8 sections (intro, system-shape diagram, module layout, posture, invariants, test pyramid, ADR index, deferred). Daniel mid-draft: *"Why is this needed?"* Honest re-evaluation surfaced that 6 of 8 sections duplicated existing artifacts — STATUS handles tactical state, ADRs hold decisions, project tree shows layout, and an ADR catalog earns its own conventional location.
  - Trimmed hard. Surviving doc (~33 lines) captures cross-ADR posture synthesis only:
    - **Opacity-on-wire** (ADRs 0003 + 0004 + 0005). Single design assertion: clients learn the contract, nothing else. Includes the wrap-chain-severance note linking back to the integration test choice above.
    - **Stdlib-purism** (ADRs 0001 + 0002 + 0006). Frameworks earn their way in via ADR; stdlib until it bites.
  - Considered an Invariants section + 8 candidate invariants. Daniel asked *"What does this serve?"* Honest evaluation: a flat list of invariants is mostly redundant with source ADRs (each lives in source ADR's Decision section). Reframe to invariant→test map was proposed (genuinely unique value: surfaces structural-only invariants that have no automated backing) but deferred to a future `tests/coverage.md` if pressure justifies. Section dropped entirely.
  - Created `adrs/README.md` as the catalog. Table indexes 0001-0006 with title / status / date / one-line. When-to-write-an-ADR criteria pulled from `.claude/rules/collaboration.md`. Numbering convention pinned (sequential, zero-padded, never reused; superseded ADRs add `**Status:** Superseded by NNNN`).
  - Commit `a1cf233` (architecture posture doc + ADR catalog).

- **Decided:**
  - **Test-DB convention: separate `flux_test` (Option A).** Physical isolation, mirrors CI shape, matches what L03 will codify.
  - **22P02 integration assertion: contract-only via `errors.Is`.** Pathway-lock via `errors.As` ruled out because `translatePgError` returns `ErrNotFound` raw. The choice itself surfaces opacity-as-design — not just the wire is opaque, the Go error chain is too.
  - **architecture.md scope: cross-ADR posture synthesis only.** Honest pruning — 6 of 8 drafted sections dropped after recognizing they duplicated existing artifacts. Surviving section is the only artifact in the repo capturing how 0003+0004+0005 share one design assertion.
  - **No invariants section.** A flat list duplicates source-ADR Decision sections without synthesizing. The invariant→test map reframe defers to a future `tests/coverage.md`.
  - **ADR catalog at `adrs/README.md`.** Conventional location — readers find ADR navigation directly inside `adrs/` without crossing module boundaries.

- **Architectural notes worth preserving:**
  - **Doc-pruning discipline.** A doc that paraphrases ADRs and restates module layout becomes a liar within 2-3 PRs. Architecture docs earn their slot only by capturing what no other artifact does. Apply this filter *before* writing — but if writing-then-pruning is what surfaces the filter, that's a valid path too. The 6/8 prune ratio for this doc is itself the lesson.
  - **Wrap-chain severance is a feature.** `translatePgError` returning `ErrNotFound` raw (vs `fmt.Errorf("...: %w", ErrNotFound)`) is the design choice that makes opacity hold across the Go error chain. Wrapping would leak the SQLSTATE to any caller doing `errors.As`. Same posture as ADR 0005's DELETE 204 — callers learn the contract, nothing else. Now explicitly named in `architecture.md`'s opacity-on-wire section.
  - **LbD design as architectural surface.** The 22P02 LbD's two options weren't symmetric — option 2 (pathway-lock) revealed a property of the system (the severed wrap) when attempted. Daniel got there by reading; could equally have got there by writing the failing assertion. Either path surfaces the same architectural fact. LbDs that bake in this kind of asymmetric reveal are higher-yield than ones with two equally-valid answers.
  - **Synthesis-vs-redundancy filter.** `architecture.md` survived only as the cross-ADR synthesis layer. ADRs are source-of-truth for decisions; STATUS is tactical state; project tree is layout; `adrs/README.md` is navigation. The 5th artifact (this doc) survives only by doing what those four can't do alone — name patterns that span them. Useful filter to apply when proposing future docs.

- **Commits this session:**
  - `3e5d125` L02 warm-up #3 tail: integration ring + flux_test setup.
  - `a1cf233` L02 warm-up #4: cross-ADR posture synthesis + ADR catalog.
  - (this commit) L02: bookkeeping — warm-up #4 closes (L02 warm-ups arc done).

- **Blocked:** None.

- **Next session target: L02 proper.**
  1. **Draft `auth-model` ADR.** Mechanism choice (JWT vs session-cookie vs API-key vs basic-auth) × auth-vs-authz boundary × dependency posture (stdlib-only vs introduce a library — ADR 0001/0002/0006 baseline says frameworks earn their slot). Compatibility constraint: presence-disclosure (404 opacity, ADRs 0003/0004/0005 cluster) must hold post-auth. The 401 challenge composes with existing 404 behavior — name this tension up front in the ADR Context section.
  2. **Then `tenancy-isolation`.** Schema column (`tenant_id` on workflows) + WHERE-clause discipline at every query site + cross-tenant access returns 404 (not 403 — opacity preserved). Decisions stack on auth ADR's tenant-id-on-context output.
  3. **Implementation after both ADRs accepted.** Order shaped by ADRs — middleware likely first (provides tenant_id in context), then column migration + WHERE-clause sweep across storage methods. Test pyramid extends naturally per ADR 0006: fakeStore tracks tenant state; integration ring gets one tenancy-isolation test.
