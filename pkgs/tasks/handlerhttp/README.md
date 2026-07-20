# pkgs/tasks/handlerhttp

Shared HTTP helpers for BC handlers: JSON encode/decode, path IDs, query
limits, `DebugHTTPRequest` / `DebugHTTPOut`, store-error mapping.

Call this package directly from handlers (no per-BC thin wrappers). See
[ADR-0077](../../../docs/adr/ADR-0077-bc-handlerhttp-migration.md).
