---
name: session-end
description: Run the session-end protocol. Updates STATUS.md with current level and next-session target, appends a dated summary entry to LOG.md (what was done / decided / blocked / next), and reminds Daniel to commit and push. Use when wrapping up a session, or when Daniel says "let's stop here" or "we're done for today."
---

# Session End

Runs the session-end protocol defined in `.claude/rules/session-protocol.md`. Can be invoked by Daniel (`/session-end`) or triggered by Claude when the session is clearly wrapping up.

## Step 1: Snapshot the session

Review this session's work. What happened?

- What concrete things got done (commits, files created/edited, tests passing)?
- What decisions were made (and is there an ADR or should there be)?
- What's blocked or waiting?
- What's the logical next step for the next session?

If any of these are unclear, ask Daniel before writing.

## Step 2: Update STATUS.md

Overwrite `STATUS.md` so it reflects the new state. STATUS.md should stay small (~30 lines) — old content gets replaced, not accumulated.

Template:

```markdown
# STATUS

**Current level:** L<NN> — <slug>
**Last updated:** YYYY-MM-DD

## This-week focus

[1-3 bullets on what we're working on right now at this level]

## Next-session target

[One concrete, actionable thing the next session should start with. Be specific: "Write the POST /workflows handler with schema validation" beats "continue L01".]

## Open blockers

- [Anything waiting on Daniel, a tool, a decision, etc. Or "none."]

## In-flight ADRs

- [List any ADRs in "Proposed" status that need Daniel's call. Or "none."]
```

## Step 3: Append to LOG.md

Append a dated entry. Do not overwrite LOG.md — it's append-only.

Template:

```markdown
## [YYYY-MM-DD] Session summary

- **Did:**
  - [Concrete item 1]
  - [Concrete item 2]
- **Decided:**
  - [Decision with pointer to ADR if applicable, or "none"]
- **Blocked:**
  - [Blocker or "none"]
- **Next:** [Next-session target, same as STATUS.md]
```

If an ADR was proposed or accepted this session, include the number.

If a level transition happened, tag the entry `[LEVEL-TRANSITION]`. If it was a retro, tag `[RETRO]`.

## Step 4: Remind to commit and push

Output:

```
Session wrapped up.

- STATUS.md updated — next session starts with: [next target]
- LOG.md appended with today's summary
- Commit + push checklist:

  git status
  git add -A
  git commit -m "<descriptive message about the session's work>"
  git push

Per the "push or it didn't happen" rule, don't skip the push. If you're not pushing, the session's state won't be visible to other environments (other Claude accounts, other machines).
```

## Step 5: Flag rule violations (if any)

If Claude broke a rule this session (or almost did), note it at the end:

```
Heads-up: [specific violation or near-miss]. Consider whether the rule needs amending or whether this was a one-off.
```

This keeps the rules honest over time.

## Anti-patterns

- ❌ Writing STATUS.md without replacing the stale content — it grows and becomes useless.
- ❌ Writing a LOG.md entry that's vague ("worked on L01"). The log should be scannable — specifics are what make it useful.
- ❌ Skipping the commit/push reminder. The rule is constitutional.
- ❌ Running this skill mid-session when you're just taking a break, not ending. Use it only when the session is actually done.
