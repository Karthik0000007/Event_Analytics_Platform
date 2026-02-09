## Decision
Deploy services on Kubernetes using Deployments for stateless services and StatefulSets for stateful dependencies.

## Context
- Services must restart safely.
- Horizontal scaling should be simple.
- Failure would look like crash loops or partial outages.

## Options Considered
1. VM-based deployment
2. Single-node container runtime
3. Kubernetes with explicit resource and probe configuration

## Choice & Rationale
- Kubernetes provides restart guarantees and scaling primitives.
- Deployments suit stateless ingestion and consumer services.
- StatefulSets suit Kafka and PostgreSQL.

## Trade-offs Accepted
- Operational complexity.
- Learning curve for configuration.

## Revisit Conditions
- Very small-scale deployment.
- Managed platform abstractions available.
