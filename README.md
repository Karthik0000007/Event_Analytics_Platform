# Event Analytics Platform

Production-grade event-driven ingestion pipeline with at-least-once delivery, idempotent persistence, bounded retries, and DLQ isolation.

---

## The Problem

Marketplace platforms generate high-volume, bursty event traffic — clicks, purchases, impressions — that downstream systems depend on for business and operational decisions. Losing events silently is unacceptable, but async pipelines fail in subtle ways: brokers go down mid-publish, consumers crash mid-write, poison messages block entire partitions, and transient database errors cascade into permanent data loss if not handled with discipline.

Most event systems optimize for throughput. This one optimizes for **correctness under failure**. Every design decision — manual offset commits, idempotent writes, error classification, bounded retries, dead-letter isolation — exists to ensure that no event is silently lost, no poison pill blocks the pipeline, and every failure is visible, classified, and recoverable.

---

## Architecture

```
                    ┌──────────────────┐
                    │   HTTP Clients   │
                    └────────┬─────────┘
                             │ POST /v1/events
                             ▼
                    ┌──────────────────┐
                    │  Ingestion API   │  Validates → 202 Accepted
                    │     :8080        │  Async publish to Kafka
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │      Kafka       │  Topic: events
                    │  (durable buffer)│  At-least-once delivery
                    └────────┬─────────┘
                             │ manual fetch + commit
                             ▼
                    ┌──────────────────┐       ┌─────────────┐
                    │  Event Consumer  │──────▶│  PostgreSQL  │
                    │  (consumer group)│       │  (idempotent │
                    └────────┬─────────┘       │   storage)   │
                             │                 └──────────────┘
                             │ poison pills /
                             │ retries exhausted
                             ▼
                    ┌──────────────────┐
                    │    DLQ Topic     │  Topic: events.dlq
                    │  (full envelope) │  Preserves original message
                    └──────────────────┘
```

> Full component breakdown, sequence diagrams, and dependency graphs: **[ARCHITECTURE.md](ARCHITECTURE.md)**

---

## Correctness Guarantees

| Guarantee | Mechanism |
|---|---|
| **At-least-once delivery** | Kafka offsets committed only after successful DB write or DLQ routing. Never before. |
| **No silent data loss** | Uncommitted messages are redelivered on consumer restart. |
| **Idempotent persistence** | `INSERT ... ON CONFLICT (event_id) DO NOTHING` — duplicate replays are safe. |
| **Bounded retry** | Transient failures retried up to `MAX_RETRIES` (default 5) with exponential backoff + jitter. |
| **Permanent failure isolation** | Constraint violations and malformed data routed immediately to DLQ — no retries wasted. |
| **Poison pill handling** | Invalid JSON and missing required fields detected before any processing — routed to DLQ, offset committed, pipeline unblocked. |
| **DLQ forensics** | Every DLQ message wraps the original key/value/offset/partition with error classification, retry count, and failure timestamp. |
| **Crash-safe shutdown** | SIGINT/SIGTERM cancels context. In-flight retries abort without committing — safe for replay on restart. |

---

## Failure Handling Philosophy

Errors are not retried blindly. Every failure is **classified**:

- **Transient** (connection refused, timeout, too many clients) → retry with exponential backoff + jitter, up to a hard ceiling.
- **Permanent** (constraint violation, invalid input syntax) → no retry, immediate DLQ routing.
- **Unknown** → defaults to transient, because retrying a mysterious error is safer than discarding data.

Poison pills (unparseable messages, missing `event_id`) are detected **before** any processing begins and routed to the DLQ immediately. The consumer never blocks on a bad message.

If the DLQ write itself fails, the offset is **not committed** — the original message will be redelivered, and the DLQ write retried on the next cycle.

---

## How To Run

```bash
# 1. Start Kafka + PostgreSQL
docker-compose up -d

# 2. Apply schema
cat migrations/000001_create_events_table.up.sql | \
  docker-compose exec -T postgres psql -U events_user -d events_db

# 3. Run the API
go run ./cmd/ingestion_api

# 4. Run the consumer (separate terminal)
go run ./cmd/event-consumer

# 5. Send an event
curl -X POST http://localhost:8080/v1/events \
  -H "Content-Type: application/json" \
  -d '{"event_id":"550e8400-e29b-41d4-a716-446655440000","event_type":"click","payload":{"page":"/home"}}'

# 6. Run tests (27 passing — error classification, retry, DLQ, failure injection)
go test -v ./internal/messaging/...
```

---

## What I Would Improve in Production

- **Synchronous Kafka publish** in the API handler — the current async goroutine can lose events if the process crashes between `202 Accepted` and the Kafka write. A local write-ahead log or sync publish would close this gap.
- **Authentication + rate limiting** — the API is currently open. The stubbed `apikey.go` and `ratelimit.go` need implementation.
- **Postgres error-code classification** — replace `strings.Contains` matching with `pq.Error.Code` inspection for reliable, driver-version-safe classification.
- **Prometheus metrics** — counters for events ingested, published, persisted, retried, and DLQ'd. Histograms for publish and insert latency.
- **Distributed tracing** — OpenTelemetry spans from API → Kafka → Consumer → DB for end-to-end event tracing.
- **Multi-broker Kafka** with replication factor ≥ 3 — the current single-broker dev setup has no fault tolerance.
- **Consumer health endpoint** — the consumer binary has no HTTP probe for Kubernetes liveness/readiness checks.
- **Structured JSON logging** — replace `map[string]any` output with `slog` or `zerolog` for machine-parseable log aggregation.

---

## Project Structure

```
cmd/ingestion_api/       → HTTP server binary
cmd/event-consumer/      → Kafka consumer binary
internal/messaging/      → Producer, Consumer, DLQ, retry, error classification
internal/storage/        → PostgreSQL client (idempotent insert)
internal/api/            → Router, handlers, middleware
internal/config/         → Environment-based configuration
migrations/              → SQL schema (versioned)
deploy/                  → Docker + Kubernetes manifests (WIP)
```

> **Deep dive →** [ARCHITECTURE.md](ARCHITECTURE.md) — full domain model, execution flow diagrams, security analysis, failure maps, and onboarding guide.
