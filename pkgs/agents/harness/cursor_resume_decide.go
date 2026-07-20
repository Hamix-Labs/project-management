package harness

import (
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// CursorResumeFacts is the pure input to DecideCursorResume (I/O pre-resolved).
type CursorResumeFacts struct {
	ForceFresh              bool
	SessionResumeEnabled    bool
	RetryMode               taskcoredomain.RetryMode
	Phase                   cyclesdomain.Phase
	ResumeNotice            bool
	ReportTampered          bool
	FirstVerifyAfterExecute bool
	GitSkipped              bool
	HasPostExecuteHead      bool
	HeadMatchesAnchor       bool
	SessionID               string
	WorkingDir              string
}

// CursorResumePolicy is the pure resume/deny decision before prompt composition.
type CursorResumePolicy struct {
	Mode        CursorResumeMode
	DenyReason  string
	AllowResume bool
}

// DecideCursorResume implements the ADR-0031 session-resume decision table.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func DecideCursorResume(in CursorResumeFacts) CursorResumePolicy {
	if in.ForceFresh {
		return CursorResumePolicy{Mode: CursorResumeFresh, DenyReason: "resume_failed"}
	}
	if !in.SessionResumeEnabled {
		return CursorResumePolicy{Mode: CursorResumeFresh, DenyReason: "settings_disabled"}
	}
	if in.RetryMode == taskcoredomain.RetryFresh {
		return CursorResumePolicy{Mode: CursorResumeFresh, DenyReason: "retry_fresh"}
	}
	// Interrupt execute path checks tamper before later gates (deny reason preserved).
	if in.ResumeNotice && in.RetryMode != taskcoredomain.RetryResume && in.Phase == cyclesdomain.PhaseExecute && in.ReportTampered {
		return CursorResumePolicy{Mode: CursorResumeFresh, DenyReason: "tamper"}
	}
	if in.Phase == cyclesdomain.PhaseVerify && in.FirstVerifyAfterExecute {
		return CursorResumePolicy{Mode: CursorResumeFresh, DenyReason: "verify_fresh_after_execute"}
	}
	if !in.GitSkipped && in.HasPostExecuteHead && !in.HeadMatchesAnchor {
		return CursorResumePolicy{Mode: CursorResumeFresh, DenyReason: "head_drift"}
	}
	if in.ReportTampered {
		return CursorResumePolicy{Mode: CursorResumeFresh, DenyReason: "tamper"}
	}
	if strings.TrimSpace(in.SessionID) == "" {
		return CursorResumePolicy{Mode: CursorResumeFresh, DenyReason: "no_session_id"}
	}
	if strings.TrimSpace(in.WorkingDir) == "" {
		return CursorResumePolicy{Mode: CursorResumeFresh, DenyReason: "workspace_mismatch"}
	}
	return CursorResumePolicy{Mode: CursorResumeContinue, AllowResume: true}
}

//funclogmeasure:skip category=hot-path reason="Pure state comparison for verify fresh-after-execute deny."
func firstVerifyAfterNewExecute(state *processState) bool {
	return state.phase.lastVerifyAfterExecuteSeq < state.phase.lastCompletedExecutePhaseSeq
}
