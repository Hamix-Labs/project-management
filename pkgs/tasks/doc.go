// Package tasks is the root import path for the taskapi composition shell.
// Domain CRUD lives in sibling bounded contexts (taskcore, taskcycles, …).
// Subpackages under this path:
//
//   - handler — REST/SSE mux; registers BC routes; bootstrap, health, RUM, writepolicy/readpolicy
//   - handlerhttp — shared HTTP JSON/path/limit helpers for BC handlers
//   - middleware — outer HTTP stack (auth, rate limit, idempotency, …)
//   - postgres — GORM open + AutoMigrate orchestration (BC models registered here)
//   - realtime — SSE wire types, Publisher, coalesce (hub transport ownership evolving)
//   - scheduling — worker readiness / pickup predicates
//   - apijson, logctx — JSON errors and request log context
//   - (call stack / helper.io: pkgs/obs/calltrace)
//   - service — HTTP-agnostic bootstrap/git/retry orchestration used by the shell
//   - wire — HandlerAPI composition interface (implemented by internal/taskapi/composition)
//   - devsim — HAMIX_SSE_TEST synthetic events
//
// Typical wiring (see cmd/taskapi):
//
//	db, err := postgres.Open(dsn, nil)
//	if err != nil { ... }
//	if err := postgres.Migrate(ctx, db); err != nil { ... }
//	s := composition.NewAPI(db)
//	hub := realtime.NewSSEHub()
//	http.Handler = handler.NewHandler(s, hub, repoRootOrNil)
package tasks
