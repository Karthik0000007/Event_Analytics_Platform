## Decision
Build a production-grade event ingestion and processing backend focused on correctness, durability, and operational safety rather than analytics features.

## Context
- Marketplace platforms depend on high-volume user events for business and operational decisions.
- Silent data loss or duplication damages trust and is difficult to detect.
- The project is built by a solo engineer with a 14-day timeline.
- Failure would look like lost events, duplicated records, or systems that are difficult to debug during incidents.

## Options Considered
1. Full analytics platform including dashboards and queries
2. Real-time stream processing system with complex semantics
3. Backend ingestion and processing platform only

## Choice & Rationale
- Chose backend ingestion and processing only.
- Keeps scope realistic and production-focused.
- Mirrors internal platform or infrastructure teams rather than product-facing systems.
- Allows focus on reliability, async correctness, and observability.

## Trade-offs Accepted
- No user-facing analytics or dashboards.
- Less visually impressive demo.

## Revisit Conditions
- If building for a real organization with downstream analytics consumers.
- If product requirements demand near real-time user-facing analytics.
