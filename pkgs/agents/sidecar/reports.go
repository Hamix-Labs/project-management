package sidecar

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrCriteriaReportMissing = errors.New("criteria report missing")
	ErrCriteriaReportInvalid = errors.New("criteria report invalid")
	ErrVerifyReportMissing   = errors.New("verify report missing")
	ErrVerifyReportInvalid   = errors.New("verify report invalid")
	ErrSubmitReceiptMissing  = errors.New("submit receipt missing")
	ErrSubmitReceiptInvalid  = errors.New("submit receipt invalid")
)

const maxReportFileBytes = 256 * 1024
const maxFieldBytes = 16 * 1024
const minVerifyReasoning = 40

// CurrentSchemaVersion is the report JSON schema written by the worker and
// expected from agent side-channel files. Major version bumps require parser
// updates; minor fields may be added without a bump.
const CurrentSchemaVersion = 1

const (
	criteriaReportFileName        = "criteria-report.json"
	verifyReportFileName          = "verify-report.json"
	criteriaSubmitReceiptFileName = "criteria-report.submitted"
	verifySubmitReceiptFileName   = "verify-report.submitted"
)

type criteriaReport struct {
	SchemaVersion int             `json:"schema_version"`
	Criteria      []CriteriaEntry `json:"criteria"`
	// Commits is worker-ingested at execute complete; ignored here so
	// ParseCriteriaReport stays compatible with ADR-0014 reports.
	Commits []struct {
		SHA    string `json:"sha"`
		Branch string `json:"branch"`
	} `json:"commits,omitempty"`
}

// CriteriaEntry is one row in criteria-report.json.
type CriteriaEntry struct {
	ID          string `json:"id"`
	ClaimedDone bool   `json:"claimed_done"`
	Evidence    string `json:"evidence"`
}

type verifyReport struct {
	SchemaVersion int           `json:"schema_version"`
	Criteria      []VerifyEntry `json:"criteria"`
}

// VerifyEntry is one row in verify-report.json.
type VerifyEntry struct {
	ID        string `json:"id"`
	Verified  bool   `json:"verified"`
	Reasoning string `json:"reasoning"`
}

// SubmitReceipt is written next to a report after a successful MCP submit.
type SubmitReceipt struct {
	Nonce     string `json:"nonce"`
	Phase     string `json:"phase"`
	CycleID   string `json:"cycle_id"`
	Tool      string `json:"tool"`
	WrittenAt string `json:"written_at,omitempty"`
}

