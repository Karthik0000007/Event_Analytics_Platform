## Decision
Expose a REST-based event ingestion API that accepts immutable, versioned events with server-side validation and rate limiting.

## Context
- External services may misbehave.
- Traffic is bursty.
- Ingestion is the highest-risk entry point.
- Failure would look like dropped events or system overload.

## Options Considered
1. Direct database writes
2. REST API with async publishing
3. gRPC streaming ingestion

## Choice & Rationale
- REST API is debuggable and operationally simple.
- Async publish to Kafka decouples ingestion from processing.
- Validation and throttling protect downstream systems.

## Trade-offs Accepted
- Slight ingestion latency.
- No streaming guarantees.

## Revisit Conditions
- Very high sustained throughput.
- Strict latency requirements.
