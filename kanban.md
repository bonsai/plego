# plego Kanban

## Backlog

| # | Title | Notes |
|---|-------|-------|
| | Filter plugins | HTML2Text, Deduped, etc. |
| | Publish plugins | Gmail, GoogleDrive, etc. |
| | Subscription plugins | Filesystem, GoogleDrive, RSS, etc. |
| | CI/CD workflow | Go test + build + release |
| | Test suite | Pipeline tests for each plugin type |
| | Dockerfile / multi-stage build | Static binary image |
| | plego.example.yaml | Document all plugin options |

## In Progress

| # | Title | Branch |
|---|-------|--------|
| PR #1 | Subscription/Filter/Publish architecture, Google Drive plugins, Gmail encoding fix | `feat/subscription-filter-publish-architecture` |

## Done

## Under Consideration

| Title | ADR | Notes |
|-------|-----|-------|
| Codebase → OpenAPI → Spec pipeline | ADR-001 | Go source → OpenAPI 3.1 → yaml/json/markdown |
| Web UI dashboard | | Real-time pipeline status |
| Prometheus metrics | | Feed processing latency, counts |
