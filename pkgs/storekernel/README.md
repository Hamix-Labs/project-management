# `pkgs/storekernel`

Shared persistence helpers used by `pkgs/tasks/store` and extracted bounded-context stores. Extracted from `pkgs/tasks/kernel` per [ADR-0050](../../docs/adr/ADR-0050-storekernel-extraction.md).

## Responsibilities

| Area | Files | Role |
| --- | --- | --- |
| Metrics | `metrics.go` | `DeferLatency`, `Op*` Prometheus label constants |
| Audit events | `events.go`, `tx.go` | `NextEventSeq`, `AppendEvent`, transactional append |
| Validation | `validate.go`, `constraints.go` | `ValidateActor`, enum/check helpers |
| IDs + errors | `ids.go`, `errors.go`, `persistence_errors.go` | `ResolveID`, `MapNotFound`, duplicate-key mapping |
| JSON | `json.go` | `NormalizeJSONObject`, `EventPairJSON` |

## Dependency rules

| May import | Must not import |
| --- | --- |
| `pkgs/tasks/domain`, `pkgs/tasks/store/model`, `pkgs/tasks/calltrace`, stdlib, GORM | `pkgs/tasks/store/internal`, `pkgs/tasks/handler`, BC handler packages |

## Tests

```powershell
go test ./pkgs/storekernel/... -count=1
```

## See also

- [pkgs/tasks/store/README.md](../tasks/store/README.md) — facade concern map
- [ADR-0045](../docs/adr/ADR-0045-bounded-context-projects.md) — first BC that shared kernel debt
