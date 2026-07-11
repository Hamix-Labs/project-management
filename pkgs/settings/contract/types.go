package contract

import "encoding/json"

// SettingsPatch is the partial-update payload for app_settings.
type SettingsPatch struct {
	AgentPaused                 *bool
	Runner                      *string
	CursorBin                   *string
	CursorModel                 *string
	MaxRunDurationSeconds       *int
	StreamIdleStuckSeconds      *int
	AgentPickupDelaySeconds     *int
	DisplayTimezone             *string
	OptimisticMutationsEnabled  *bool
	SSEReplayEnabled            *bool
	RunnerConfigs               *json.RawMessage
	VerifyMaxRetries            *int
	VerifyRunnerName            *string
	VerifyRunnerModel           *string
	VerifyCommandTimeoutSeconds *int
	CursorSessionResumeEnabled  *bool
}

// IsEmpty reports whether the patch has nothing to apply.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p SettingsPatch) IsEmpty() bool {
	return p.AgentPaused == nil &&
		p.Runner == nil &&
		p.CursorBin == nil &&
		p.CursorModel == nil &&
		p.MaxRunDurationSeconds == nil &&
		p.StreamIdleStuckSeconds == nil &&
		p.AgentPickupDelaySeconds == nil &&
		p.DisplayTimezone == nil &&
		p.OptimisticMutationsEnabled == nil &&
		p.SSEReplayEnabled == nil &&
		p.RunnerConfigs == nil &&
		p.VerifyMaxRetries == nil &&
		p.VerifyRunnerName == nil &&
		p.VerifyRunnerModel == nil &&
		p.VerifyCommandTimeoutSeconds == nil &&
		p.CursorSessionResumeEnabled == nil
}
