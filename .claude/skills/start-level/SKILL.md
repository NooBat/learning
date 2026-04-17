---
name: start-level
description: Scaffold a transition to a new level. Updates LEVELS.md (marks previous level [x], new level [~]), creates a level brief at levels/LXX-slug.md from template, updates STATUS.md to point at the new level, and appends a [LEVEL-TRANSITION] entry to LOG.md. Use when Daniel is moving from one level to the next.
argument-hint: <level-number> <slug>
disable-model-invocation: true
---

# Start a new level

Scaffolds a level transition. Daniel invokes this explicitly when he's ready to move to the next level — it has side effects (file edits, level brief creation), so it's user-only.

**Usage:** `/start-level 02 auth-tenancy`

- `$0` = level number (e.g., `02`)
- `$1` = slug (e.g., `auth-tenancy`) — lowercase, hyphenated

## Step 1: Verify the transition

Before writing anything:

1. Read `LEVELS.md` — confirm the previous level is actually complete (has exit criteria met).
2. Read `STATUS.md` — confirm the current level matches what we think it is.
3. If exit criteria aren't met or there's uncertainty, **stop and ask Daniel** instead of scaffolding.

## Step 2: Update LEVELS.md

- Mark the previous level `[x]` (done)
- Mark the new level `[~]` (in-progress)
- Leave later levels as `[ ]` (open)

## Step 3: Create the level brief

Create `levels/L$0-$1.md` with this template:

```markdown
# L$0 — [Human-readable title]

## Goal

[1-2 sentences: what capability the keystone project gains at this level]

## Why this level exists

[1 paragraph: what concept this level teaches and why it matters on the SA path]

## Prerequisites

- [Previous level's exit criteria, stated]
- [Any specific tooling or reading that should be done first]

## Exit criteria

Concrete, verifiable items. Each should be checkable as done/not done.

- [ ] [Code-level criterion — e.g., "POST /workflows returns 201 with created resource"]
- [ ] [Operational criterion — e.g., "graceful shutdown works on SIGTERM"]
- [ ] [Test criterion — e.g., "integration test covers happy path"]
- [ ] [Documentation criterion — e.g., "architecture.md updated with the new component"]
- [ ] [ADR written for any architecturally significant decision]

## Scope

What's IN:
- [explicit list]

What's OUT (anti-scope):
- [explicit list — things tempting but deferred to later levels]

## Reading (triggered by this level)

- [Book/chapter from ROADMAP.md reading list — only what's relevant]
- [Official docs link if critical]

## ADR-worthy decisions likely to come up

- [e.g., "Which HTTP router? (chi vs net/http vs gin)" — will need /write-adr]

## Stretch (optional, only if flying through)

- [Items that extend the level without expanding scope]
```

After filling the template, leave the exit criteria specific and Daniel-editable — Claude drafts them, Daniel refines.

## Step 4: Update STATUS.md

Replace the "current level" section to point at L$0. Update the "next-session target" to an opening step for this level (e.g., "Read levels/L$0-$1.md and set up the first task").

## Step 5: Append to LOG.md

```markdown
## [YYYY-MM-DD] [LEVEL-TRANSITION] L<prev> → L$0

- **Completed:** L<prev> — [1-sentence summary of what was shipped]
- **Started:** L$0 ($1) — [1-sentence goal]
- **Key artifacts from L<prev>:** [pointer to any new ADRs, architecture doc sections, tests]
- **Carry-over from L<prev>:** [anything not finished that will affect L$0, or "none"]
```

## Step 6: Summarize for Daniel

Output:

```
Scaffolded L$0 — $1.

- Level brief: levels/L$0-$1.md (draft — review exit criteria before starting)
- STATUS.md updated
- LEVELS.md updated
- LOG.md transition entry added

Next: read the level brief, refine the exit criteria if needed, commit the scaffold.
```

Remind Daniel to commit: `git add -A && git commit -m "scaffold L$0: $1"`.

## Anti-patterns

- ❌ Auto-confirming the previous level is complete without checking exit criteria.
- ❌ Writing the entire level brief without leaving it Daniel-editable (he should refine the exit criteria himself).
- ❌ Pre-writing levels further than the one being started (that violates the "one level ahead" rule).
- ❌ Skipping the LOG entry.
