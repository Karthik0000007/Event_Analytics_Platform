# Event-Driven Marketplace Analytics Platform

Production-grade backend system for ingesting and analyzing marketplace events using an event-driven architecture.

## Architecture (in progress)
- FastAPI API Gateway
- PostgreSQL for persistent storage
- RabbitMQ for async event processing
- Docker Compose for local orchestration

## How to Run

```bash
cd infra
docker-compose up --build
