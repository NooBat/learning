# Levels

Status: `[ ]` open · `[~]` in-progress · `[x]` done

## Tier 1 — Backend Engineer

- [x] **L01 — mvs** — Minimum viable service: Go HTTP + Postgres CRUD for workflows (shipped 2026-04-20; PUT/DELETE deferred to L02 as mechanical extension)
- [~] **L02 — auth-tenancy** — Identity & multi-tenancy, first integration tests
- [ ] **L03 — production-local** — Docker Compose, migrations, structured logs, graceful shutdown
- [ ] **L04 — first-deployment** — Ship to Fly.io / Railway with CI

## Tier 2 — Mid-Level

- [ ] **L05 — async-work** — Background jobs, idempotency, retries, DLQ
- [ ] **L06 — caching-perf** — Redis cache, k6 load testing, pprof profiling
- [ ] **L07 — resilience** — Timeouts, circuit breakers, rate limiting, chaos testing
- [ ] **L08 — observability** — OpenTelemetry traces + metrics, Grafana, written SLOs

## Tier 3 — Senior

- [ ] **L09 — cloud-aws** — AWS migration via Terraform (ECS/EKS, RDS, ElastiCache, ALB)
- [ ] **L10 — event-driven** — Split execution engine via NATS/Kafka, outbox pattern
- [ ] **L11 — polyglot-persistence** — Second data store (search / time-series / graph)
- [ ] **L12 — security-posture** — Threat modeling, OWASP, secrets mgmt, IAM least-privilege

## Tier 4 — Staff / Architect

- [ ] **L13 — scale-multi-region** — Active-passive failover, geo data, latency budgets
- [ ] **L14 — ddd-refactor** — Bounded contexts, strategic DDD, ADRs per decision
- [ ] **L15 — architecture-rfc** — Formal RFC + public critique
- [ ] **L16 — capstone** — Design a different system end-to-end (portfolio artifact)

## Tier exits (natural certs)

- After **Tier 2**: AWS Solutions Architect Associate
- After **Tier 4** (at L14): AWS Solutions Architect Professional

## How to update this file

- `/start-level <NN> <slug>` flips `[ ]` → `[~]` and marks previous `[x]`.
- Manual edits OK — git history is the audit trail.
- Only one level should be `[~]` at a time (focus rule). If you need to pause L<N> and work on L<N-1>, flip L<N> back to `[ ]` and re-flip L<N-1> to `[~]`.
