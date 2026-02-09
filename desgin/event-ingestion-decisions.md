## Decision
Expose a REST-based event ingestion API that validates, enriches, and asynchronously publishes immutable events to a durable message queue.

## Context
- Event ingestion is the highest-risk entry point.
- Producers may misbehave or retry aggressively.
- Traffic is bursty and unpredictable.
- Failure would look like dropped events or system overload.

## Options Considered
1. Direct database writes from producers
2. REST API with async publishing
3. Streaming ingestion using gRPC

## Choice & Rationale
- REST allows clear validation, throttling, and authentication.
- Async publishing decouples ingestion from downstream processing.
- Kafka acknowledgments define acceptance of responsibility.

## Trade-offs Accepted
- Slight ingestion latency.
- No streaming guarantees.

## Revisit Conditions
- Sustained extremely high ingestion throughput.
- Requirement for bidirectio
