# 0006. Test strategy (pyramid shape + assertion model)

**Status:** Accepted
**Date:** 2026-05-01
**Level:** L02 (auth-tenancy)

## Context

L02 warm-up #3: write the first Go test suite. Three architectural decisions stack, each load-bearing for the rest of the level (and forward to L03+).

Forces converging:

- **Consumer-defined `storage` interface (L01).** The unexported `storage` interface in `handler.go` was deliberately chosen over a producer-exported `IStorage` because it doubles as a test boundary. A `fakeStore` satisfying the interface is the L02 design payoff for that L01 decision. This ADR cashes the affordance.
- **Pairing B opacity invariant (ADR 0005).** DELETE on existing / just-deleted / never-existed / malformed-UUID returns byte-identical 204 No Content. Verified via runtime smoke this session, but smoke is one-time. Promoting it to an automated assertion catches regressions when future code drifts (e.g., a refactor breaks 22P02 translation; a logging change leaks a body on 204; a migration to Shape 3 envelope changes 4xx body shape).
- **Envelope shape stickiness (ADR 0004).** Shape 2 (`{"error": "..."}`) → Shape 3 (`{"error": {"code": ..., "message": ...}}`) was named as a breaking-change risk. Tests that lock onto raw bytes calcify Shape 2 into the test suite; tests that decode the envelope through a helper migrate by editing one function.
- **22P02 translation rule (ADR 0003).** Lives in `translatePgError` at the storage boundary. The rule is unit-testable in isolation, but only meaningful end-to-end (does a malformed UUID actually produce a 404 at the wire?). Two assertion levels for one rule.
- **Test count growth.** L02 ships ~5 endpoints (Create / GetByID / List / Update / Delete). L03 adds runs (likely +5). L04+ adds auth + tenancy variants (multiplicative). Test pyramid shape needs to scale; choosing wrong now means rewriting at L04+.

Decision scope: three independent axes.

1. **Pyramid shape.** What test layers exist and how they share coverage (handler-level fakeStore vs storage-level real Postgres vs end-to-end).
2. **Assertion strategy for response bodies.** Raw-byte compare vs envelope-decode helper.
3. **`fakeStore` discipline.** What state the fake tracks; whether it includes fault-injection hooks; how realistic it must be relative to real `Storage`.

Out of scope: testing tools selection (default to stdlib `testing` + `httptest` — no third-party assertion libraries at L02; revisit if `t.Helper()` boilerplate becomes painful), property-based / fuzzing tests (defer until L05+ async work introduces real state-space concerns), benchmarks (L05+ when performance has meaning), CI integration (lands at L03 alongside the migration tooling decision).

## Options Considered

### 1. Pyramid shape

#### Option 1A — Handler-level only

`httptest.NewServer` (or `httptest.NewRecorder`) + `fakeStore`. No real Postgres in tests.

- **Pros:**
  - Fast — millisecond-scale per test, no DB roundtrip.
  - No test infrastructure (no test DB provisioning, no per-test fixtures, no schema reset).
  - Single dependency — `testing` stdlib only.
  - Tests run anywhere — laptops, CI, no Docker required.
- **Cons:**
  - SQL queries untested. Soft-delete WHERE clause typo → tests pass, prod fails.
  - 22P02 translation untestable (real `*pgconn.PgError` requires real driver).
  - `RETURNING` semantics untestable (fake can return whatever it wants).
  - Schema drift undetected — `deleted_at` column rename, hand-applied migrations get out of sync with code, fake stays consistent with handler expectations.

#### Option 1B — Storage-level only

Test files exercise `Storage` directly against a real Postgres instance. No handler tests.

- **Pros:**
  - Catches SQL bugs.
  - Validates `translatePgError` against real driver errors.
  - Schema/code drift surfaces on test run.
- **Cons:**
  - No HTTP-layer testing. Error ladder, status code, envelope shape all untested.
  - Test infra required — Postgres in CI, schema reset between tests, transaction rollback or truncate-tables discipline.
  - 10-100× slower than handler-level — DB roundtrip per test.
  - Misses opacity invariants — those are wire-level properties.

#### Option 1C — Handler-level primary + integration ring

Most tests at handler level using `fakeStore`. Small ring of integration tests against real Postgres covering only the things the fake cannot simulate. Build-tag separation (`//go:build integration`) — fast tests run by default, integration tests run on demand or in CI.

