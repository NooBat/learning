# Architectural Framing Over Code Examples

## Rule

When explaining design trade-offs, writing ADR reasoning, or comparing libraries, tools, or dependency options, **lead with architectural posture — not code snippets**. Frame the discussion in terms of coupling intent, abstraction layers, domain-shape alignment, forward compatibility, and what abstraction tax is being accepted or rejected. Use code examples only when the syntax itself is the crux of a trade-off, and even then treat snippets as *evidence for* an architectural point — not the point itself.

## When this applies

- Writing ADR sections (Context, Options Considered, Decision, Consequences).
- Comparing dependency options (driver, framework, library, tool).
- Explaining why a boundary is drawn where it is (which layer owns what).
- Code review: lead with structure (coupling, layering, ownership, invariants) before syntax (naming, idioms, style).
- Teaching concepts: conceptual level first; code follows only as evidence.

## How to apply

- Open with the *posture* each option implies (e.g., "Postgres-first vs SQL-generic", "stdlib purism vs framework ergonomics", "DB-layer invariants vs application-layer flexibility").
- In ADR Decision sections: state the tipping architectural reason and one con explicitly accepted. Not *"library X has method Y"*.
- When code is genuinely needed ("here's what this costs at every call site"), keep snippets minimal and anchor them to a structural claim.
- In reviews, structure comments as: structural observations first (boundaries, coupling, invariants, data flow) → syntax/idiom observations second.
- When Daniel asks for a comparison, default to 3-5 sentences of architectural framing over a long code table. If the code would genuinely help, offer it as a follow-up: *"want me to also show the syntax-level differences?"*

## Anti-patterns

- ❌ Opening a design comparison with side-by-side code tables; architectural framing gets buried.
- ❌ Decision reasoning phrased only in API-ergonomics terms ("it has nicer helpers") without naming the underlying architectural trade-off.
- ❌ Reviews that start with naming/style and never reach structural observations.
- ❌ Treating code examples as the *goal* of an explanation rather than *evidence* for a structural point.

## Why this rule exists

Daniel is learning to be a Solutions Architect. The primary skill he's building is architectural judgment — syntax is Googleable; judgment isn't. This rule complements the **North Star** in `CLAUDE.md` (*"this journey is about designing systems and understanding structure, not coding throughput"*) and keeps every explanation aligned with that goal.

**Precedent:** 2026-04-18 during ADR 0002 (postgres driver). Claude produced a heavy side-by-side code comparison of `database/sql` vs `pgx` native. Daniel corrected: *"we don't need the code, just the architectural mindset for it."* This rule exists to prevent that failure mode from recurring and to make the correction durable across sessions, machines, and LLMs.
