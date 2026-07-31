package sidecar

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const commitRegisterFileName = "commit-register.json"

var (
	// ErrCommitRegisterInvalid is returned when commit-register.json is corrupt or violates schema.
	ErrCommitRegisterInvalid = errors.New("commit register invalid")
	// ErrCommitRegisterDuplicate is returned when appending a SHA already in the register.
	ErrCommitRegisterDuplicate = errors.New("commit register duplicate sha")
)

// CommitRegisterEntry is one MCP-recorded commit for a cycle execute visit.
type CommitRegisterEntry struct {
	SHA       string `json:"sha"`
	Message   string `json:"message,omitempty"`
	Branch    string `json:"branch,omitempty"`
	WrittenAt string `json:"written_at,omitempty"`
}

type commitRegisterFile struct {
	SchemaVersion int                   `json:"schema_version"`
	Commits       []CommitRegisterEntry `json:"commits"`
}

// CommitRegisterPath is ReportDir/<cycleID>/commit-register.json.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func CommitRegisterPath(reportDir, cycleID string) string {
	return filepath.Join(ReportCycleDir(reportDir, cycleID), commitRegisterFileName)
}

// ParseCommitRegister reads the cycle commit register.
// Missing file returns an empty slice without error.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ParseCommitRegister(reportDir, cycleID string) ([]CommitRegisterEntry, error) {
	path := CommitRegisterPath(reportDir, cycleID)
	var reg commitRegisterFile
	if err := readJSONFile(path, &reg); err != nil {
		if errors.Is(err, ErrCriteriaReportMissing) {
			return nil, nil
		}
		if errors.Is(err, ErrCriteriaReportInvalid) {
			return nil, fmt.Errorf("%w: %v", ErrCommitRegisterInvalid, err)
		}
		return nil, err
	}
	if err := validateSchemaVersion(reg.SchemaVersion); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCommitRegisterInvalid, err)
	}
	if reg.Commits == nil {
		return nil, nil
	}
	out := make([]CommitRegisterEntry, 0, len(reg.Commits))
	seen := make(map[string]struct{}, len(reg.Commits))
	for _, c := range reg.Commits {
		sha := strings.TrimSpace(c.SHA)
		if sha == "" {
			return nil, fmt.Errorf("%w: empty commit sha", ErrCommitRegisterInvalid)
		}
		if _, dup := seen[sha]; dup {
			return nil, fmt.Errorf("%w: duplicate commit sha %s", ErrCommitRegisterInvalid, sha)
		}
		seen[sha] = struct{}{}
		out = append(out, CommitRegisterEntry{
			SHA:       sha,
			Message:   strings.TrimSpace(c.Message),
			Branch:    strings.TrimSpace(c.Branch),
			WrittenAt: strings.TrimSpace(c.WrittenAt),
		})
	}
	return out, nil
}

// AppendCommitRegister appends one commit to the cycle register (atomic rewrite).
// Rejects duplicate SHAs.
//
//funclogmeasure:skip category=hot-path reason="Atomic file rewrite; operation trace is emitted by the MCP/harness caller."
func AppendCommitRegister(reportDir, cycleID string, entry CommitRegisterEntry) error {
	sha := strings.TrimSpace(entry.SHA)
	if sha == "" {
		return fmt.Errorf("%w: empty commit sha", ErrCommitRegisterInvalid)
	}
	if len(entry.Message) > maxFieldBytes {
		return fmt.Errorf("%w: message too long", ErrCommitRegisterInvalid)
	}
	if err := EnsureReportCycleDir(reportDir, cycleID); err != nil {
		return err
	}
	existing, err := ParseCommitRegister(reportDir, cycleID)
	if err != nil {
		return err
	}
	for _, e := range existing {
		if e.SHA == sha {
			return fmt.Errorf("%w: %s", ErrCommitRegisterDuplicate, sha)
		}
	}
	writtenAt := strings.TrimSpace(entry.WrittenAt)
	if writtenAt == "" {
		writtenAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	next := append(append([]CommitRegisterEntry{}, existing...), CommitRegisterEntry{
		SHA:       sha,
		Message:   strings.TrimSpace(entry.Message),
		Branch:    strings.TrimSpace(entry.Branch),
		WrittenAt: writtenAt,
	})
	reg := commitRegisterFile{
		SchemaVersion: CurrentSchemaVersion,
		Commits:       next,
	}
	return writeJSONAtomic(CommitRegisterPath(reportDir, cycleID), reg)
}