// ReportCycleDir is the worker-managed scratch directory for one
// cycle's report files. Lives under Options.ReportDir (defaulted by
// NewWorker to <os.TempDir()>/hamix-worker) so the operator's RepoRoot
// is never touched.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ReportCycleDir(reportDir, cycleID string) string {
	return filepath.Join(reportDir, cycleID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func CriteriaReportPath(reportDir, cycleID string) string {
	return filepath.Join(ReportCycleDir(reportDir, cycleID), criteriaReportFileName)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func VerifyReportPath(reportDir, cycleID string) string {
	return filepath.Join(ReportCycleDir(reportDir, cycleID), verifyReportFileName)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func CriteriaSubmitReceiptPath(reportDir, cycleID string) string {
	return filepath.Join(ReportCycleDir(reportDir, cycleID), criteriaSubmitReceiptFileName)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func VerifySubmitReceiptPath(reportDir, cycleID string) string {
	return filepath.Join(ReportCycleDir(reportDir, cycleID), verifySubmitReceiptFileName)
}

// EnsureReportCycleDir creates <reportDir>/<cycleID>/ with a permissive
// directory mode so the agent CLI can write its report into it.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func EnsureReportCycleDir(reportDir, cycleID string) error {
	return os.MkdirAll(ReportCycleDir(reportDir, cycleID), 0o755)
}

// ScrubCycleArtifacts removes the per-cycle report subdirectory before
// the next execute attempt writes into it.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ScrubCycleArtifacts(reportDir, cycleID string) error {
	return os.RemoveAll(ReportCycleDir(reportDir, cycleID))
}

// CleanupReportDir removes <reportDir>/<cycleID>/ at cycle terminate time.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func CleanupReportDir(reportDir, cycleID string) error {
	return os.RemoveAll(ReportCycleDir(reportDir, cycleID))
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func readJSONFile(path string, dest any) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrCriteriaReportMissing
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%w: symlink not permitted", ErrCriteriaReportInvalid)
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(io.LimitReader(f, maxReportFileBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		return fmt.Errorf("%w: %v", ErrCriteriaReportInvalid, err)
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func validateSchemaVersion(v int) error {
	if v == 0 {
		return nil
	}
	if v > CurrentSchemaVersion {
		return fmt.Errorf("%w: unsupported schema_version %d", ErrCriteriaReportInvalid, v)
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func validateCriteriaReportSchema(rep *criteriaReport) error {
	return validateSchemaVersion(rep.SchemaVersion)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func validateVerifyReportSchema(rep *verifyReport) error {
	if err := validateSchemaVersion(rep.SchemaVersion); err != nil {
		if errors.Is(err, ErrCriteriaReportInvalid) {
			return ErrVerifyReportInvalid
		}
		return err
	}
	return nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ParseCriteriaReportPartial(reportDir, cycleID string) (map[string]CriteriaEntry, error) {
	path := CriteriaReportPath(reportDir, cycleID)
	var rep criteriaReport
	if err := readJSONFile(path, &rep); err != nil {
		return nil, err
	}
	if err := validateCriteriaReportSchema(&rep); err != nil {
		return nil, err
	}
	out := make(map[string]CriteriaEntry, len(rep.Criteria))
	for _, e := range rep.Criteria {
		id := strings.TrimSpace(e.ID)
		if id == "" {
			continue
		}
		out[id] = e
	}
	return out, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ParseCriteriaReport(reportDir, cycleID string, expectedIDs map[string]struct{}) (map[string]CriteriaEntry, error) {
	path := CriteriaReportPath(reportDir, cycleID)
	var rep criteriaReport
	if err := readJSONFile(path, &rep); err != nil {
		return nil, err
	}
	if err := validateCriteriaReportSchema(&rep); err != nil {
		return nil, err
	}
	out := make(map[string]CriteriaEntry, len(rep.Criteria))
	for _, e := range rep.Criteria {
		id := strings.TrimSpace(e.ID)
		if id == "" {
			return nil, fmt.Errorf("%w: empty criterion id", ErrCriteriaReportInvalid)
		}
		if _, dup := out[id]; dup {
			return nil, fmt.Errorf("%w: duplicate criterion id %s", ErrCriteriaReportInvalid, id)
		}
		if len(e.Evidence) > maxFieldBytes {
			return nil, fmt.Errorf("%w: evidence too long", ErrCriteriaReportInvalid)
		}
		out[id] = e
	}
	for id := range expectedIDs {
		if _, ok := out[id]; !ok {
			return nil, fmt.Errorf("%w: missing criterion %s", ErrCriteriaReportInvalid, id)
		}
	}
	return out, nil
}

// CriteriaCommitClaim is one agent-declared commit in criteria-report.json.
type CriteriaCommitClaim struct {
	SHA    string
	Branch string
}

// ParseCriteriaReportCommits reads commits[] from criteria-report.json for execute ingest.
// Missing report returns nil claims without error.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ParseCriteriaReportCommits(reportDir, cycleID string) ([]CriteriaCommitClaim, error) {
	path := CriteriaReportPath(reportDir, cycleID)
	var rep criteriaReport
	if err := readJSONFile(path, &rep); err != nil {
		if errors.Is(err, ErrCriteriaReportMissing) {
			return nil, nil
		}
		return nil, err
	}
	if err := validateCriteriaReportSchema(&rep); err != nil {
		return nil, err
	}
	if len(rep.Commits) == 0 {
		return nil, nil
	}
	out := make([]CriteriaCommitClaim, 0, len(rep.Commits))
	seen := make(map[string]struct{}, len(rep.Commits))
	for _, c := range rep.Commits {
		sha := strings.TrimSpace(c.SHA)
		if sha == "" {
			return nil, fmt.Errorf("%w: empty commit sha", ErrCriteriaReportInvalid)
		}
		if _, dup := seen[sha]; dup {
			return nil, fmt.Errorf("%w: duplicate commit sha %s", ErrCriteriaReportInvalid, sha)
		}
		seen[sha] = struct{}{}
		out = append(out, CriteriaCommitClaim{
			SHA:    sha,
			Branch: strings.TrimSpace(c.Branch),
		})
	}
	return out, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ParseVerifyReport(reportDir, cycleID string, expectedIDs map[string]struct{}) (map[string]VerifyEntry, error) {
	path := VerifyReportPath(reportDir, cycleID)
	var rep verifyReport
	if err := readJSONFile(path, &rep); err != nil {
		if errors.Is(err, ErrCriteriaReportMissing) {
			return nil, ErrVerifyReportMissing
		}
		return nil, err
	}
	if err := validateVerifyReportSchema(&rep); err != nil {
		return nil, err
	}
	out := make(map[string]VerifyEntry, len(rep.Criteria))
	for _, e := range rep.Criteria {
		id := strings.TrimSpace(e.ID)
		if id == "" {
			return nil, fmt.Errorf("%w: empty criterion id", ErrVerifyReportInvalid)
		}
		if _, dup := out[id]; dup {
			return nil, fmt.Errorf("%w: duplicate criterion id %s", ErrVerifyReportInvalid, id)
		}
		if e.Verified && len(strings.TrimSpace(e.Reasoning)) < minVerifyReasoning {
			return nil, fmt.Errorf("%w: reasoning too short for verified criterion %s", ErrVerifyReportInvalid, id)
		}
		if len(e.Reasoning) > maxFieldBytes {
			return nil, fmt.Errorf("%w: reasoning too long", ErrVerifyReportInvalid)
		}
		out[id] = e
	}
	for id := range expectedIDs {
		if _, ok := out[id]; !ok {
			return nil, fmt.Errorf("%w: missing criterion %s", ErrVerifyReportInvalid, id)
		}
	}
	return out, nil
}

// WriteCriteriaReport atomically writes criteria-report.json for the cycle.
func WriteCriteriaReport(reportDir, cycleID string, criteria []CriteriaEntry, commits []CriteriaCommitClaim) error {
	if err := EnsureReportCycleDir(reportDir, cycleID); err != nil {
		return err
	}
	rep := criteriaReport{
		SchemaVersion: CurrentSchemaVersion,
		Criteria:      criteria,
	}
	if len(commits) > 0 {
		rep.Commits = make([]struct {
			SHA    string `json:"sha"`
			Branch string `json:"branch"`
		}, len(commits))
		for i, c := range commits {
			rep.Commits[i].SHA = c.SHA
			rep.Commits[i].Branch = c.Branch
		}
	}
	return writeJSONAtomic(CriteriaReportPath(reportDir, cycleID), rep)
}

// WriteVerifyReport atomically writes verify-report.json for the cycle.
func WriteVerifyReport(reportDir, cycleID string, criteria []VerifyEntry) error {
	if err := EnsureReportCycleDir(reportDir, cycleID); err != nil {
		return err
	}
	rep := verifyReport{
		SchemaVersion: CurrentSchemaVersion,
		Criteria:      criteria,
	}
	return writeJSONAtomic(VerifyReportPath(reportDir, cycleID), rep)
}

// WriteSubmitReceipt writes the MCP submit receipt next to the report.
func WriteSubmitReceipt(path string, receipt SubmitReceipt) error {
	return writeJSONAtomic(path, receipt)
}

// RequireCriteriaSubmitReceipt ensures the criteria receipt exists and matches nonce.
func RequireCriteriaSubmitReceipt(reportDir, cycleID, nonce string) error {
	return requireSubmitReceipt(CriteriaSubmitReceiptPath(reportDir, cycleID), nonce)
}

// RequireVerifySubmitReceipt ensures the verify receipt exists and matches nonce.
func RequireVerifySubmitReceipt(reportDir, cycleID, nonce string) error {
	return requireSubmitReceipt(VerifySubmitReceiptPath(reportDir, cycleID), nonce)
}

func requireSubmitReceipt(path, nonce string) error {
	var rec SubmitReceipt
	if err := readJSONFile(path, &rec); err != nil {
		if errors.Is(err, ErrCriteriaReportMissing) {
			return ErrSubmitReceiptMissing
		}
		if errors.Is(err, ErrCriteriaReportInvalid) {
			return fmt.Errorf("%w: %v", ErrSubmitReceiptInvalid, err)
		}
		return err
	}
	if strings.TrimSpace(rec.Nonce) == "" || rec.Nonce != nonce {
		return fmt.Errorf("%w: nonce mismatch", ErrSubmitReceiptInvalid)
	}
	return nil
}

func writeJSONAtomic(path string, v any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".hamix-sidecar-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	enc := json.NewEncoder(tmp)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

// MinVerifyReasoningChars is the minimum reasoning length when verified=true.
func MinVerifyReasoningChars() int { return minVerifyReasoning }

// MaxFieldBytes is the max evidence/reasoning field size.
func MaxFieldBytes() int { return maxFieldBytes }
