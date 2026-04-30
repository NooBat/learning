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
