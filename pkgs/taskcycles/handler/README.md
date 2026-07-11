# pkgs/taskcycles/handler

HTTP surface for execution cycles, indexed commits, and the global cycle-failures list.

## Routes (`Register`)

| Method | Path | Handler |
| --- | --- | --- |
| GET | `/tasks/cycle-failures` | `cycleFailures` |
| POST | `/tasks/{id}/cycles` | `postTaskCycle` |
| GET | `/tasks/{id}/cycles` | `getTaskCycles` |
| GET | `/tasks/{id}/cycles/{cycleId}` | `getTaskCycle` |
| PATCH | `/tasks/{id}/cycles/{cycleId}` | `patchTaskCycle` |
| POST | `/tasks/{id}/cycles/{cycleId}/phases` | `postTaskCyclePhase` |
| PATCH | `/tasks/{id}/cycles/{cycleId}/phases/{phaseSeq}` | `patchTaskCyclePhase` |
| GET | `/tasks/{id}/cycles/{cycleId}/stream` | `getTaskCycleStream` |
| GET | `/tasks/{id}/cycles/{cycleId}/verdicts` | `getTaskCycleVerdicts` |
| GET | `/tasks/{id}/commits` | `getTaskCommits` |

Wired from `pkgs/tasks/handler/handler_routes.go` via `taskcycleshandler.Register`.

## Files

| File | Role |
| --- | --- |
| `handler.go` | `Register`, `Deps`, enriched SSE notify bridge |
| `http_helpers.go` | JSON I/O, store errors, path parsing, ETag responses |
| `handler_cycles.go` | Cycle + phase mutation handlers |
| `handler_cycles_query.go` | List/stream query parsing, pagination helpers |
| `handler_cycles_json.go` | Request/response DTOs |
| `handler_cycles_response.go` | Domain → wire projection |
| `handler_commits.go` | Task commit index list |
| `handler_cycle_failures.go` | Global cycle failures page |

Contract tests against the full mux remain in `pkgs/tasks/handler/handler_http_cycles*_test.go`.
