# ADR-0065: Shared handlerhttp JSON helpers

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

`pkgs/taskcore/handler/http_helpers.go` and `pkgs/tasks/handler/handler_http_json.go` were near-identical (~364 lines each). BC handlers duplicated the same JSON encode/decode, ETag, and store-error mapping patterns.

## Decision

1. Extract shared helpers to **`pkgs/tasks/handlerhttp`** (alongside existing `pkgs/tasks/apijson`).
2. **`pkgs/tasks/handler`** keeps thin unexported shims for same-package callers; re-exports `WrapPrometheusHandler` for `cmd/taskapi`.
3. **`pkgs/taskcore/handler`** deletes `http_helpers.go`; uses `http_shim.go` forwarding to `handlerhttp`.
4. BC handlers may import `handlerhttp` but must not import `pkgs/tasks/handler`.

## Consequences

### Positive

- Single source for store-error HTTP mapping referenced by `pkgs/repo` comments.
- Future BC `http_helpers` dedup can migrate incrementally to `handlerhttp`.

### Negative / Trade-offs

- `taskcore/handler` still carries a small shim file until call sites import `handlerhttp` directly.

## See also

- Tier 5A PR1 ([#191](https://github.com/AlexsanderHamir/Hamix/pull/191)); later Phase 4 DRY — [#218](https://github.com/AlexsanderHamir/Hamix/pull/218)–[#222](https://github.com/AlexsanderHamir/Hamix/pull/222)
