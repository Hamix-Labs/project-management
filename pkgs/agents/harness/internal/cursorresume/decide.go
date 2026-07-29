package cursorresume

import (
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

// Mode is logged on every runner.Run for ADR-0031 observability.
type Mode string

const (
	ModeFresh    Mode = "fresh"
	ModeContinue Mode = "resume"
	ModeFallback Mode = "resume_fallback"
)

// Facts is the pure input to Decide (I/O pre-resolved).
type Facts struct {
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

// Policy is the pure resume/deny decision before prompt composition.
type Policy struct {
	Mode        Mode
	DenyReason  string
	AllowResume bool
}

// Decide implements the ADR-0031 / ADR-0085 / ADR-0090 session-resume decision table.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func Decide(in Facts) Policy {
	if in.ForceFresh {
		return Policy{Mode: ModeFresh, DenyReason: "resume_failed"}
	}
	if !in.SessionResumeEnabled {
		return Policy{Mode: ModeFresh, DenyReason: "settings_disabled"}
	}
	if in.RetryMode == taskcoredomain.RetryFresh {
		return Policy{Mode: ModeFresh, DenyReason: "retry_fresh"}
	}
	// Interrupt execute path checks tamper before later gates (deny reason preserved).
	if in.ResumeNotice && in.RetryMode != taskcoredomain.RetryResume && in.Phase == cyclesdomain.PhaseExecute && in.ReportTampered {
		return Policy{Mode: ModeFresh, DenyReason: "tamper"}
	}
	if in.Phase == cyclesdomain.PhaseVerify && in.FirstVerifyAfterExecute {
		return Policy{Mode: ModeFresh, DenyReason: "verify_fresh_after_execute"}
	}
	if !in.GitSkipped && in.HasPostExecuteHead && !in.HeadMatchesAnchor {
		return Policy{Mode: ModeFresh, DenyReason: "head_drift"}
	}
	if in.ReportTampered {
		return Policy{Mode: ModeFresh, DenyReason: "tamper"}
	}
	if strings.TrimSpace(in.SessionID) == "" {
		return Policy{Mode: ModeFresh, DenyReason: "no_session_id"}
	}
	if strings.TrimSpace(in.WorkingDir) == "" {
		return Policy{Mode: ModeFresh, DenyReason: "workspace_mismatch"}
	}
	return Policy{Mode: ModeContinue, AllowResume: true}
}

// SessionPhaseForResume returns the phase whose terminal session_id should
// be loaded for --resume. Command-verify always resumes the execute chat
// (ADR-0090).
//
//funclogmeasure:skip category=hot-path reason="Pure phase mapping without I/O."
func SessionPhaseForResume(phase cyclesdomain.Phase) cyclesdomain.Phase {
	if phase == cyclesdomain.PhaseVerify {
		return cyclesdomain.PhaseExecute
	}
	return phase
}
