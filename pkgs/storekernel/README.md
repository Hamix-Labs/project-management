# `pkgs/storekernel`

Domain-agnostic persistence primitives shared by bounded-context stores.
Extracted from the retired `pkgs/tasks/kernel` per [ADR-0050](../../docs/adr/ADR-0050-storekernel-extraction.md).
Composition wiring lives in [ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md) — there is no `pkgs/tasks/store` facade.

Task-specific adapters live with their owning BC:

| Concern | Home |
| --- | --- |
| Status / priority / actor validators | `pkgs/taskcore/domain` |
| Phase validators | `pkgs/taskcycles/domain` |
| Audit seq + event insert | `pkgs/taskevents/store/audit` |
| Load task row in a TX | `pkgs/taskcore/store/taskload` |

## Responsibilities

| Area | Files | Role |
| --- | --- | --- |
| Metrics | `metrics.go` | `DeferLatency`, `Op*` Prometheus label constants |
| IDs | `ids.go` | `ResolveID` |
| Constraint classifiers | `constraints.go` | `IsDuplicateKey`, FK/check helpers |
| Error mapping | `errors.go`, `persistence_errors.go` | `MapNotFound` / `MapWriteError` / `MapPayloadPersistenceError` (callers pass BC sentinels) |
| JSON | `json.go`, `jsonmap/` | `NormalizeJSONObject` (callers pass invalid sentinel), `EventPairJSON`; RawMessage↔datatypes helpers |
| Field parity | `parity/` | Domain↔model field-parity assertion helpers (`AssertFieldParity`) |

## Dependency rules

| May import | Must not import |
| --- | --- |
| stdlib, GORM, Prometheus, `google/uuid`, `pkgs/obs/calltrace` (tests may use fixtures) | Any domain BC (`taskcore`, `taskcycles`, `taskevents`, `tasks`, `projects`, …), BC `handler` packages, `internal/taskapi/composition` |

`MapNotFound` / `MapWriteError` / `NormalizeJSONObject` take the caller's BC sentinels so projects can emit `projects/domain` errors while task paths keep `taskcore/domain` errors.

## Tests

```powershell
go test ./pkgs/storekernel/... -count=1
```

## See also

- [docs/agent-map.md](../../docs/agent-map.md) — persistence row (composition + BC stores)
- [ADR-0045](../../docs/adr/ADR-0045-bounded-context-projects.md) — first BC that shared kernel debt
- [ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md) — facade deletion / composition root
