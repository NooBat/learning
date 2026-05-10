# Architecture — cross-ADR synthesis

> What lives here: cross-ADR posture synthesis — design assertions
> that span multiple ADRs and don't belong to any single one.
>
> What lives elsewhere: tactical state in `STATUS.md`, decisions in
> `adrs/`, module layout in the project tree (`tree project/internal/`),
> ADR catalog in `adrs/README.md`. This doc does not duplicate them.
>
> Update on ADR acceptance, level transition, or when a posture
> becomes a lie.

## Posture

### Opacity-on-wire (ADRs 0003 + 0004 + 0005)

External response shape reveals only what the contract names; mechanism does not leak.

- **ADR 0003** — malformed UUID translates to `ErrNotFound` at the storage boundary. Surfaces as 404 — same response as a missing row. Driver-level SQLSTATE 22P02 stays below the boundary.
- **ADR 0004** — Shape 2 JSON envelope `{"error": "..."}`. HTTP status owned by the handler; domain errors stay transport-agnostic. Migrating to Shape 3 edits one function (`httpx.WriteError`).
- **ADR 0005** — PUT strict / DELETE idempotent 204. DELETE collapses {existing, just-deleted, never-existed, malformed-UUID} into byte-identical 204.

Single design assertion across the three: **clients learn the contract, nothing else.** The wrap chain at the storage boundary (`translatePgError`) is severed by design — this is why the integration test for 22P02 asserts only `errors.Is(err, ErrNotFound)`, not the wrapped driver error (which doesn't exist).

### Stdlib-purism (ADRs 0001 + 0002 + 0006)

Frameworks earn their way in via ADR. Baseline:

- **ADR 0001** — `net/http` + `http.ServeMux` (Go 1.22 method+pattern routing). No chi / gin / echo.
- **ADR 0002** — `pgx/v5` is the only departure from stdlib at the persistence layer. Justified by Postgres-first posture (ADR 0002 names this explicitly vs SQL-generic).
- **ADR 0006** — `testing` + `net/http/httptest` + `encoding/json`. No testify, no generated mocks.

Posture for visibility, not asceticism. The first time it bites, trade for a framework via a new ADR.
