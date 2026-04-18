# 0001. HTTP router choice

**Status:** Accepted
**Date:** 2026-04-18
**Level:** L01 (mvs)

## Context

L01 requires an HTTP server exposing 5 CRUD endpoints for workflow resources:

- `POST /workflows`
- `GET /workflows`
- `GET /workflows/{id}`
- `PUT /workflows/{id}`
- `DELETE /workflows/{id}`

Each endpoint needs method-based routing and a path parameter (`{id}`). The router choice shapes:

- **Import surface** — stdlib only vs external dep.
- **Handler signature** — canonical `func(http.ResponseWriter, *http.Request)` vs framework-specific types.
- **Middleware idiom** — `http.Handler` wrapping vs framework chain API.
- **Migration ease** — how hard it is to change course at L02/L03 when we add auth, structured logging, request IDs, and eventually rate limiting and circuit breakers.

Since L01 is the first Go project in this repo, this decision also sets the tone for *how idiomatic-Go* the codebase will feel. The collaboration rule states: *"Go idioms, not translations … Prefer stdlib when reasonable."*

## Options Considered

### Option A: `net/http` stdlib (Go 1.22+ ServeMux)

- **Pros:**
  - Zero dependencies — nothing to learn beyond the stdlib itself.
  - Go 1.22 (Feb 2024) added method-prefixed patterns and path wildcards:
    - `mux.HandleFunc("POST /workflows", handler)` — method matching in-pattern.
    - `{id}` params read via `r.PathValue("id")`.
  - Handler signature is `func(http.ResponseWriter, *http.Request)` — matches every Go tutorial and book.
  - Middleware is "just a function that takes an `http.Handler` and returns one" — no framework-specific chain API.
  - Teaches the canonical model first; any framework you adopt later just wraps this.
- **Cons:**
  - Middleware composition is verbose (nested function calls or a small helper).
  - No built-in route groups; you emulate with multiple muxes or a helper func.
  - Pattern matcher is less expressive than dedicated routers (no regex, no constraints).
- **Notes:** Since 1.22, ecosystem momentum has been pushing new projects *back* toward stdlib. Aligns with the repo rule "prefer stdlib when reasonable."

### Option B: `chi` (go-chi/chi)

- **Pros:**
  - Lightweight (~1500 LOC). Idiomatic Go style throughout.
  - 100% compatible with `http.Handler` / `http.HandlerFunc` — no custom context type.
  - Clean route groups and middleware chains: `r.Route("/workflows", func(r chi.Router) { ... })`.
  - Widely used in production Go APIs (good hiring signal later).
- **Cons:**
  - An external dep to track and update.
  - One more mental model (the `chi.Router` interface), though it mirrors stdlib closely.
- **Notes:** Often called the "sweet spot" between stdlib purism and framework ergonomics. `go-chi/chi/v5` is the current import path.

### Option C: `gin` (gin-gonic/gin)

- **Pros:**
  - Very fast routing (radix tree).
  - Large ecosystem with built-in helpers (JSON binding, validation, render funcs).
- **Cons:**
  - Custom `gin.Context` — handlers are `func(*gin.Context)`, *not* `http.Handler`. This locks you into Gin's middleware ecosystem and makes swapping routers painful.
  - Heavier; includes templating, file serving, and other features unused at L01.
  - Philosophy clashes with idiomatic Go — more Express.js-flavored.
- **Notes:** Popular but the custom-context lock-in is the crux. Harder to "unuse" once adopted.

### Option D: `gorilla/mux`

- **Pros:**
  - Regex path patterns and host matching — most expressive of any option.
  - `http.Handler` compatible.
- **Cons:**
  - Library was archived in Dec 2022, resurrected by new maintainers in 2023, but momentum remains below `chi` and stdlib.
  - Its unique features (regex routes, host matching) aren't needed at L01.
  - Most new Go projects have moved to `chi` or stdlib.
- **Notes:** Included for comparison, but essentially a legacy choice today.

## Decision

**Chosen:** A — `net/http` stdlib (Go 1.22+ ServeMux).

The HTTP model is foundational, and the goal at L01 is to learn the canonical `http.Handler` pattern before layering any abstraction over it. Crucially, this choice is *cheaply reversible*: both stdlib and `chi` compose around `http.Handler`, so if middleware ergonomics become painful at L03, migrating is mechanical — swap `http.NewServeMux()` for `chi.NewRouter()` and every handler keeps its existing signature. Picking `gin` or `echo` would close that escape hatch since their custom context types aren't `http.Handler`. I'm explicitly accepting that middleware composition will be more verbose until L03, and that I'll likely write a small `chain(handlers ...func(http.Handler) http.Handler) http.Handler` helper once logging, request IDs, and recovery start stacking up.

## Consequences

**Positive** (apply to any choice):

- L01 exit criterion (5 working CRUD endpoints over `curl`) is reachable in a weekend with any option.
- The ADR itself becomes the first artifact of the "decision before implementation" rhythm — precedent for future levels.

**Negative** (to finalize once the Decision is made; option-conditional):

- If A (stdlib): middleware composition gets verbose at L03 when we add request IDs + structured logging + recovery; revisit there.
- If B (chi): one external dep to keep current; small cognitive overhead for the `chi.Router` interface.
- If C (gin): handler signatures diverge from stdlib; migrating away later means rewriting every handler.
- If D (gorilla): adopting a library with reduced community momentum; future contributors may not be familiar.

**Neutral:**

- Any option produces roughly similar L01 code: one `main.go`, 5 handler funcs, ~200 lines of Go total.

## Revisit triggers

- **L02** — adding auth middleware. If chain composition feels painful with the chosen option, reconsider.
- **L03** — adding structured logging (`log/slog`) + request ID middleware + graceful shutdown. The middleware stack grows; reevaluate if ergonomics suffer.
- **L07** — resilience (rate limiting, circuit breakers, timeouts at the HTTP layer). Pre-built middleware availability in the chosen ecosystem matters here.
- **Anytime** — if routing becomes a measurable bottleneck (unlikely before ~thousands of req/s).

## References

- [Go 1.22 release notes — enhanced routing patterns](https://go.dev/blog/routing-enhancements)
- [`net/http.ServeMux` docs](https://pkg.go.dev/net/http#ServeMux)
- [go-chi/chi](https://github.com/go-chi/chi)
- [Gin](https://gin-gonic.com/)
- [gorilla/mux](https://github.com/gorilla/mux)
- Repo rule: `.claude/rules/collaboration.md` — *"Go idioms, not translations … Prefer stdlib when reasonable."*
