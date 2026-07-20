# pkgs/tasks/wire

Wiring-edge composed contracts only (`HandlerAPI`). Implemented by
`internal/taskapi/composition.API`. Handlers should depend on focused BC
contracts when possible; use `HandlerAPI` / `handler.HandlerAPI` at the
composition shell edge.
