## Decision
Persist events using an append-only PostgreSQL table with database-enforced idempotency via unique event identifiers.

## Context
- Consumers may reprocess events.
- Crashes can occur between processing steps.
- Failure would look like duplicated records or inconsistent state.

## Options Considered
1. Blind inserts without constraints
2. Application-level deduplication only
3. Database-enforced uniqueness with transactional writes

## Choice & Rationale
- Database constraints provide a final correctness boundary.
- Append-only model simplifies reasoning and auditing.
- Transactions ensure atomic persistence.

## Trade-offs Accepted
- Storage growth over time.
- Less flexible schema evolution.

## Re
