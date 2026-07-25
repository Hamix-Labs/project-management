# ADR-0035: Cross-Stack Constant Ownership

> **Note** — Product renamed T2A to Hamix; identifiers below reflect the name at decision time unless updated inline.

**Date:** 2026-06-23
**Status:** Accepted
**Deciders:** Engineering

## Context

Hardcoded wire values and observability field strings had drifted across Go and TypeScript (`taskapi` slog `cmd`, SSE change types, audit event types, RUM types, verifier kinds). A prior bulk refactor consolidated duplicate Go `cmd` literals; web cycles A1–A5 added typed mirrors and drift-guard tests. We still needed a single documented map of **who owns which constant** and which values are **intentionally omitted** from UI or server promotion.

## Decision

### Ownership map

| Surface | Authoritative source | Web mirror (when applicable) | Drift guard |
| --- | --- | --- | --- |
| slog `cmd` for taskapi HTTP/store/agent traces | [`pkgs/obs/calltrace/const.go`](../../pkgs/obs/calltrace/const.go) `LogCmd` | — | `logctx.TraceCmd` aliases `calltrace.LogCmd` |
| helper.io `obs_category` / `phase` | [`pkgs/obs/calltrace/const.go`](../../pkgs/obs/calltrace/const.go) | — | `pkgs/obs/calltrace/observe_test.go` |
| http_io observability | [`pkgs/tasks/handlerhttp/httplog.go`](../../pkgs/tasks/handlerhttp/httplog.go) | — | handler observability tests |
| SSE `change` types | [`pkgs/tasks/realtime/wire.go`](../../pkgs/tasks/realtime/wire.go) | [`web/src/types/task.ts`](../../web/src/types/task.ts) `SSE_CHANGE_TYPES` | `pkgs/tasks/realtime/wire_test.go`, `web/src/types/contractManifest.test.ts` |
| Audit / task event types | [`pkgs/tasks/domain/enums.go`](../../pkgs/tasks/domain/enums.go) | [`web/src/types/task.ts`](../../web/src/types/task.ts) `TASK_EVENT_TYPES` | `pkgs/tasks/domain/event_types_manifest_test.go`, `contractManifest.test.ts` |
| RUM promoted types | [`pkgs/tasks/handler/handler_rum.go`](../../pkgs/tasks/handler/handler_rum.go) `validRUMTypes` | [`web/src/api/rum.ts`](../../web/src/api/rum.ts) `RUM_PROMOTED_TYPES` | handler RUM tests |
| RUM forward-compat (client-only) | — | `RUM_FORWARD_COMPAT_TYPES` in `rum.ts` | documented only; no server mirror until promoted |
| Verifier kinds | [`pkgs/tasks/domain/verifier_kind.go`](../../pkgs/tasks/domain/verifier_kind.go) | [`web/src/types/cycle.ts`](../../web/src/types/cycle.ts) `VERIFIER_KINDS` | parser + cycle types |
| Cycle failure sort params | store stats query contract | [`web/src/constants/api.ts`](../../web/src/constants/api.ts) `CYCLE_FAILURE_SORTS` | API parser tests |

**Rules:**

1. Go domain/realtime packages own wire enums consumed by HTTP or SSE.
2. `calltrace` owns **process identity** (`LogCmd`) and **helper.io taxonomy** only — not HTTP `http_io` or business enums.
3. Web mirrors live next to parsers (`web/src/types/`, `web/src/api/`, `web/src/constants/`) and must not reintroduce literals in feature code.
4. Agent adapter process labels (e.g. `adapterLogCmd = "claudecode"`) stay local — they are not `calltrace.LogCmd`.

### Intentional omission registry

Values may exist on the wire or in parsers but are **not** fully surfaced in product UI or server promotion. Cross-links:

| Omission | Registry | Notes |
| --- | --- | --- |
| Launch UI gates (projects, task tags, deps/milestone, gates, schedule) | [`docs/omitted-features.md`](../omitted-features.md), [`web/src/launch/omittedFeatures.ts`](../../web/src/launch/omittedFeatures.ts) | Backend routes remain for tests |
| `on_task_done` audit event | parser + generic label only | No PR/commit panel UI |
| `subtask_added`, `task_type` event types | excluded from manifest tests | Removed per ADR-0010 / ADR-0011 |
| `navigation_timing` RUM | `RUM_FORWARD_COMPAT_TYPES` in web only | Server `validRUMTypes` unchanged until ADR-0026 Phase 2 |
| SSE hint-only enrichment | ADR-0026 S3 | `task_gate_changed`, `task_dependency_changed`, etc. stay id-only |

Drift guards (`event_types_manifest_test.go`, `contractManifest.test.ts`) encode these omissions as **negative cases** — restoring a removed enum or promoting RUM types requires an explicit ADR/doc update, not a silent mirror change.

### RUM forward-compat vs promotion

[`web/src/api/rum.ts`](../../web/src/api/rum.ts) may list client-only types in `RUM_FORWARD_COMPAT_TYPES` so the SPA can emit measurements before the server accepts them. Promotion to `validRUMTypes` in `handler_rum.go` is a separate, deliberate step documented against [ADR-0026](./ADR-0026-backend-data-coherence.md) Phase 2 — not automatic when the web constant appears.

## Consequences

### Positive

- Contributors know where to add a new enum or log field without guessing package boundaries.
- Drift tests and this ADR share the same omission list — fewer accidental UI/API restores.
- B1/B2 calltrace consolidation has a documented scope ceiling (`http_io` stays in handler).

### Negative / Trade-offs

- Web mirrors require paired updates when Go wire enums change — mitigated by manifest tests.
- Forward-compat RUM types can confuse readers until promotion; comments in `rum.ts` and `handler_rum.go` point here.

## Alternatives Considered

| Alternative | Reason Rejected |
| --- | --- |
| Single generated OpenAPI/JSON schema for all enums | High setup cost; current manifest tests suffice at this scale |
| Put all observability strings in `calltrace` | Would pull handler HTTP concerns into a stdlib-only helper package |
| Auto-sync web from Go via codegen in CI | Deferred until enum churn justifies tooling |
