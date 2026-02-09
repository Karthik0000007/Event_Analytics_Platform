## Decision
Use Kafka with at-least-once delivery semantics and explicit producer acknowledgments and consumer offset management.

## Context
- Events must not be lost under failures.
- Consumers may crash or restart.
- Backpressure must be handled safely.
- Failure would look like silent loss or stuck partitions.

## Options Considered
1. At-most-once delivery
2. At-least-once delivery
3. Exactly-once delivery

## Choice & Rationale
- At-least-once ensures durability.
- Duplicate events are acceptable when combined with idempotent consumers.
- Avoids the operational and conceptual complexity of exactly-once semantics.

## Trade-offs Accepted
- Duplicate event processing.
- Additional logic for idempotency.

## Revisit Conditions
- Financial transactions.
- Regulatory or compliance-driven guarantees.
