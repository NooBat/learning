---
name: retro
description: Run a retrospective on recent levels or a tier. Appends a structured [RETRO] entry to LOG.md covering what stuck, what didn't, motivation check, and possible pivots. Use every 3-4 levels cleared, at end of each tier, or whenever Daniel seems stuck/unmotivated for more than two sessions in a row.
---

# Retrospective

A retro is the main defense against silent drift in a 6-7 year learning plan. Do it regularly, honestly. A retro that says "all good, keep going" is fine — a retro that's skipped is not.

## When to run

- **Every 3-4 levels cleared** — short retro, 10 minutes.
- **End of each tier** (after L04, L08, L12, L16) — bigger retro, revisit ROADMAP.md.
- **Stuck or unmotivated for 2+ sessions** — immediate retro, no waiting.
- **After a significant failure** (level re-started, major scope change, pivot considered).

Daniel can invoke (`/retro`) or Claude can suggest running one when the conditions above apply.

## Step 1: Ask the retro questions

Go through these one at a time. Don't ask all at once — it becomes a survey.

1. **What stuck?** Which concepts from the recent levels do you feel you genuinely understand — could teach to a junior engineer? (Teach-back in retro form.)
2. **What didn't?** Which concepts feel shaky, forgotten, or glossed over? Flag them for revisit.
3. **Effort vs. value.** Did the time spent match the learning gained? Where was it worth it? Where was it wasted?
4. **The project itself.** Is the keystone project still interesting enough to sustain? Any part that's become a chore?
5. **Motivation check.** How's the enthusiasm? Is the 5-7 h/week sustainable or straining? Any life changes affecting it?
6. **Pivots or scope changes?** Anything in `ROADMAP.md` that no longer fits — books to drop/add, levels to reorder, a domain we should commit to?

Use teach-back style on question 1: if Daniel can't articulate a concept, add it to the "revisit" list.

## Step 2: Append the retro to LOG.md

Structured entry:

```markdown
## [YYYY-MM-DD] [RETRO] After L<NN> (or Tier N end)

### What stuck
- [List concepts/skills that feel solid]

### What didn't — revisit
- [Concept]: [why it's shaky, what to do — e.g., "re-read Chapter 4 of Learning Go", "redo exercise X"]

### Effort vs. value
- **Worth it:** [what paid off]
- **Overinvested:** [what took too long for what it taught]
- **Underinvested:** [what deserved more time]

### Project status
- [How Daniel feels about the keystone project right now]

### Motivation
- [Honest assessment — is this sustainable?]

### Actions
- [Concrete changes to apply: pivot, scope change, re-sequence levels, swap a book, change habits]
- [Link to any ADR created for a major pivot — use /write-adr if needed]
```

## Step 3: Apply changes

If the retro produces action items that affect the plan:

- **Book swaps** → edit `ROADMAP.md` reading list. Commit `roadmap: swap <old> for <new> — <reason>`.
- **Level reorder / scope changes** → edit `LEVELS.md` and the affected level briefs. Commit.
- **Project pivot** → invoke `/write-adr` and draft an ADR. Don't silently pivot.
- **Rule changes** → edit `.claude/rules/*.md`. Commit `rules: <what changed + why>`.

Changes to the plan are normal. The audit trail is git.

## Step 4: Set next-session target

Update `STATUS.md` with the next-session target informed by the retro. If the retro identified a concept to revisit, the next session might be "read Chapter 4 of Learning Go and redo the exercise" rather than "continue L<NN+1>."

## Anti-patterns

- ❌ Retro-as-performance: going through the motions without actually answering honestly. Better to skip than to fake it.
- ❌ Retro-as-meta-work: spending two sessions debating the retro instead of getting back to the project.
- ❌ "Everything's great" without interrogation. If nothing feels off, probe harder — are there concepts that haven't been stress-tested?
- ❌ Skipping the action items. A retro that doesn't change behavior is a journal entry, not a retro.
