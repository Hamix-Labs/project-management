package cursorresume

import (
	"errors"
	"fmt"
)

// Stable failure_kind values for same-chat session hard failures (operator UI).
const (
	FailureKindMissingSessionID = "cursor_missing_session_id"
	FailureKindResumeSession    = "cursor_resume_session"
)

// Operator-facing messages persisted as standardized_message / failure_summary.
const (
	MsgMissingSessionAfterExecute = "Cursor finished execute but did not return a chat session id. " +
		"Verification cannot continue in the same chat. Retry the task or check the Cursor CLI."
	MsgMissingSessionForVerify = "No Cursor chat session id from execute is available to resume for verification. " +
		"Hamix did not start a new chat (avoids re-sending full context). Retry the task or Start over."
	MsgResumeSessionFailed = "Cursor could not resume the prior chat session. " +
		"Hamix did not start a new chat (avoids re-sending full context). Retry or Start over."
)

// HardFailError is a product failure that must not soft-fall back to a fresh Cursor chat.
type HardFailError struct {
	Kind    string
	Message string
	Cause   error
}

func (e *HardFailError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Kind != "" {
		return e.Kind
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "cursor session hard failure"
}

func (e *HardFailError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// MissingSessionAfterExecute returns a hard-fail for a successful Cursor execute
// that omitted session_id while resume is enabled.
func MissingSessionAfterExecute() *HardFailError {
	return &HardFailError{
		Kind:    FailureKindMissingSessionID,
		Message: MsgMissingSessionAfterExecute,
	}
}

// MissingSessionForVerify returns a hard-fail when PhaseVerify needs an execute session_id.
func MissingSessionForVerify() *HardFailError {
	return &HardFailError{
		Kind:    FailureKindMissingSessionID,
		Message: MsgMissingSessionForVerify,
	}
}

// ResumeSessionFailed wraps ErrResumeSession (or equivalent) without promising a fresh chat.
func ResumeSessionFailed(cause error) *HardFailError {
	return &HardFailError{
		Kind:    FailureKindResumeSession,
		Message: MsgResumeSessionFailed,
		Cause:   cause,
	}
}

// AsHardFail extracts *HardFailError from err.
func AsHardFail(err error) (*HardFailError, bool) {
	var hf *HardFailError
	if errors.As(err, &hf) && hf != nil {
		return hf, true
	}
	return nil, false
}

// DetailsMap returns failure_kind + standardized_message for phase details_json.
func (e *HardFailError) DetailsMap() map[string]any {
	if e == nil {
		return nil
	}
	out := map[string]any{}
	if e.Kind != "" {
		out["failure_kind"] = e.Kind
	}
	if e.Message != "" {
		out["standardized_message"] = e.Message
	}
	return out
}

// FormatReason returns a stable cycle terminate reason code.
func (e *HardFailError) FormatReason() string {
	if e == nil || e.Kind == "" {
		return "cursor_session_hard_fail"
	}
	return e.Kind
}

// Explain returns Message or a fallback for summaries.
func (e *HardFailError) Explain() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("%s", e.Kind)
}
