## Decision
Design the system assuming partial failures are normal and frequent.

## Context
- Distributed systems fail by default.
- Silent failures are worse than loud failures.
- Failure would look like unnoticed data loss.

## Options Considered
1. Optimistic design assuming healthy systems
2. Defensive design with retries and idempotency

## Choice & Rationale
- Defensive design protects data correctness.
- At-least-once delivery combined with idempotent consumers.
- Explicit DLQ handling prevents poison pills.

## Trade-offs Accepted
- Increased complexity.
- Duplicate processing.

## Revisit Conditions
- Strong transactional guarantees required.
- Regulatory compliance demands stronger semantics.
