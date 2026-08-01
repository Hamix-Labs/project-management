// Package taskapi wires the instrumented task HTTP handler for hosts: the
// standard middleware stack around pkgs/tasks/handler.NewHandler. Shared
// process bootstrap (DB + agents + NewHTTPHandler) lives in
// internal/taskapiruntime to avoid an import cycle with agentworker.
//
// Env-driven startup flags live in internal/taskapiconfig (see
// docs/configuration.md). Agent worker supervisor lifecycle lives in
// internal/taskapi/agentworker.
package taskapi
