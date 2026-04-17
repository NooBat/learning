# Roadmap: Junior FE → Solutions Architect

> The 6-7 year plan, in durable form. This replaces the session-scoped plan file. Amend freely as things change — git is the changelog. For the moment-to-moment status, see `STATUS.md`.

## The approach in one page

- **One keystone project**, built in **Go**, lives in `project/`. A Developer Automation Platform (Zapier/n8n/Temporal-style) — workflows = triggers + steps. Starts as CRUD at L01, grows into a distributed, event-driven, multi-region system by Tier 4.
- **Level-based progression**, not calendar-based. Advance on exit criteria.
- Reading, certifications, writing happen *in support of* the project, not in parallel to it.
- The **repo is the memory.** Every decision is in git.

## Level Map

### Tier 1 — Backend Engineer (L01-L04)

"I can build and ship a real Go service."

| Level | Goal | New concepts |
|-------|------|--------------|
| **L01** — Minimum viable service | Go + Postgres CRUD for workflow definitions; no execution yet | Go modules, HTTP handlers, `database/sql` or `pgx`, schema design |
| **L02** — Identity & multi-tenancy | Auth (session or JWT), tenant isolation, first integration tests | Auth flows, row-level security / tenant scoping, `testcontainers-go` |
| **L03** — Production-ready locally | Docker Compose, env config, migrations, graceful shutdown, structured logs | `goose`/`golang-migrate`, `log/slog`, health checks |
| **L04** — First real deployment | Ship to Fly.io or Railway with managed Postgres + CI | GitHub Actions, the "works on my machine" gap |

**Tier 1 exit:** a real URL on the internet. Postmortem in `LOG.md`.

### Tier 2 — Mid-Level (L05-L08)

"This service doesn't fall over under load."

| Level | Goal | New concepts |
|-------|------|--------------|
| **L05** — Async work | Background jobs executing workflow steps; retries, DLQ | `asynq` or `river`, idempotency, at-least-once semantics |
| **L06** — Caching & performance | Redis cache with explicit invalidation; load test with k6 | Cache invalidation, `pprof`, performance profiling |
| **L07** — Resilience | Timeouts, circuit breakers, rate limiting, chaos testing | Failure modes, degradation strategies |
| **L08** — Observability | OpenTelemetry traces + metrics, Grafana dashboard, written SLOs | Traces, metrics, logs-with-correlation |

**Tier 2 exit:** 1000 req/s in load test + debuggable via observability alone. Natural cert: **AWS Cloud Practitioner** or straight to **AWS Solutions Architect Associate**.

### Tier 3 — Senior (L09-L12)

"I think in distributed systems."

| Level | Goal | New concepts |
|-------|------|--------------|
| **L09** — Cloud migration (AWS) | Fly.io → AWS (ECS/EKS + RDS + ElastiCache + ALB) via Terraform | IaC, VPC / subnets / security groups |
| **L10** — Event-driven split | Extract execution engine into separate service via NATS/Kafka | Outbox pattern, exactly-once vs at-least-once |
| **L11** — Polyglot persistence | Add search (Meilisearch) or time-series / graph where it belongs | "Right tool" in practice; cost awareness |
| **L12** — Security posture | Threat-model service, OWASP fixes, secrets to AWS SM, IAM least-privilege | Threat modeling, secrets mgmt, IAM |

**Tier 3 exit:** can give a 30-min talk on the service's architecture with explicit trade-offs. Architecture doc at `project/docs/architecture.md`. *DDIA* reading through this tier.

### Tier 4 — Staff / Architect (L13-L16)

"I own the technical direction."

| Level | Goal | New concepts |
|-------|------|--------------|
| **L13** — Scale & multi-region | Active-passive failover, geo data concerns, latency budgets | Geography as a physical constraint |
| **L14** — DDD refactor | Identify bounded contexts, refactor, ADRs for every significant call | Strategic design. **AWS SAP cert** here. |
| **L15** — Architecture artifacts | Formal RFC for a major change, invite public critique | Architects are writers first |
| **L16** — Capstone: different system | Design (on paper) a system in a different domain with ADRs + costs | Transferable architecture thinking |

**Tier 4 exit:** portfolio — running keystone service + ADRs + blog + fully-designed second system. What a Solutions Architect interview panel looks at.

## Parallel Tracks

These run *when the project demands it*, not on a schedule.

- **Reading.** Triggered by the level. See reading list below.
- **Writing.** Start a blog at ~L06. Target: 1 post per level cleared. Posts come from real problems.
- **Certifications.** AWS SAA around Tier 2→3 boundary; AWS SAP at L14.
- **Communication.** From L10: write ADRs at the day job on real decisions. Highest-leverage skill for senior→staff→architect.

## Reading List

Grouped by tier. Read each book **when the level demands it, not before** — theory without practice has a short half-life. ★ = the single most important per tier.

### Tier 1 — Go & Backend Foundations