- **Pros:**
  - 80%+ tests fast and dependency-free.
  - SQL-level concerns (22P02, soft-delete WHERE, RETURNING) covered by the integration ring.
  - Industry-standard test pyramid shape.
  - Build tags let local dev skip slow tests; CI runs all.
- **Cons:**
  - Two test boundaries to maintain (fake + real). Risk: fake drifts from real Storage behavior, handler tests pass but production fails.
  - Build-tag discipline required — easy to forget integration test on a CI run.
  - Initial setup heavier — schema reset script, `DATABASE_URL_TEST` env var convention.

#### Option 1D — End-to-end only

`httptest.NewServer` driving real handler + real Storage + real Postgres. No fakes.

- **Pros:**
  - Highest fidelity — every test exercises the full stack.
  - No fake-vs-real drift class.
- **Cons:**
  - Slowest pyramid shape — every test pays DB roundtrip + HTTP roundtrip cost.
  - Test failure attribution is muddy — "where did this break?" requires bisection across layers.
  - Same test infra cost as 1B but applied to all tests, not just integration ring.
  - At L04+ scale (auth + tenancy variants), test runs become 10-minute affairs.

### 2. Assertion strategy

#### Option 2A — Raw-byte compare

```go
if got := w.Body.String(); got != `{"error":"workflow not found"}` + "\n" {
    t.Fatalf("body = %q", got)
}
```

- **Pros:**
  - Zero abstraction. Reader sees exactly what the wire returns.
  - One line per assertion.
- **Cons:**
  - Locks tests to Shape 2 envelope. Shape 2 → Shape 3 (ADR 0004 named this as a real future-migration risk) requires editing every test, not one helper.
  - Fragile to whitespace, trailing newlines, key ordering (json.Encoder produces deterministic ordering, but raw-byte tests still feel brittle to readers).
  - Repeats the envelope shape assumption at every call site; doesn't centralize it.

#### Option 2B — Decode + assert on parsed struct

```go
var got struct { Error string `json:"error"` }
if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil { t.Fatalf(...) }
if got.Error != "workflow not found" { t.Fatalf(...) }
```

- **Pros:**
  - Forward-compatible to envelope changes — decode struct evolves.
  - Asserts on semantic content, not byte-exact format.
- **Cons:**
  - Boilerplate per test — 3 lines × N tests.
  - Repeats the decode shape at every call site; first thing future-you will refactor anyway.

#### Option 2C — Envelope-decode helper

```go
// errorMessage decodes Shape 2 (ADR 0004). Migrating to Shape 3 changes
// only this function — every assertion call site keeps working.
func errorMessage(t *testing.T, body []byte) string {
    t.Helper()
    var env struct { Error string `json:"error"` }
    if err := json.Unmarshal(body, &env); err != nil {
        t.Fatalf("decode envelope: %v; body=%s", err, body)
    }
    return env.Error
}
```

Used as: `if got := errorMessage(t, w.Body.Bytes()); got != "workflow not found" { ... }`

- **Pros:**
  - Single source of truth for envelope shape. ADR 0004's noted Shape 2 → Shape 3 migration cost = editing one function.
  - `t.Helper()` makes test-failure stack point at the assertion site, not the helper.
  - Self-documenting — function name names the contract.
- **Cons:**
  - Adds a test-helper file/function before tests exist that need it. "Just in case" abstraction tax.
  - One more thing for a future contributor to learn before writing a test.

### 3. fakeStore discipline

#### Option 3A — Minimal map-backed fake

```go
type fakeStore struct {
    rows map[string]*Workflow
    deleted map[string]bool
}
```

Implements all five `storage` methods. No mutex. No fault injection. Tracks live rows and soft-delete state separately (cleaner than mixing into the Workflow type).

- **Pros:**
  - ~30 lines. Easy to read, easy to maintain.
  - Matches L02 scale — single test process, sequential test execution by default.
  - Doesn't pre-design for needs that haven't surfaced.
- **Cons:**
  - Cannot exercise concurrent-access bugs.
  - Cannot test 500-error paths (connection-loss class) without fault injection.
  - Diverges from real `Storage` over time — manual discipline to keep in sync.

#### Option 3B — Realistic concurrent-safe fake with fault injection

Add `sync.Mutex`, optional `forceErr error` field for fault injection, transaction-like semantics. Implements every observable behavior of real `Storage`.

