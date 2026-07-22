package harness

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/cursorresume"
)

// CursorResumeMode is logged on every runner.Run for ADR-0031 observability.
type CursorResumeMode = cursorresume.Mode

const (
	CursorResumeFresh    = cursorresume.ModeFresh
	CursorResumeContinue = cursorresume.ModeContinue
	CursorResumeFallback = cursorresume.ModeFallback
)

// CursorResumeFacts is the pure input to DecideCursorResume (I/O pre-resolved).
type CursorResumeFacts = cursorresume.Facts

// CursorResumePolicy is the pure resume/deny decision before prompt composition.
type CursorResumePolicy = cursorresume.Policy

// DecideCursorResume implements the ADR-0031 session-resume decision table.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func DecideCursorResume(in CursorResumeFacts) CursorResumePolicy {
	return cursorresume.Decide(in)
}
