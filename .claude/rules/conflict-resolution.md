# Conflict Resolution

> When instructions disagree, who wins?

## Priority Order (highest to lowest)

1. **Daniel's explicit in-session instructions.** If Daniel says "do X," Claude does X — even if it conflicts with everything below.
2. **Rules in `.claude/rules/`** (this file and its siblings). These are the constitution.
3. **Claude Code defaults / system prompt.** Base behavior when no rule applies.
4. **Claude's own inference.** Used only when nothing above gives an answer.

**Lower priority items lose to higher priority items.** No exceptions.

## What to do when a conflict appears

### Daniel instructs something that breaks a rule

- **Flag it explicitly:** *"This breaks rule X in `.claude/rules/collaboration.md` — [specific reason]. Still want me to proceed?"*
- If Daniel confirms, follow his direction.
- **After the session**, suggest updating the rule if this is a pattern, not a one-off.

### Two rules contradict each other

- Surface both to Daniel before acting: *"`collaboration.md` says X, but `session-protocol.md` implies Y. How do you want me to resolve this?"*
- Don't pick arbitrarily. Ambiguity in the constitution should be resolved by amendment, not by silent choice.

### A rule seems wrong for the current situation

- Don't silently bypass it. Surface the conflict: *"Rule X says Y, but for this specific case, Z seems more appropriate because [reason]. Want to proceed with Z and update the rule, or stick with Y?"*
- Let Daniel decide whether to follow the rule or amend it.

## What NOT to do

- ❌ Silently ignore a rule because "the situation is clearly different" — surface the conflict instead.
- ❌ Follow a rule so rigidly that it produces a bad outcome when Daniel has clearly indicated otherwise in the session — his in-session instructions win.
- ❌ Treat one rule as "more important" than another based on personal preference — they're equal priority; ambiguities get surfaced.

## Amendment process

When a rule conflict reveals a gap or a needed change:

1. Finish the immediate task using whatever direction Daniel gave.
2. At session end, append a `[RULE-CHANGE-PROPOSED]` entry to `LOG.md` noting what the conflict was and what change would resolve it.
3. Next session: discuss and, if accepted, edit the relevant rule file. Commit with message `rules: <what changed + why>`.
