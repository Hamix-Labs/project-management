# pkgs/runners

HTTP-only surface for the SPA runner registry (`/runners*`).

## Naming

| Package | Role |
| --- | --- |
| **`pkgs/agents/runner`** | `Runner` interface, capability interfaces (`Configurer`, `Attributor`, …), cursor adapter, registry |
| **`pkgs/runners`** (this package) | HTTP `/runners*` only — probe, list-models, config schema/validate |

Do not add CLI adapters here. See [ADR-0052](../../docs/adr/ADR-0052-runners-http-handler.md) and [handler/README.md](./handler/README.md).

```powershell
go test ./pkgs/runners/... -count=1
```
