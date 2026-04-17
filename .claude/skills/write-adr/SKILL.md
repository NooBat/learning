---
name: write-adr
description: Create a numbered Architecture Decision Record in adrs/ with the Context → Options Considered → Decision → Consequences template. Use whenever an ADR-gated decision comes up — a new load-bearing dependency, an architectural split, a schema migration, an auth change, anything security-touching, a major refactor, or a cloud resource decision. ADRs are written BEFORE implementation.
argument-hint: <slug>
disable-model-invocation: true
---

# Write an ADR

Creates a numbered Architecture Decision Record before implementing an ADR-gated decision. Enforced by the collaboration rule — pause, write the ADR, then implement.

**Usage:** `/write-adr use-chi-router`

- `$0` = slug (e.g., `use-chi-router`, `postgres-row-level-security`) — lowercase, hyphenated, descriptive

## Step 1: Determine the next number

List existing files in `adrs/`. Find the highest `NNNN` prefix. Next number = that + 1, zero-padded to 4 digits.

- `adrs/0001-foo.md` exists → next is `0002`
- `adrs/` is empty → next is `0001`

## Step 2: Create the ADR

Create `adrs/NNNN-$0.md` with this template:

```markdown
# NNNN. [Human-readable title]

**Status:** Proposed | Accepted | Superseded by [link]
**Date:** YYYY-MM-DD
**Level:** L<NN> (from STATUS.md)

## Context

[The forces at play. What problem prompts this decision? What constraints exist? What's the current state if something similar already exists?]

## Options Considered

### Option A: [name]

- **Pros:** [list]
- **Cons:** [list]
- **Notes:** [anything else]

### Option B: [name]

- **Pros:** [list]
- **Cons:** [list]
- **Notes:** [anything else]

### Option C: [name] (if applicable)

...

## Decision

**Chosen:** Option [X]

[1-2 paragraphs explaining WHY this option was chosen. Not just "it's better" — what specific trade-off tipped the choice.]

## Consequences

**Positive:**
- [what becomes easier / better]

**Negative:**
- [what becomes harder / worse — be honest]

**Neutral:**
- [what changes that isn't clearly good or bad]

## Revisit triggers

When should we reconsider this decision?

- [e.g., "if we hit 500 req/s sustained and this router becomes the bottleneck"]
- [e.g., "if the Go ecosystem standardizes on a different router"]

## References

- [Link to related ADRs, RFCs, blog posts, books, docs]
```

## Step 3: Fill as much as Claude can

Fill in the template based on the current conversation. Leave:

- **Status** as "Proposed" (not Accepted until Daniel explicitly approves).
- **Pros/Cons** as drafts — Daniel should refine.
- **Decision** as "TBD" if the decision hasn't actually been made yet; this is often the case, the ADR helps structure the decision process.

## Step 4: Flag it in LOG.md

Append a short entry to `LOG.md`:

```markdown
## [YYYY-MM-DD] [ADR-PROPOSED] NNNN — [title]

Pause on implementation. Decision doc drafted at `adrs/NNNN-$0.md`. Daniel to review options and set Status=Accepted before proceeding.
```

## Step 5: Output

Tell Daniel:

```
Drafted ADR adrs/NNNN-$0.md (Status: Proposed).

Review the Options Considered and make the call. When you decide:
- Update Status to "Accepted"
- Fill in the Decision section with the reasoning
- Commit: git add adrs/NNNN-$0.md && git commit -m "adr: NNNN <title>"

Then we can implement.
```

## Anti-patterns

- ❌ Writing the Decision section as if the choice is obvious. If it's obvious, maybe it doesn't need an ADR — but if it IS ADR-gated, it's because there's a real trade-off.
- ❌ Skipping the Consequences section. The "negative" part is where ADRs earn their keep — forcing honesty about what we're giving up.
- ❌ Silently implementing before Daniel marks the ADR Accepted.
- ❌ Writing only one option. An ADR with one option isn't a decision, it's a default.