- **Pros:**
  - Closer to real Storage; less drift.
  - Enables 500-path testing.
  - Concurrent test execution (`t.Parallel()`) safe.
- **Cons:**
  - 100+ lines. The fake becomes its own complexity surface.
  - Pre-engineering — L02 has no concurrency tests, no real-error-path tests.
  - Risk: the fake encodes assumptions that don't match real Storage semantics, and bugs leak through both.

#### Option 3C — Generated mock (mockery, gomock)

Library-generated mock from the `storage` interface signature.

- **Pros:**
  - No hand-written fake. Mock generation is mechanical.
  - Per-test method-call expectations (`store.EXPECT().GetByID(...)`).
- **Cons:**
  - New dependency. ADR 0001 / 0002 stdlib-purism posture argues against this at L02 scale.
  - Per-test mock setup boilerplate — heavier than a shared fake for simple cases.
  - Couples tests to method-call sequences, not behavior — refactors break tests for no semantic reason.
  - Generated mocks can't easily simulate stateful behavior (soft-delete + retrieval) without verbose `.DoAndReturn` plumbing.

## Decision

**Path 1C + 2C + 3A.**

**Tipping factor:** This ADR cashes the L01 design payoff (`storage` interface as test boundary) without pre-engineering for needs that haven't surfaced. 1C inherits the same stdlib-purism posture as ADRs 0001 / 0002 (`net/http`, `pgx`) — handler-level fakeStore covers what stdlib affords cheaply; the integration ring catches what only real Postgres can. 2C centralizes ADR 0004's envelope-shape assumption in one function so the noted Shape 2 → Shape 3 migration is a single-function edit, not a sweep across N tests. 3A keeps the fake minimal at L02 scale and accepts the "fake drifts from real" risk explicitly — the integration ring is the backstop for that drift.

**Cons accepted:**

