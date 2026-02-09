## Decision
Build a production-grade event ingestion and processing backend focused on correctness and operational safety, not analytics features.

## Context
- Marketplace platforms rely on user events for business decisions.
- Data loss or duplication silently damages trust.
- Solo engineer, 14-day timeline.
- Failure would look like missing events, duplicated rows, or un-debuggable incidents.

## Options Considered
1. End-to-end analytics platform (ingestion + dashboards)
2. Real-time stream processing system
3. Backend ingestion + processing only

## Choice & Rationale
- Chose ingestion + processing only.
- Keeps scope realistic and production-focused.
- Mirrors internal platform teams rather than product teams.

## Trade-offs Accepted
- No visible output dashboards.
- Less visually impressive demo.

## Revisit Conditions
- If building for a real company team.
- If downstream analytics consumers are added.
