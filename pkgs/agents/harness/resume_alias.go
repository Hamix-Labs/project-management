package harness

import (
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/resume"
)

// ContinuationBundle rehydrates cross-cycle resume context from a parent attempt.
type ContinuationBundle = resume.ContinuationBundle

type resumeCheckpoint = resume.Checkpoint
type resumeEntry = resume.Entry

const (
	resumeEntryExecute             = resume.EntryExecute
	resumeEntryVerifyOnly          = resume.EntryVerifyOnly
	resumeEntryAfterExecuteSuccess = resume.EntryAfterExecuteSuccess
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Harness) resumeSvc() *resume.Service {
	if h.resume == nil {
		h.resume = resume.NewService(h.store, resume.Options{
			WorkingDir: h.opts.WorkingDir,
			GitRepo:    h.gitSvc().Repo(),
		})
	}
	return h.resume
}

// cloneVerdictMap copies locked-pass verdicts from a checkpoint into process state.
// CriterionVerdict is an alias of verify.Verdict, so no field mapping is needed.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func cloneVerdictMap(m map[string]criterionVerdict) map[string]criterionVerdict {
	if len(m) == 0 {
		return map[string]criterionVerdict{}
	}
	out := make(map[string]criterionVerdict, len(m))
	for id, v := range m {
		key := v.ID
		if key == "" {
			key = id
		}
		v.ID = key
		out[key] = v
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func reasonRemediation(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ""
	}
	return "Prior attempt failed: " + reason
}
