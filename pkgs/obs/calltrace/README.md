# `pkgs/obs/calltrace`

Per-request **call stack** for structured logs (`call_path`, `helper.io`): `Push`, `Path`, `WithRequestRoot`, `RunObserved`, and `HelperIOIn` / `HelperIOOut`.

**Consumers:** `pkgs/tasks/handler` (HTTP handlers and JSON helpers), `pkgs/tasks/middleware` (access log `call_path` via `Path` passed into `Stack`), `internal/taskapi` (wires `middleware.Stack(..., calltrace.Path)`).

Constant ownership across Go and web mirrors: [ADR-0035](../../../docs/adr/ADR-0035-cross-stack-constant-ownership.md).

**Dependencies:** stdlib and `log/slog` only in production code; tests use `pkgs/tasks/logctx` for `log_seq` assertions.

| File | Role |
|------|------|
| `const.go` | Shared `LogCmd` (`taskapi`) for slog `cmd` field; helper.io `obs_category` / `phase` constants. |
| `stack.go` | `Push`, `Path`, `WithRequestRoot`. |
| `observe.go` | `RunObserved`, `HelperIOIn`, `HelperIOOut`, internal helper.io emitters. |

See `handler/doc.go` for usage conventions.
