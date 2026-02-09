## Decision
Define strict boundaries between event producers, ingestion API, message broker, consumers, and storage.

## Context
- Async systems fail at boundaries.
- Ambiguous ownership causes data loss.
- This system must be debuggable by on-call engineers.

## Options Considered
1. Tight coupling between services
2. Fully decoupled async pipeline
3. Shared database writes

## Choice & Rationale
- Fully decoupled async pipeline.
- Kafka as durability boundary.
- Consumers are stateless and restartable.

## Trade-offs Accepted
- Eventual consistency.
- More components to operate.

## Revisit Conditions
- Strong real-time requirements.
- Need for exactly-once semantics.
