# `pkgs/storekernel`

Shared persistence helpers used by bounded-context stores and `internal/taskapi/composition`. Extracted from the retired `pkgs/tasks/kernel` per [ADR-0050](../../docs/adr/ADR-0050-storekernel-extraction.md). Composition wiring lives in [ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md) — there is no `pkgs/tasks/store` facade.

## Responsibilities

| Area | Files | Role |
| --- | --- | --- |
| Metrics | `metrics.go` | `DeferLatency`, `Op*` Prometheus label constants |
| Audit events | `events.go`, `tx.go` | `NextEventSeq`, `AppendEvent`, transactional append |
| Validation | `validate.go`, `constraints.go` | `ValidateActor`, enum/check helpers |
| IDs + errors | `ids.go`, `errors.go`, `persistence_errors.go` | `ResolveID`, `MapNotFound`, duplicate-key mapping |
| JSON | `json.go`, `jsonmap/` | `NormalizeJSONObject`, `EventPairJSON`; RawMessage↔datatypes helpers in `jsonmap` |
| Field parity | `parity/` | Domain↔model field-parity assertion helpers (`AssertFieldParity`) |
| Task load | `taskload/` | Shared GORM task-row load helpers |

## Dependency rules

| May import | Must not import |
| --- | --- |
| `pkgs/taskcore/domain`, `pkgs/taskcore/store/model`, `pkgs/taskcycles/domain`, `pkgs/taskevents/domain` + `store/model`, `pkgs/obs/calltrace`, stdlib, GORM | `pkgs/tasks/handler`, BC `handler` packages, `internal/taskapi/composition` |

`MapNotFound` / `MapWriteError` emit **taskcore** sentinels (`ErrNotFound`, `ErrConflict`, `ErrInvalidInput`) so HTTP mappers can use `errors.Is` without importing GORM.

## Tests

```powershell
go test ./pkgs/storekernel/... -count=1
```

## See also

- [docs/agent-map.md](../../docs/agent-map.md) — persistence row (composition + BC stores)
- [ADR-0045](../../docs/adr/ADR-0045-bounded-context-projects.md) — first BC that shared kernel debt
- [ADR-0079](../../docs/adr/ADR-0079-facade-deletion.md) — facade deletion / composition root
