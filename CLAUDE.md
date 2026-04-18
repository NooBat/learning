# Learning Repo — Claude Entry Point

> Auto-loaded by Claude Code at session start. Canonical entry for any LLM (Claude Code, Claude.ai, API, or other) working in this repo.

## What this repo is

A **6-7 year structured learning path** from Junior Frontend Engineer to Solutions Architect. Centered on one keystone **Go** project (a Developer Automation Platform — Zapier/n8n/Temporal-style) in `project/` that evolves through ~16 levels. Reading, certifications, and writing happen in support of the project.

**The repo is the memory.** Claude-side memory (`~/.claude/memory/`, `claude-mem`, per-account memory) is treated as disposable cache. Everything durable lives here and is pushed to GitHub.

## North Star

**This journey is about designing systems and understanding structure — not about coding throughput.** Code is a vehicle for architectural judgment; judgment is the artifact. When the work forks between *"does this surface a structural lesson?"* and *"does this add to coverage, velocity, or feature count?"*, the first wins. Target role is Solutions Architect, not Senior Engineer — every suggestion, level scope, review, and ADR should be filtered through this lens.

Practical consequences for any session:

- Prioritize design-level activities (ADRs, contracts, boundary decisions, schema shape, trade-off analyses) over implementation breadth.
- Reviews lead with structure (coupling, layering, ownership, invariants), not syntax.
- Learn-by-Doing requests target design decisions, not boilerplate.
- Reading list favors design references (DDIA, Release It!, Richards & Ford, postmortems) over language tutorials beyond what a level demands.
- If a level's exit criteria feel like checkbox coding rather than structural learning, flag it — the criteria need refining.

## Session Bootstrap

Claude Code auto-loads these:
- This file (`CLAUDE.md`)
- All `.claude/rules/*.md` files without `paths` frontmatter (constitutional rules)

Read these yourself after auto-load finishes:

1. `ABOUT-ME.md` — user profile, goals, constraints
2. `STATUS.md` — current level, this-week focus, blockers
3. tail of `LOG.md` — recent session history (~100 lines)
4. `LEVELS.md` — full level map (reference as needed)

`ROADMAP.md` is the long-form plan — read only when you need the *why* behind the structure.

## Directory Map

```
CLAUDE.md              ← You are here
ABOUT-ME.md            ← User profile
ROADMAP.md             ← 6-7 year plan (tiers, levels, reading list)
STATUS.md              ← Current focus (small, re-read often)
LEVELS.md              ← Checklist of all levels
LOG.md                 ← Append-only session log
.claude/
  rules/               ← The constitution (collab, session protocol, conflicts)
  skills/              ← Repetitive workflows: /bootstrap, /start-level, /write-adr, /session-end, /retro
  settings.json        ← (later) per-repo permissions
project/               ← THE keystone Go service ("flux" working name)
levels/                ← Level briefs, one file per level (LXX-slug.md)
adrs/                  ← Architecture Decision Records (NNNN-slug.md)
infrastructure/
  setup-guides/        ← Step-by-step guides Daniel follows himself
notes/                 ← Daniel's notes per concept
reading/               ← Book notes
```

## Quick Orientation

- **User:** Daniel Nguyen (khoi27012003@gmail.com). Confident in React/TS; basic backend. See `ABOUT-ME.md`.
- **Backend:** Go. **Frontend (when needed):** React/TS.
- **Cloud:** AWS from Tier 3; Fly.io at L04 for first deploy (lower friction).
- **Tracking:** level-based, not calendar-based. Advance on exit criteria only.
- **Git:** private GitHub remote. Push after every meaningful session or the work is lost to other environments.

## Commands

> Populate as the project grows. Each new operation should land here at the level it's introduced.

### Repo-level (available now)

```bash
# Session-end checklist (run before ending)
git status                    # anything to commit?
git add -A && git commit -m "..."
git push                      # push-or-it-didn't-happen rule
```

### Project-level (`project/`)

```
# To be added at L01 (Go service skeleton).
# Populate with: go run ./cmd/..., go test ./..., docker compose up, migrate, etc.
```

### Setup

All one-time setup lives in `infrastructure/setup-guides/` as numbered guides. Daniel follows them; Claude does not run them.

## Rules

Constitutional rules live in `.claude/rules/*.md` and **auto-load at session start** (per [Claude Code memory docs](https://code.claude.com/docs/en/memory#organize-rules-with-claude/rules/)). Always-apply rules have no frontmatter; path-scoped rules use a `paths` field.

Current rules (all always-apply):

- `collaboration.md` — what Claude does vs. what Daniel does
- `session-protocol.md` — session start/end protocol
- `conflict-resolution.md` — priority when instructions conflict
- `verify-claims.md` — verify against docs before asserting facts about Claude Code / tooling

### Path-scoped rules pattern (for future rules)

To scope a rule to specific files, add YAML frontmatter with `paths`:

```yaml
---
paths:
  - "project/**/*.go"
---

# Go style
- Prefer `pgx` over `database/sql` for Postgres.
- Use `log/slog` for structured logging.
```

These load only when Claude reads a matching file, keeping context lean. We'll add path-scoped rules as they become relevant (e.g., `go-style.md` at L01, `adr-format.md` for `adrs/*.md`).

## When to update this file

Update CLAUDE.md when any of these change:

- **Directory structure** — add/remove top-level directories in the Directory Map.
- **Project commands** — add new `go`/`docker`/`make` commands at the level they're introduced.
- **Bootstrap order** — if a new file belongs in the startup read path.
- **Critical orientation facts** — e.g., cloud target switches, primary language changes.

Do **not** update CLAUDE.md for:

- Level-specific details (those live in `levels/LXX-*.md`).
- Day-to-day progress (that's `STATUS.md` + `LOG.md`).
- Rule wording (that's `.claude/rules/*.md`).

## When in doubt

Ask Daniel. Don't assume. Anything that changes state (installs, migrations, deploys, dependency additions, major refactors) requires explicit confirmation.
