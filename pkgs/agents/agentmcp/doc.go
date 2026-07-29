// Package agentmcp is the Hamix agent MCP host: bind-scoped sessions, a tool
// registry, and stdio MCP serving. Tools may import pkgs/agents/sidecar only
// (plus stdlib / MCP SDK) — no store or task HTTP.
//
// Extension: add a file under tools/, implement Tool, and register it in
// DefaultTools. Do not import harness.
package agentmcp