- **★ core:** *Learning Go* (Jon Bodner, 2nd ed.)
- **core:** *100 Go Mistakes and How to Avoid Them* (Teiva Harsanyi) — after ~1000 LOC of Go
- **optional:** *The Go Programming Language* (Donovan & Kernighan)
- **optional:** *Grokking Simplicity* (Eric Normand)

### Tier 2 — Systems-Aware Engineering

- **★ core:** *Release It!* (Michael Nygard, 2nd ed.) — before L07
- **core:** *System Design Interview Vol 1* (Alex Xu) — one chapter/week
- **core:** *Database Internals* (Alex Petrov)
- **optional:** *System Design Interview Vol 2*
- **optional:** *The Art of Scalability* (Abbott & Fisher)

### Tier 3 — Distributed Systems

- **★ core:** *Designing Data-Intensive Applications* (Martin Kleppmann) — the book. Start at L09, 3-4 months. If you read nothing else, read this.
- **core:** *Building Microservices* (Sam Newman, 2nd ed.) — before L10
- **core:** *Microservices Patterns* (Chris Richardson)
- **optional:** *Designing Distributed Systems* (Brendan Burns)
- **optional:** *Distributed Systems* (van Steen & Tanenbaum, free PDF)

### Tier 4 — Architecture & Leadership

- **★ core:** *Fundamentals of Software Architecture* (Richards & Ford)
- **core:** *Software Architecture: The Hard Parts* (Richards & Ford)
- **core:** *Domain-Driven Design Distilled* (Vaughn Vernon)
- **core:** *The Software Architect Elevator* (Gregor Hohpe)
- **optional:** *Implementing Domain-Driven Design* (Vernon)
- **optional:** *Domain-Driven Design* (Evans, "Blue Book")
- **optional:** *Staff Engineer* (Will Larson)

### Cross-cutting (any time)

- **core:** *The Pragmatic Programmer* (Hunt & Thomas, 20th anniv.)
- **core:** *Accelerate* (Forsgren, Humble, Kim)
- **core:** *On Writing Well* (Zinsser)
- **optional:** *Team Topologies* (Skelton & Pais) — from L14

### Security (dip into at L12)

- **core:** *The Tangled Web* (Michal Zalewski)
- **optional:** *Threat Modeling: Designing for Security* (Adam Shostack)

### Cloud / AWS

- AWS **Well-Architected Framework** docs (free) — read at L09, re-read at L13
- *AWS SAA Study Guide (SAA-C03)* — Tier 2→3
- *AWS SAP Study Guide (SAP-C02)* — L14
- **optional:** *Cloud Native Patterns* (Cornelia Davis)

### Deliberately Missing

- TOGAF / Zachman — enterprise frameworks. Sampled selectively at Tier 4 only if target role demands.
- "Clean code" / "clean architecture" (Martin) — mixed reputations, better material above.
- Kubernetes-specific book — docs + *Designing Distributed Systems* suffice.
- React/TS books — Daniel is already strong.

## Revising This Plan

A 6-7 year plan that can't change is useless. Changes are expected.

| Scope | Who decides | How |
|-------|-------------|-----|
| Level brief (scope, exit criteria, reading) | Either, discussed | Edit the brief. `[LEVEL-CHANGE]` in `LOG.md` with reason. |
| Book swaps / additions | Daniel | Edit this file. `LOG.md` note. |
| Ground rules | Daniel + Claude | Edit `.claude/rules/*.md`. Commit `rules: <what + why>`. |
| Project pivot | Daniel | ADR `adrs/NNNN-project-pivot.md`. Update everything. |
| Tier/level ordering | Either, discussed | Edit `LEVELS.md` + briefs. `LOG.md` entry. |
| Career target | Daniel | Update `ABOUT-ME.md` + reshape later tiers. |

### Proactive revisit triggers

- **Every 4 levels cleared** — short retro in `LOG.md`. Still the right direction?
- **End of each tier** — bigger retro. Was the effort proportional to value?
- **Stuck / unmotivated for 2+ sessions** — stop and talk. Motivation loss is signal.

### Audit trail

Git log is the changelog. `git log ROADMAP.md` shows every change and the commit reason.

## Decisions Locked In

- **Keystone project:** Developer Automation Platform (working name `flux`; rename freely). Covers scheduling, notifications, feature flags, and general backend depth in one system.
- **Backend language:** Go.
- **Git remote:** GitHub, private.
- **Cloud target:** AWS from Tier 3. Fly.io for first deploy at L04 (lower friction).

## Critical Files (priority read order for any Claude session)

1. `CLAUDE.md` — auto-loaded, entry point.
2. `.claude/rules/*.md` — auto-loaded, the constitution.
3. `ABOUT-ME.md` — user profile.
4. `STATUS.md` — current focus.
5. `LEVELS.md` — level map.
6. Tail of `LOG.md` — recent history.
