// Package sidecar implements the agent↔worker side-channel JSON protocol under
// ReportDir/<cycle_id>/ (criteria-report.json, verify-report.json, submit receipts).
// It is a leaf package: no store or runner imports. Shared by the harness and
// the agent MCP host so parsers cannot drift.
package sidecar
