# ADR-0077: BC handlers migrate to pkgs/tasks/handlerhttp

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Post–Tier 5 bounded-context split left eight BC `http_helpers.go` files duplicating JSON encode/decode, error envelopes, store-error mapping, and `X-Actor` parsing. `pkgs/tasks/handlerhttp` already centralizes that behavior for the composition shell (`pkgs/tasks/handler`) and a thin `taskcore/handler/http_shim.go` forwarder layer.

Duplication drifted observability (http.io logging, helper.io traces) and made contract fixes require N copies.

## Decision

1. **Canonical HTTP JSON helpers** live in `pkgs/tasks/handlerhttp` (`WriteJSON`, `WriteJSONError`, `WriteError`, `DecodeJSON`, `ActorFromRequest`, `WriteStoreError`, `StoreErrHTTPResponse`, `WriteJSONWithETag`, etc.).
2. **BC handlers** (`repo`, `runners`, `taskcompose`, `projects`, `settings`, `taskchecklist`, `taskcycles`, `taskcore`) call `handlerhttp` directly from handler files.
3. **BC `http_helpers.go`** keeps only BC-specific path parsers, query parsers, debug http.io helpers, and error mappers that are not generic (e.g. `repoErrUserMessage`, projects/settings `writeStoreError` for BC domain sentinels, gitinventory `GitErrHTTP` / `WriteGitStoreError`).
4. **gitinventory exception:** `handlerhttp.WriteStoreError` delegates git-coded errors to `gitinventory/handler.WriteGitStoreError`, creating an import cycle if git handlers import `handlerhttp`. Git handlers retain local thin JSON helpers in `http_helpers.go` that call `pkgs/tasks/apijson` directly (same wire shape, documented exception).
5. **Composition shell cleanup:** delete `taskcore/handler/http_shim.go`; slim `pkgs/tasks/handler/handler_http_json.go` to `WrapPrometheusHandler` re-export and test-only `jsonErrorBody` type.
6. **taskevents** inline JSON helpers remain for a follow-up (PR9).

## Consequences

### Positive

- One place to fix JSON security headers, request-id error envelopes, and decode validation.
- BC handler files read consistently; `http_helpers.go` files shrink to domain-specific glue.
- `handlerhttp` doc comment matches actual import graph.

### Negative / Trade-offs

- gitinventory cannot use `handlerhttp` until git error dispatch is inverted (e.g. callback registry or moving `WriteGitStoreError` into `handlerhttp`).
- BCs with their own domain sentinels (`projects`, `settings`) keep small `writeStoreError` wrappers until those errors unify on `taskcore/domain` sentinels.

## Alternatives

- **Keep per-BC copies** — rejected; caused the drift this ADR removes.
- **Move git helpers into handlerhttp now** — deferred; larger blast radius than Child 2 scope.

## See also

- [ADR-0070](./ADR-0070-taskapi-shell-ownership.md) — composition shell vs BC handlers
- `pkgs/tasks/handlerhttp/doc.go`
