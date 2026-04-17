---
name: bootstrap
description: Orient a cold Claude session with full context from this learning repo. Reads ABOUT-ME.md, STATUS.md, LEVELS.md, and tail of LOG.md, then outputs a structured briefing of who Daniel is, what level he's on, what the last session accomplished, and what the next target is. Use when starting a fresh session (especially from a new account, machine, or LLM where conversation history is missing), when the user says "get up to speed" or "catch up," or when you need a ground-truth snapshot before taking any action.
---

# Bootstrap

Goal: in one compact output, give any agent (you, another Claude session, a different LLM) enough context to be useful in this repo without re-reading everything.

## Step 1: Read the canonical files

Read these, in order:

1. `ABOUT-ME.md` — Daniel's profile, goals, constraints, learning preferences
2. `STATUS.md` — current level, this-week focus, blockers, next-session target
3. `LEVELS.md` — full level map with completion status
4. Tail of `LOG.md` (last ~100 lines) — recent session history

Claude Code auto-loads `CLAUDE.md` and `.claude/rules/*.md` at session start, so no need to read those again unless clarifying a specific rule.

## Step 2: Output the briefing

Format:

```markdown
## Bootstrap summary

**Who:** [1 sentence about Daniel — role, goal, time budget]

**Where we are:** [current level from STATUS.md, with 1-sentence goal restatement]

**Last session:** [most recent LOG.md entry, 2-3 sentences]

**Blockers / open questions:** [from STATUS.md or recent LOG entries, if any]

**Next target:** [from STATUS.md — the concrete thing to do next]

**Active rules to remember:**
- [Any non-obvious rules from recent context — e.g., pending ADR, in-progress refactor, rule amendment proposed]
```

Keep the summary under 200 words. Longer = more context burned for no gain.

## Step 3: Ask before acting

End with one question: *"Ready to continue with [next target]? Or something else?"*

Do not start implementing until Daniel confirms the direction. The point of bootstrap is orientation, not action.

## Anti-patterns

- ❌ Dumping the full file contents instead of summarizing.
- ❌ Reading ROADMAP.md at bootstrap time (it's reference material, read on demand).
- ❌ Launching into work without confirming direction.
- ❌ Using this skill mid-session when you already have context — it's wasteful.
