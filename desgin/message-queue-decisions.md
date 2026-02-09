## Decision
Use Kafka with at-least-once delivery semantics and explicit consumer offset management.

## Context
- Events must not be lost.
- Consumers may crash or restart.
- Backpressure must be handled safely.

## Options Considered
1. At-most-once delivery
2. At-least-once delivery
3. Exactly-once delivery

## Choice & Rationale
- At-least-once ensures durability.
- Duplicate events are handled via idempotency.
- Operational complexity remains manageable.

## Trade-offs Accepted
- Duplicate processing.
- Slightly more complex consumers.

## Revisit Conditions
- Financial transactions.
- Strong regulatory constraints.
