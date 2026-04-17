# Session Protocol

> What every Claude session does at start and end. Enforced regardless of how the session was invoked (Claude Code, Claude.ai, API, other LLM).

## Session Start

Before taking any action — even a clarifying question — read these files in order:

1. `CLAUDE.md` (auto-loaded by Claude Code; if another environment, read it manually)
2. `.claude/rules/collaboration.md`
3. `.claude/rules/session-protocol.md` (this file)
4. `.claude/rules/conflict-resolution.md`
5. `ABOUT-ME.md`
6. `STATUS.md`
7. Tail of `LOG.md` (~100 lines is plenty; use `tail -n 100` equivalent)
8. `LEVELS.md` if the current work touches level boundaries

After this, Claude should be able to answer:
- What level is Daniel on?
- What did the last session accomplish?
- What was left blocked?
- What was the next-session target set by the previous session?

If any of those are unclear, ask Daniel before proceeding.

### Cost-benefit on re-reading

These files are small (~1-2k lines total). Re-read them every session — do not rely on memory from a previous session, another account, or another model.

## During Session

- **Commit often.** Each meaningful chunk of work (a passing test, a completed level criterion, an ADR drafted) is a commit. Small commits make git the changelog.
- **Update `STATUS.md` when state changes** — level transitions, new blockers, direction changes. Don't batch this for the end.
- **Write ADRs in the moment** when decisions happen. `adrs/NNNN-<slug>.md`. Never postpone — context evaporates.
- **Respect "one level ahead."** Don't pre-write level briefs further than current level + 1.

## Session End

Before yielding control back to Daniel, Claude performs this checklist:

1. **Update `STATUS.md`** — current level (if changed), any new blockers, explicit next-session target.
2. **Append to `LOG.md`** with dated entry:
   ```
   ## [YYYY-MM-DD] Session summary
   - Did: <list of concrete things>
   - Decided: <any decisions made, with pointer to ADR if applicable>
   - Blocked: <anything waiting>
   - Next: <next session's target>
   ```
3. **Remind Daniel to commit and push** if changes weren't already pushed:
   ```
   git status        # verify state
   git add -A && git commit -m "<descriptive message>"
   git push
   ```
   Per the "push or it didn't happen" rule, the session isn't durable until this completes.
4. **Flag rule violations**, if any occurred, so Daniel can decide whether to update the rules or enforce them more strictly.

## Special Session Types

### Level transition (starting a new level)

- Update `LEVELS.md`: mark previous level `[x]`, current level `[~]`.
- Update `STATUS.md`: new current level + opening context.
- Write the next level's brief if it doesn't exist (`levels/L<N+1>-*.md`).
- Append `LOG.md` with `[LEVEL-TRANSITION]` tag noting what carried over.

### Level retro (every 3-4 levels)

- Append to `LOG.md` with `[RETRO]` tag.
- Questions: What stuck? What didn't? Is this project still interesting? Any pivots?
- If a pivot is warranted, write an ADR before executing.

### Tier transition

- Bigger retro: was the tier's effort proportional to value?
- Review `ROADMAP.md` — adjust next tier if needed.
- `LOG.md` with `[TIER-TRANSITION]` tag.

## Anti-patterns to avoid

- **Skipping the bootstrap read.** "I remember from last time" is a memory hazard across accounts/models.
- **Batching updates until end of session** — if the session crashes, the state is lost.
- **Long sessions without commits** — commits are cheap insurance.
- **Writing level briefs for L+2, L+3, L+4...** — they rot. Write one level ahead, max.
