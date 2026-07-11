# ADR-0048: Extract `pkgs/taskcompose` bounded context

**Date:** 2026-07-11  
**Status:** Accepted  
**Deciders:** Hamix maintainers

## Context

Task drafts and templates shared `namedpayload` persistence, compose JSON shapes, and HTTP routes under `/task-drafts*` and `/task-templates*`. They lived in `pkgs/tasks/store/internal/{drafts,templates,namedpayload}` and `pkgs/tasks/handler`. Projects, gitinventory, and settings extractions (ADR-0045–0047) established the composition pattern; drafts + templates are one bounded context because they share payload normalization and GORM models.

Template instantiate still creates tasks — that path stays in `pkgs/tasks/handler` via an injected callback (same pattern as settings `AgentWorkerControl`).

## Decision

1. **New bounded context** — `pkgs/taskcompose/{domain,contract,store,handler}` owns draft/template domain types, GORM models, persistence, and HTTP routes (10 routes including `POST /task-templates/instantiate`).

2. **Composition root unchanged** — `pkgs/tasks/store.Store` embeds `pkgs/taskcompose/store.Store` and delegates through `facade_compose.go`.

3. **Route registration** — `pkgs/taskcompose/handler.Register` mounts routes from `pkgs/tasks/handler/handler_routes.go`. No path or JSON shape changes.

4. **Contract kernel** — `pkgs/tasks/contract` keeps `DraftStore` and `TemplateStore` with wire types aliased from `pkgs/taskcompose/contract` so bootstrap and handler fakes compile unchanged.

5. **Create-from-draft** — `composestore.DeleteDraftByIDInTx` is exported for `pkgs/tasks/store/internal/tasks` CRUD.

6. **Import gate** — CI rejects `pkgs/taskcompose` importing `pkgs/tasks/handler` or `pkgs/tasks/store/internal` (`scripts/check-go.sh` → `step_taskcompose_boundary`).

7. **Postgres migrate** — `pkgs/tasks/store/model.AutoMigrateAll` imports GORM models from `pkgs/taskcompose/store/model`.

## Consequences

### Positive

- Draft/template ownership is explicit; namedpayload is no longer buried under tasks store internals.
- Template instantiate coupling is documented and injected at registration time.

### Negative / trade-offs

- Compose payload validation for template save/patch remains in tasks handler (`normalizeComposePayloadRaw`); taskcompose receives a callback.
- Cross-route HTTP contract tests remain in `pkgs/tasks/handler` because they exercise the full taskapi mux.

## See also

- [pkgs/taskcompose/README.md](../../pkgs/taskcompose/README.md)
- [ADR-0047](./ADR-0047-bounded-context-settings.md)
