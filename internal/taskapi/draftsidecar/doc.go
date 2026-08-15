// Package draftsidecar owns the lifetime of the hamix-draft-agent sidecar
// process and adapts its loopback HTTP+SSE surface to draftassist
// contract.Runner. ResolveBinary locates the launcher (HAMIX_DRAFT_AGENT_BIN,
// a sibling of the current executable, then exec.LookPath). The supervisor
// spawns it with --port 0 so a free port is chosen, parses the
// "listening on <port>" line from stdout to learn the port, then keeps a
// health probe running against GET /readyz. On crashes it respawns with
// exponential backoff (1s / 2s / 5s / 15s cap) and bumps the
// taskapi_draftassist_sidecar_restart_total counter.
//
// CURSOR_API_KEY reaches the child only via the process environment; it is
// never logged, echoed to stderr, or persisted. Supervisor also implements
// draftassist/handler.ReadyProbe so /draft-assist/ready surfaces
// missing_key / sidecar_down / no_runner using the shared constants in
// pkgs/draftassist/metrics.
//
// MustHost is the production boot gate: resolve binary, require
// CURSOR_API_KEY, spawn, wait until /readyz is ready. There is no fake
// fallback; the caller must fail to serve on error.
package draftsidecar
