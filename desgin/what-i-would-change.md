## Decision
Accept known limitations for a solo-built, time-boxed system.

## Context
- System is built within 14 days by one engineer.
- Failure would look like overengineering or unfinished core functionality.

## Options Considered
1. Build for hypothetical massive scale
2. Build for realistic internal platform scope

## Choice & Rationale
- Chose realistic internal platform scope.
- Focused on correctness, clarity, and debuggability.

## Trade-offs Accepted
- No schema registry.
- No stream processing layer.
- Manual operational workflows.

## Revisit Conditions
- Team size increases.
- Production usage grows significantly.
