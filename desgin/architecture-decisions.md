## Decision
Use a REST-based ingestion API backed by Kafka for buffering and PostgreSQL for durable storage with at-least-once delivery semantics.

## Context
- The system ingests bursty, write-heavy event traffic.
- Correctness and debuggability are more important than raw throughput.
- Failure would look like silent event loss or unbounded backpressure.

## Options Considered
1. REST + Kafka + PostgreSQL
2. gRPC + Kafka + NoSQL store
3. Direct synchronous database writes

## Choice & Rationale
- REST is simple, debuggable, and operationally familiar.
- Kafka provides durable buffering and backpressure handling.
- PostgreSQL provides strong c
