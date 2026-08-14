package domain

import "time"

// EventKind is the SSE named-event key emitted on the wire.
type EventKind string

const (
	// EventSession is emitted once when the SSE stream attaches; carries the
	// session snapshot so the SPA can reconcile server state before Send.
	EventSession EventKind = "session"
	// EventStatus is a run-status transition (thinking / streaming / …).
	EventStatus EventKind = "status"
	// EventToken carries assistant chat-thread text.
	EventToken EventKind = "token"
	// EventTool is an MCP tool-call trace (kind = start / end).
	EventTool EventKind = "tool"
	// EventPatch is a prompt-mutation the SPA must apply in place.
	EventPatch EventKind = "patch"
	// EventError is a run-level error before Done.
	EventError EventKind = "error"
	// EventDone is the final frame of a run.
	EventDone EventKind = "done"
	// EventHeartbeat is the SSE keep-alive comment (empty data payload).
	// The handler writes a comment line, not a named event.
	EventHeartbeat EventKind = "heartbeat"
)

// Event is the wire envelope for a draft-assist SSE frame.
// Each named event has a small payload struct in Data; empty for done.
type Event struct {
	// ID is the monotonic per-session sequence used for Last-Event-ID replay.
	ID uint64 `json:"id"`
	// Kind selects the SSE event name.
	Kind EventKind `json:"kind"`
	// RunID is the run this event belongs to. Empty for the initial session
	// event and heartbeats.
	RunID string `json:"run_id,omitempty"`
	// At is the server clock at emit time; used for latency diagnostics.
	At time.Time `json:"at"`
	// Data is the kind-specific payload. Handler JSON-encodes it as-is.
	Data any `json:"data,omitempty"`
}

// SessionEventData is the payload for EventSession.
type SessionEventData struct {
	SessionID     string       `json:"session_id"`
	WorktreeID    string       `json:"worktree_id,omitempty"`
	Snapshot      FormSnapshot `json:"snapshot"`
	SchemaVersion int          `json:"schema_version"`
}

// StatusEventData is the payload for EventStatus.
type StatusEventData struct {
	Status RunStatus `json:"status"`
	Reason string    `json:"reason,omitempty"`
}

// TokenEventData is the payload for EventToken.
type TokenEventData struct {
	Delta string `json:"delta"`
}

// ToolEventData is the payload for EventTool.
type ToolEventData struct {
	Name  string `json:"name"`
	Phase string `json:"phase"` // "start" | "end"
	OK    bool   `json:"ok,omitempty"`
	Error string `json:"error,omitempty"`
}

// PatchOp is a bounded operation the MCP prompt tools emit.
type PatchOp string

const (
	// PatchOpSet replaces the whole prompt with Value.
	PatchOpSet PatchOp = "set"
	// PatchOpFindReplace does one find/replace substitution.
	PatchOpFindReplace PatchOp = "find_replace"
	// PatchOpAppend appends Value to the prompt.
	PatchOpAppend PatchOp = "append"
)

// PatchEventData is the payload for EventPatch.
type PatchEventData struct {
	Op      PatchOp `json:"op"`
	Find    string  `json:"find,omitempty"`
	Value   string  `json:"value,omitempty"`
	Summary string  `json:"summary,omitempty"`
}

// ErrorEventData is the payload for EventError.
type ErrorEventData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DoneEventData is the payload for EventDone. Status is the terminal state
// (done / cancelled / failed).
type DoneEventData struct {
	Status RunStatus `json:"status"`
}