- **Two test boundaries (fake + real) means drift risk.** The integration ring is the canary — when it fails on behavior the fake said was fine, the fake gets updated. At L02 the integration ring is small (~3 tests covering 22P02, soft-delete WHERE, RETURNING semantics) so the maintenance cost is low.
- **No fault injection at L02 means no 500-path tests for connection-loss class.** Acceptable: those failure modes are ops concerns, not contract concerns. If a real-world incident later surfaces a class of bugs that fault injection would catch, revisit.
- **Envelope helper "just in case" tax.** Accepted because the helper costs ~10 lines and lands the day the first test does — there's no period where tests exist without the helper.
- **Build-tag discipline.** `//go:build integration` could be forgotten in CI. Mitigated by documenting the run command in `project/docs/architecture.md` (warm-up #4) and adding a `make test-all` target if/when a Makefile lands.

**Composition with prior ADRs:**

- **ADR 0001 / 0002 (stdlib-purism):** test stack is stdlib-only — `testing`, `net/http/httptest`, `encoding/json`. No third-party assertion libraries (`testify`, `gomega`) at L02. Revisit only if the stdlib affordances become genuinely painful, not "not-as-nice."
- **ADR 0003 (22P02 translation):** unit-testable on `translatePgError` directly (storage-level, but in-process — synthesize a `*pgconn.PgError` and assert the return). Also exercised end-to-end via the integration ring (real malformed UUID → real 22P02 → real 404).
- **ADR 0004 (envelope shape):** assertion helper centralizes the Shape 2 contract. Migration to Shape 3 = edit `errorMessage` helper signature/body, every test still passes (or fails loudly with one clear failure mode).
- **ADR 0005 (pairing B opacity):** the four-DELETE-paths byte-identical assertion is the highest-value handler test in the suite. It exercises ADRs 0003 + 0004 + 0005 in one go.

## Consequences

**File layout:**

```
project/internal/workflows/
├── handler.go
├── handler_test.go          ← new: handler-level tests + fakeStore
├── storage.go
├── storage_integration_test.go  ← new: //go:build integration tag
├── types.go
└── (existing files)
```

`fakeStore` lives in `handler_test.go` (same package, can reference unexported types). If it grows past ~80 lines, split to `fakes_test.go` — defer until that pressure surfaces.

**`handler_test.go` shape:**

- One `fakeStore` type implementing the unexported `storage` interface.
- Test functions per handler method: `TestCreate_*`, `TestGetByID_*`, `TestList_*`, `TestUpdate_*`, `TestDelete_*`.
- Each test: spin up `httptest.NewServer(mux)` (or `httptest.NewRecorder` for unit-y tests with no HTTP transport — pick per test), exercise the endpoint, assert status + body via `errorMessage` helper for 4xx and full-decode for 2xx.
- One critical test: `TestDelete_OpacityInvariant` — fires DELETE on four ID classes (existing, just-deleted, never-existed, malformed) and asserts all four responses are byte-identical. The runtime smoke this session promoted to compile-time guarantee.

**`storage_integration_test.go` scope (initial):**

- Real Postgres connection via `DATABASE_URL_TEST` env var (separate from `DATABASE_URL` to avoid clobbering dev data).
- Per-test schema reset (truncate `workflows` table; cheaper than `DROP/CREATE`).
- Tests:
  1. `TestStorage_Update_RETURNING` — verify `Update` populates `id`/`created_at`/`updated_at` server-side.
  2. `TestStorage_GetByID_22P02_to_NotFound` — pass a malformed UUID, assert `ErrNotFound` returned (validates `translatePgError` against real driver error).
  3. `TestStorage_Delete_SoftDeleteFilter` — soft-delete a row, assert `GetByID` returns `ErrNotFound` and `List` excludes it.

Three tests cover the three behaviors handler-level fakes can't exercise. Add more only when a real bug surfaces that the existing ring missed.

**Build-tag discipline:**

```go
//go:build integration

package workflows
```

Default `go test ./...` skips integration tests. CI and "full check" runs use `go test -tags=integration ./...`. Document the command in `project/docs/architecture.md` (warm-up #4) and any future `Makefile`.

**Helper inventory (initial):**

```go
// errorMessage decodes the Shape 2 envelope (ADR 0004) and returns the
// message field. Migrating to Shape 3 (structured) updates this helper
// only — every assertion call site keeps working.
func errorMessage(t *testing.T, body []byte) string

// requireStatus is a one-liner that fails the test if status != want.
// Saves three lines per status assertion.
func requireStatus(t *testing.T, got, want int)

// decodeWorkflow JSON-decodes a 2xx response body into Workflow.
// Asserts decode success internally so call sites get just the value.
func decodeWorkflow(t *testing.T, body []byte) Workflow
```

Three helpers, ~20 lines combined. Lands in `handler_test.go` initially; split to `helpers_test.go` if it grows past 50.

**fakeStore initial shape:**

```go
type fakeStore struct {
    rows    map[string]*Workflow
    deleted map[string]bool
}

// Methods satisfying the unexported storage interface:
// Create, GetByID, List, Update, Delete.
```

No mutex (sequential tests), no fault injection (no 500-path tests at L02), no schema validation (validation belongs in real Postgres, not the fake). When concurrent test execution or 500-path testing becomes a need, revisit; do not pre-add.

**Reversibility per choice:**

- **Pyramid (1C → 1A or 1D):** mechanical. Drop integration ring (1A) or convert handler tests to end-to-end (1D). 1C → 1A is cheap; 1C → 1D is a per-test rewrite.
- **Assertion (2C → 2A):** mechanical reverse — inline the helper across call sites. Sticky if Shape 3 migration happened; reverting before that point is one PR.
- **fakeStore (3A → 3B or 3C):** 3A → 3B is incremental (add mutex, add fault-injection field). 3A → 3C is a rewrite + new dep (an ADR-worthy change at that point).

**Future ADR locks/opens:**

- **CI pipeline (L03):** the `make test-all` or equivalent will codify the build-tag run convention. This ADR pins the tag; the pipeline ADR pins the runner.
- **Test DB provisioning (L03 alongside migration tooling):** local Docker-Compose Postgres for `DATABASE_URL_TEST`; CI picks up its own. Out of scope here.
- **Property-based / fuzzing (L05+):** when async work introduces state-space concerns. This ADR doesn't preclude — adding a property-based test alongside the existing pyramid is additive.
- **Test data factories (L04+):** if test setup boilerplate grows past `fakeStore.Create(...)` calls (e.g., complex multi-row scenarios for tenancy), introduce a factory pattern via an ADR.
- **Auth-tenancy tests (next L02 ADR target):** the testing pyramid this ADR locks scales to auth + tenancy variants by adding tests at the same handler level — fakeStore extends to track tenant state, integration ring extends with one tenant-isolation test.
