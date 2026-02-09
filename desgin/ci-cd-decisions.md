## Decision
Use a simple CI/CD pipeline with automated build, test, and deploy stages gated by quality checks.

## Context
- Frequent iteration increases risk of regressions.
- Solo engineer must rely on automation.
- Failure would look like broken deployments or unsafe changes.

## Options Considered
1. Manual deployment
2. CI-only without automated deploy
3. CI with gated automated deployment

## Choice & Rationale
- Automated builds and tests catch issues early.
- Gated deploys prevent unsafe changes reaching production.
- Keeps deployment repeatable and boring.

## Trade-offs Accepted
- Slower iteration speed.
- Initial setup overhead.

## Revisit Conditions
- Larger team with release engineering support.
- Multi-environment deployment requirements.
