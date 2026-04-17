# About Daniel

> Portable user profile. The source of truth for who Daniel is, what he's aiming for, and how he prefers to work. Any Claude session should read this before anything else.

## Identity

- **Name:** Daniel Nguyen
- **Email:** khoi27012003@gmail.com
- **Current role:** Junior Frontend Engineer
- **Target role:** Solutions Architect
- **Timeframe:** 6-7 years

## Current Skills

### Strong

- **React** + **TypeScript** — day-job confident. No need for basic FE training.
- Working knowledge of modern FE tooling (bundlers, linting, CSS frameworks).

### Basic / some exposure

- Backend CRUD — has written some, but not deeply.

### Zero or near-zero (going in blind)

- **Go** — this is the backend language we're learning from scratch.
- Postgres beyond the basics (indexes, replication, etc.).
- Docker / containers operationally.
- Cloud platforms (AWS specifically).
- Distributed systems concepts.
- Architecture patterns (microservices, event-driven, DDD, CQRS, etc.).
- Observability, SRE practices.
- Security posture beyond "don't commit secrets."

## Constraints

- **Time:** 5-7 focused hours per week. Sustainable, not aggressive.
- **Budget:** personal (not expensable) — prefer free/cheap infra until a level genuinely requires paid services (L04 deploy = Fly.io free tier; L09 AWS = cost-aware).
- **Learning happens alongside a full-time job** — weekends are the primary work window.

## Learning Preferences

Daniel learns best through **all four modes combined**:

1. **Building projects end-to-end.** Primary mode. One keystone project that evolves through levels.
2. **Deep reading.** DDIA, Release It!, Richards & Ford — treated as canonical references, read when the project demands them.
3. **Structured courses / certs.** AWS SAA → SAP at natural tier boundaries.
4. **Studying real codebases & systems.** Postmortems, open-source projects, architecture case studies.

### Pedagogical preferences

- **Level-based progression, not calendar-based.** Daniel advances on exit criteria, not by month. No guilt when a busy week eats the learning slot.
- **Teach-back over lecture.** After Claude introduces a concept, Daniel explains it back in his own words. If he can't, he doesn't understand it yet.
- **He does the setup himself.** Install Go, start Postgres, run migrations, provision cloud resources — Daniel runs the commands. Claude writes guides, Daniel follows them.
- **ADR-gated decisions.** Any real architectural decision gets a written ADR before implementation.
- **Push-or-it-didn't-happen.** Git push after every meaningful session. The repo is the durable memory.

## Goals

### Primary (6-7 year arc)

- Become a **Solutions Architect** — technical breadth across frontend, backend, data, cloud, and security, with depth in 1-2 areas (chosen around Year 3-4).
- Build a portfolio of real artifacts: a long-running keystone service, ADRs, architecture docs, a blog.
- Pass AWS SAA (~Tier 2-3 boundary) and AWS SAP (~L14).

### Secondary

- Write a technical blog from L06 onward (~1 post per level cleared).
- At each tier boundary, have one "running example" in the keystone project that Claude can point at for architectural discussion.

## Not Goals

- **No mobile / iOS / Android** specialization.
- **No ML model training** — will learn to architect AI systems (retrieval, inference, evals) but not train models.
- **No becoming a frontend specialist** — FE skills maintained for pragmatic use, not deepened.
- **No second backend language** unless a specific level genuinely demands it (e.g., Rust for a perf-critical component much later).

## Collaboration Notes for Claude

These are Daniel-specific preferences, separate from the universal rules in `.claude/rules/`:

- **He pushes back hard on confidently wrong claims.** Claude should fetch docs before asserting facts about tools, config, or conventions (see `.claude/rules/verify-claims.md`).
- **He prefers existing conventions over invented ones.** Align with documented Claude Code patterns; don't create novel folder structures or metadata formats without explicit agreement.
- **He wants to learn the setup, not have it handed to him.** Claude writes guides; Daniel runs commands.
- **He's OK with bluntness.** No performative hedging or over-padded explanations. Direct is better.
- **He redirects early and often.** If a proposal doesn't fit, he'll say so. Corrections should be treated as course corrections, not failures.

## Contact Points & External Systems

- **Git remote:** private GitHub repo (created at L04 setup, or earlier).
- **Email for technical accounts / cloud services:** khoi27012003@gmail.com.

## Last Updated

Daniel should update this file at every tier boundary and any time a major goal/constraint changes. Git history is the changelog.
