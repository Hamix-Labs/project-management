# ADR-0030: Attempt Phase Correlation and Activity Filter

> **Note** - Product renamed T2A to Hamix; identifiers below reflect the name at decision time unless updated inline.

**Date:** 2026-06-19
**Status:** Accepted
**Deciders:** Engineering

## Context

Operators and agents debug attempts using task id, cycle id, and phase_seq across separate surfaces (phase ledger, Cursor stream rows, audit mirrors, taskapi logs). After in-cycle verify-only retry ([ADR-0028](ADR-0028-in-cycle-verify-only-retry.md)), multiple verify phases in one cycle make log grep ambiguous. The attempt-detail SPA shows Cursor and Audit lists with phase badges but no first-class filter or shareable deep link.

## Decision

1. **Store:** On every `StartPhase`, seed `details_json.run_correlation_id` with a new UUID. On `CompletePhase`, shallow-merge incoming details and **never drop** the seeded id.
2. **Harness:** Read id from the phase row after start; attach to phase-scoped slog, `runner.Request`, and live SSE progress.
3. **Web:** Phase-click filter on attempt detail with URL `?phase={phase_seq}`; optional copy of debug id from phase details.

### Stable handles

| Handle | Purpose |
| --- | --- |
| `?phase=N` | UI filter for Cursor + Audit on attempt detail |
| `run_correlation_id` | Single grep key for harness + runner logs for one phase run |

### Field name

`run_correlation_id` (snake_case) in phase `details_json`, `phase_started` audit payload, and optional SSE field.

## Consequences

### Positive

- One log query isolates a single execute or verify invocation.
- Agents and humans share documented URL + field names ([docs/api.md](../api.md), [docs/agent-map.md](../agent-map.md)).
- No new tables; additive JSON and SSE fields.

### Negative / trade-offs

- Legacy phases lack `run_correlation_id`; UI filter still works via `phase_seq`.
- `CompletePhase` merge adds a small store helper (callers need not change).

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| OTel trace spans | Heavier ops dependency; defer |
| Column on `task_cycle_stream_events` | `phase_seq` sufficient for UI; id lives on phase row |
| Third Activity tab | Filters on existing data are enough |

## Non-goals (v1)

- Prometheus label per run id
- Combined Cursor+Audit timeline tab
- Persisting correlation id on every stream event row
