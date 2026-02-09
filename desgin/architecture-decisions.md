## Decision
Use a REST-based ingestion API backed by Kafka and PostgreSQL with at-least-once delivery semantics.

## Context
- Internal event ingestion workload.
- Bursty traffic.
- Solo engineer.
- Failure would look like silent data loss.

## Options Considered
1. REST + Kafka + Postgres
2. gRPC + Kafka + NoSQL
3. Direct DB writes

## Choice & Rationale
- REST is debuggable and flexible.
- Kafka provides durable buffering.
- Postgres ensures strong consistency.

## Trade-offs Accepted
- Duplicate processing possible.
- Lower theoretical throughput.

## Revisit Conditions
- Extreme scale (> millions events/sec).
- Strict latency SLAs.
