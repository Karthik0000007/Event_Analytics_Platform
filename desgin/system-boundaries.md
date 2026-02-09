## Decision
Define strict boundaries between event producers, ingestion API, message broker, consumer services, and storage systems.

## Context
- Distributed systems fail most often at unclear ownership boundaries.
- Async pipelines require explicit responsibility separation.
- Failure would look like unclear debugging paths or silent data loss.

## Options Considered
1. Tight coupling between producers and database
2. Shared database writes across services
3. Fully decoupled async pipeline using a message broker

## Choice & Rationale
- Selected a fully decoupled async pipeline.
- Ingestion API owns validation and acceptance.
- Message broker acts as the durability boundary.
- Consumer services own persistence and correctness.
- Storage systems are isolated from producers.

## Trade-offs Accepted
- Eventual consistency.
- Additional oper
