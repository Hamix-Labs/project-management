# `internal/handlertest`

Black-box HTTP tests for [`pkgs/tasks/handler`](../../pkgs/tasks/handler): only the exported `handler` API, `httptest`, and `net/http`. [`server.go`](./server.go) duplicates the former `newTaskTestServer*` helpers so production code does not carry test-only exports. Baseline security header assertions live in [`internal/httpsecurityexpect`](../httpsecurityexpect/) so `handler` tests can share them **without** an import cycle (`handler` → `handlertest` → `handler`).

**Whitebox tests** (unexported symbols, `decodeJSON`, path helpers) stay next to the code under `pkgs/tasks/handler`. See [`pkgs/tasks/handler/README.md`](../../pkgs/tasks/handler/README.md).

## Additional exported helpers ([`bound_server.go`](./bound_server.go))

Re-usable test infrastructure for `package middleware_test` files and any future black-box suites that need a fully-wired task handler with a seeded git repo:

| Symbol | Purpose |
|--------|---------|
| `NewBoundServer(t, build)` | `httptest.Server` backed by `BoundTaskHandler` with seeded git binding registered under the server URL. |
| `NewDirectBoundHandler(t, build)` | Handler (no HTTP server) with git binding registered under `DirectHandlerTestURL`. |
| `BoundTaskHandler(st, opts...)` | `handler.NewHandler` + `NewSSEHub` + `NewSettingsRepoProvider`. |
| `WithCreateChecklistForURL(baseURL, body)` | Injects `checklist_items` and, when a git binding exists, `project_id` + `worktree_id`. |
| `DirectHandlerTestURL` | Sentinel base URL for `NewDirectBoundHandler`-style tests. |
| `TestCriterionText` | Default checklist item text used by `WithCreateChecklistForURL`. |
| `JSONErrorBody` | `{Error string}` struct for decoding API error envelopes. |
| `GitBinding` | Struct holding `RepositoryID`, `ProjectID`, `WorktreeID` for a seeded repo. |
| `GitBindingForURL(url)` | Looks up the registered binding for a base URL. |
| `DrainSSE(t, ch, want, timeout)` | Collects up to `want` SSE events from a string channel within a deadline. |
| `SummarizeSSEEvents(events)` | Collapses `[]realtime.Event` into a stable sorted `[]string` for comparison. |
