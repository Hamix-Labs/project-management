// Package policy holds handler-side read limits and SSE publish classification.
// It stays free of HTTP and database imports so CI can enforce purity.
package policy

import (
	"log/slog"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

// EnrichedTaskChangeEvent reports whether a change type should carry the full
// domain.Task in Event.Data after a successful store mutation.
func EnrichedTaskChangeEvent(typ realtime.ChangeType) bool {
	slog.Debug("trace", "operation", "policy.EnrichedTaskChangeEvent")
	switch typ {
	case realtime.TaskCreated, realtime.TaskUpdated:
		return true
	default:
		return false
	}
}

// HintOnlyChangeTypes are published with id (and optional cycle_id) only so
// clients refetch the sidecar resource. See ADR-0026 invariant S3.
var HintOnlyChangeTypes = []realtime.ChangeType{
	realtime.TaskDeleted,
	realtime.TaskGateChanged,
	realtime.TaskDependencyChanged,
	realtime.TaskEventChanged,
	realtime.ProjectCreated,
	realtime.ProjectUpdated,
	realtime.ProjectDeleted,
	realtime.ProjectContextChanged,
	realtime.SettingsChanged,
}

// IsHintOnly reports whether typ is classified as hint-only in the publish policy.
func IsHintOnly(typ realtime.ChangeType) bool {
	slog.Debug("trace", "operation", "policy.IsHintOnly")
	for _, hint := range HintOnlyChangeTypes {
		if typ == hint {
			return true
		}
	}
	return false
}

// ScopelessHintChangeTypes are hint-only events published without task or
// project id. Clients refetch global resources (settings, agent run state).
var ScopelessHintChangeTypes = []realtime.ChangeType{
	realtime.SettingsChanged,
	realtime.AgentRunCancelled,
}

// IsScopelessHint reports whether typ is published via notifyScopelessChange.
func IsScopelessHint(typ realtime.ChangeType) bool {
	slog.Debug("trace", "operation", "policy.IsScopelessHint")
	for _, scopeless := range ScopelessHintChangeTypes {
		if typ == scopeless {
			return true
		}
	}
	return false
}
