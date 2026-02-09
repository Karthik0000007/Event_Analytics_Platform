## Decision
Design the system assuming partial failures are normal and frequent.

## Context
- Distributed systems fail by default.
- Silent failures are worse than loud failures.
- Engineers must debug incidents quickly.

## Options Considered
1. Optimistic design assuming healthy systems
2. Defensive design with retries and idempotency

## Choice & Rationale
- Defensive design.
- At-least-once delivery with idempotent consumers.
- Explicit dead-letter handling.

## Trade-offs Accepted
- More code paths.
- Slightly higher latency.

## Revisit Conditions
- Strong transactional guarantees required.
