package domain

import (
	"fmt"
	"strings"
)

// RetryMode selects operator retry behavior after a terminal task failure.
type RetryMode string

const (
	RetryFresh  RetryMode = "fresh"
	RetryResume RetryMode = "resume"
)

// PendingRunKind distinguishes failure-retry, polish, and open-pr in pending_retry JSON.
type PendingRunKind string

const (
	PendingKindRetry  PendingRunKind = "retry"
	PendingKindPolish PendingRunKind = "polish"
	PendingKindOpenPR PendingRunKind = "open_pr"
)

// PendingRetry is ephemeral intent set by POST /tasks/{id}/retry or
// POST /tasks/{id}/polish and consumed when the worker transitions the task
// from ready to running. Column name remains pending_retry.
type PendingRetry struct {
	Kind                PendingRunKind `json:"kind,omitempty"`
	Mode                RetryMode      `json:"mode"`
	ParentCycleID       string         `json:"parent_cycle_id"`
	Instructions        string         `json:"instructions,omitempty"`
	FlaggedCriterionIDs []string       `json:"flagged_criterion_ids,omitempty"`
	NewCriterionIDs     []string       `json:"new_criterion_ids,omitempty"`
	SkipVerify          bool           `json:"skip_verify,omitempty"`
}

// NormalizeKind returns the effective kind (empty/omitted => retry).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p *PendingRetry) NormalizeKind() PendingRunKind {
	if p == nil {
		return PendingKindRetry
	}
	k := PendingRunKind(strings.TrimSpace(string(p.Kind)))
	if k == "" {
		return PendingKindRetry
	}
	return k
}

// Validate normalizes and checks a pending retry/polish payload.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p *PendingRetry) Validate() error {
	if p == nil {
		return fmt.Errorf("%w: pending retry", ErrInvalidInput)
	}
	kind := p.NormalizeKind()
	switch kind {
	case PendingKindRetry, PendingKindPolish, PendingKindOpenPR:
	default:
		return fmt.Errorf("%w: pending kind", ErrInvalidInput)
	}
	p.Kind = kind
	switch p.Mode {
	case RetryFresh, RetryResume:
	default:
		return fmt.Errorf("%w: retry mode", ErrInvalidInput)
	}
	parent := strings.TrimSpace(p.ParentCycleID)
	if parent == "" {
		return fmt.Errorf("%w: parent_cycle_id", ErrInvalidInput)
	}
	p.ParentCycleID = parent
	instructions := strings.TrimSpace(p.Instructions)
	switch kind {
	case PendingKindPolish:
		if p.Mode != RetryResume {
			return fmt.Errorf("%w: polish requires mode resume", ErrInvalidInput)
		}
		if instructions == "" {
			return fmt.Errorf("%w: polish instructions", ErrInvalidInput)
		}
		p.Instructions = instructions
		p.FlaggedCriterionIDs = normalizeIDList(p.FlaggedCriterionIDs)
		p.NewCriterionIDs = normalizeIDList(p.NewCriterionIDs)
		p.SkipVerify = len(p.FlaggedCriterionIDs) == 0 && len(p.NewCriterionIDs) == 0
	case PendingKindOpenPR:
		if p.Mode != RetryResume {
			return fmt.Errorf("%w: open_pr requires mode resume", ErrInvalidInput)
		}
		p.Instructions = ""
		p.FlaggedCriterionIDs = nil
		p.NewCriterionIDs = nil
		p.SkipVerify = true
	default:
		p.Instructions = ""
		p.FlaggedCriterionIDs = nil
		p.NewCriterionIDs = nil
		p.SkipVerify = false
	}
	return nil
}

// Clone returns a shallow copy for consumption after the row is cleared.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p *PendingRetry) Clone() *PendingRetry {
	if p == nil {
		return nil
	}
	cp := *p
	if p.FlaggedCriterionIDs != nil {
		cp.FlaggedCriterionIDs = append([]string(nil), p.FlaggedCriterionIDs...)
	}
	if p.NewCriterionIDs != nil {
		cp.NewCriterionIDs = append([]string(nil), p.NewCriterionIDs...)
	}
	return &cp
}

// Equal reports whether two pending payloads match for idempotency.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p *PendingRetry) Equal(other *PendingRetry) bool {
	if p == nil || other == nil {
		return p == nil && other == nil
	}
	return p.NormalizeKind() == other.NormalizeKind() &&
		p.Mode == other.Mode &&
		p.ParentCycleID == other.ParentCycleID &&
		strings.TrimSpace(p.Instructions) == strings.TrimSpace(other.Instructions) &&
		stringSlicesEqual(p.FlaggedCriterionIDs, other.FlaggedCriterionIDs) &&
		stringSlicesEqual(p.NewCriterionIDs, other.NewCriterionIDs) &&
		p.SkipVerify == other.SkipVerify
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func normalizeIDList(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
