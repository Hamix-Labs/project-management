// Package kernel holds shared store helpers: Prometheus latency histogram
// (DeferLatency + Op* constants), audit-log primitives, domain validators,
// and generic transactional helpers. Extracted from store/internal so
// bounded contexts (pkgs/projects) can share the same kernel until
// pkgs/storekernel/ exists.
package kernel

// LogCmd is the cmd label every kernel slog call uses, mirroring the
// historical "taskapi" tag from the original pkgs/tasks/store package.
// Sub-packages that wrap kernel helpers should set their own cmd label
// when they wish to differentiate; the kernel itself is considered part
// of the same operational surface.
