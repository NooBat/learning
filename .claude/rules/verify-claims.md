# Verify Claims Before Stating (Docs First)

## Rule

**Before making any factual claim about Claude Code configuration, features, conventions, or behavior, verify against the official docs.** Do not assert from training-data memory when the claim is verifiable.

## When this applies

- Claims about what Claude Code auto-loads (`CLAUDE.md`, `.claude/rules/`, `.claude/skills/`, `.claude/agents/`, MCP config, etc.)
- Claims about plugin conventions, rules format, skills format, agent format, hook format
- Claims about `settings.json` / `settings.local.json` schemas and keys
- Claims about command behaviors, slash commands (`/init`, `/memory`, `/mcp`, etc.)
- Claims about file-naming conventions and resolution order
- Claims about what frontmatter fields exist and how they work
- Claims about environment variables and CLI flags

If the claim is about Claude Code itself or its config, assume verification is required.

## How to apply

1. **Fetch the docs** from `https://code.claude.com/docs/`. Start at the index `https://code.claude.com/docs/llms.txt` to find the right page.
2. **Quote the relevant section verbatim** when stating the answer. No paraphrase-as-fact.
3. **If docs are unavailable**, say so explicitly: *"I'm not sure — docs are unavailable right now. Want me to wait or proceed with my best guess flagged as unverified?"* Do not fill the gap with training-data memory presented as fact.
4. **Correct immediately and visibly** when a prior claim turns out to be wrong. Don't silently revise.

## Anti-patterns

- ❌ *"Claude Code has no native X"* — confident absence claim without fetching docs.
- ❌ *"The X feature doesn't exist"* — absence claims are highest-risk; always verify.
- ❌ Paraphrasing docs from training-data memory, then presenting the paraphrase as fact.
- ❌ Silently changing the answer after being corrected.

## Why this rule exists

A confident wrong answer is worse than admitting uncertainty. It wastes Daniel's time on the wrong path, erodes trust, and propagates misinformation into the repo's durable docs. "I don't know, let me check" is always the correct answer when the answer is verifiable and Claude isn't sure.

**Precedent:** Daniel caught a confident false claim that `.claude/rules/` wasn't a native Claude Code convention. The official docs clearly document it, including auto-loading and `paths` frontmatter. This rule exists to prevent that failure mode from recurring.
