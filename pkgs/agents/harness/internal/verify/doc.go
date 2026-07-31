// Package verify runs claim-only criteria acceptance after execute: load the
// execute self-report, accept claimed_done as execute_claim (including criteria
// with operator verify_commands the agent was instructed to self-check), and
// persist verdict mirrors. Completions are applied by the harness effect layer.
package verify
