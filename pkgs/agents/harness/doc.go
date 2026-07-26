// Package harness orchestrates one task execution cycle around a runner.
//
// A harness is everything wrapped around runner.Run to turn "run a prompt
// in a working directory" into a trustworthy, auditable unit of work:
//
//   - Phase choreography: StartCycle → execute → (verify → execute)* → TerminateCycle
//   - Prompt engineering: criteria injection, verify feedback
//   - Agent↔worker report-file contracts (criteria-report.json, verify-report.json)
//   - Adversarial verification and git integrity (tamper detection)
//   - Crash/shutdown recovery of in-flight cycle state
//
// The queue consumer (pkgs/agents/worker) handles admission — reload,
// readiness, ready→running transition, ack ordering — then delegates to
// Harness.Run for the cycle body. See docs/architecture.md and
// docs/adr/ADR-0005-extract-agent-harness.md.
package harness
