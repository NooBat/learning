# Session Log

Append-only. Newest entries at the bottom. Use the session-end skill (`/session-end`) to append structured entries; manual entries are OK but keep the format consistent so the tail is scannable.

Entry types:
- **Session summary** (default, every session)
- `[LEVEL-TRANSITION]` — when moving from one level to the next
- `[RETRO]` — every 3-4 levels or end of tier
- `[ADR-PROPOSED]` / `[ADR-ACCEPTED]` — decision record created or approved
- `[RULE-CHANGE-PROPOSED]` — when a `.claude/rules/` file should be amended
- `[PIVOT]` — when the roadmap changes direction

---

## [2026-04-17] Repo initialized

**Did:**
- Scaffolded learning repo from approved plan (`~/.claude/plans/i-am-currently-a-shiny-crab.md`).
- Created `CLAUDE.md` (auto-loaded entry point) + `ABOUT-ME.md` + `ROADMAP.md` + `STATUS.md` + `LEVELS.md` + this file.
- Created `.claude/rules/`: `collaboration.md`, `session-protocol.md`, `conflict-resolution.md`, `verify-claims.md`.
- Created `.claude/skills/`: `bootstrap`, `start-level`, `write-adr`, `session-end`, `retro`.
- Created `.claude/settings.json` with `"outputStyle": "Learning"` — per-repo learning mode.
- Wrote `levels/L01-mvs.md` (first level brief).
- Wrote `infrastructure/setup-guides/01-go-toolchain.md` and `02-postgres-local.md`.
- Wrote `.gitignore`, ran `git init`.

**Decided:**
- Keystone project: Developer Automation Platform (working name `flux`) — covers scheduling, notifications, feature flags as one system.
- Backend: Go (not TypeScript).
- Tracking: level-based, not calendar-based.
- Source of truth: this repo pushed to private GitHub. Claude-side memory treated as disposable cache.
- Rules format: `.claude/rules/*.md` — native Claude Code convention. Always-apply rules have no frontmatter per docs.
- Collaboration model: Claude writes guides + scaffolds, Daniel runs commands + writes business logic.

**Blocked:**
- Private GitHub repo not yet created. Daniel needs to create `learning` (or another name) as a private repo on github.com, then push.

**Next:**
- Daniel creates the remote repo and pushes.
- Install Go per `infrastructure/setup-guides/01-go-toolchain.md`.
- Install Postgres per `infrastructure/setup-guides/02-postgres-local.md`.
- Begin L01: init Go module in `project/`, write first handler.
