package domain

// RunStatus is the observable state of a run inside a draft-assist session.
// The wire form uses the lower-case string; the enum values below are the
// canonical spellings the handler and runner exchange.
type RunStatus string

const (
	// RunStatusIdle: no run is active on the session.
	RunStatusIdle RunStatus = "idle"
	// RunStatusThinking: run accepted, model has not produced tokens yet.
	RunStatusThinking RunStatus = "thinking"
	// RunStatusStreaming: model is emitting tokens.
	RunStatusStreaming RunStatus = "streaming"
	// RunStatusTool: model is executing an MCP tool.
	RunStatusTool RunStatus = "tool"
	// RunStatusCancelling: cancel accepted; terminal cancelled frame follows.
	RunStatusCancelling RunStatus = "cancelling"
	// RunStatusCancelled: run was cancelled by the operator.
	RunStatusCancelled RunStatus = "cancelled"
	// RunStatusDone: run completed successfully.
	RunStatusDone RunStatus = "done"
	// RunStatusFailed: run terminated with an error.
	RunStatusFailed RunStatus = "failed"
)

// IsTerminal reports whether s is a run-terminating state.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s RunStatus) IsTerminal() bool {
	switch s {
	case RunStatusCancelled, RunStatusDone, RunStatusFailed:
		return true
	default:
		return false
	}
}
