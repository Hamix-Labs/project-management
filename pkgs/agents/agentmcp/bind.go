package agentmcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const BindSchemaVersion = 1

const (
	PhaseExecute = "execute"
	PhaseVerify  = "verify"
)

// BindFile is agent-tool-bind.json written by the harness before a runner.Run.
type BindFile struct {
	BindSchemaVersion  int      `json:"bind_schema_version"`
	TaskID             string   `json:"task_id"`
	CycleID            string   `json:"cycle_id"`
	Phase              string   `json:"phase"`
	ReportDir          string   `json:"report_dir"` // parent Options.ReportDir (not cycle subdir)
	WorkingDir         string   `json:"working_dir"`
	ActiveCriterionIDs []string `json:"active_criterion_ids"`
	LockedCriterionIDs []string `json:"locked_criterion_ids"`
	SubmitNonce        string   `json:"submit_nonce"`
}

// Session is the in-process view of a bind file for tool handlers.
type Session struct {
	TaskID             string
	CycleID            string
	Phase              string
	ReportDir          string
	WorkingDir         string
	ActiveCriterionIDs map[string]struct{}
	LockedCriterionIDs map[string]struct{}
	SubmitNonce        string
}

// LoadBind reads and validates a bind file from path.
func LoadBind(path string) (*Session, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("bind path is empty")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read bind: %w", err)
	}
	var b BindFile
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("parse bind: %w", err)
	}
	return sessionFromBind(b)
}

func sessionFromBind(b BindFile) (*Session, error) {
	if b.BindSchemaVersion != BindSchemaVersion {
		return nil, fmt.Errorf("unsupported bind_schema_version %d (want %d)", b.BindSchemaVersion, BindSchemaVersion)
	}
	phase := strings.TrimSpace(b.Phase)
	if phase != PhaseExecute && phase != PhaseVerify {
		return nil, fmt.Errorf("invalid phase %q", b.Phase)
	}
	if strings.TrimSpace(b.TaskID) == "" || strings.TrimSpace(b.CycleID) == "" {
		return nil, fmt.Errorf("task_id and cycle_id are required")
	}
	if strings.TrimSpace(b.ReportDir) == "" {
		return nil, fmt.Errorf("report_dir is required")
	}
	if strings.TrimSpace(b.SubmitNonce) == "" {
		return nil, fmt.Errorf("submit_nonce is required")
	}
	active := make(map[string]struct{}, len(b.ActiveCriterionIDs))
	for _, id := range b.ActiveCriterionIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			active[id] = struct{}{}
		}
	}
	locked := make(map[string]struct{}, len(b.LockedCriterionIDs))
	for _, id := range b.LockedCriterionIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			locked[id] = struct{}{}
		}
	}
	return &Session{
		TaskID:             strings.TrimSpace(b.TaskID),
		CycleID:            strings.TrimSpace(b.CycleID),
		Phase:              phase,
		ReportDir:          strings.TrimSpace(b.ReportDir),
		WorkingDir:         strings.TrimSpace(b.WorkingDir),
		ActiveCriterionIDs: active,
		LockedCriterionIDs: locked,
		SubmitNonce:        strings.TrimSpace(b.SubmitNonce),
	}, nil
}

// WriteBind writes bind JSON for harness callers.
func WriteBind(path string, b BindFile) error {
	b.BindSchemaVersion = BindSchemaVersion
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}

// BindPath returns the standard bind file path under the cycle report dir.
func BindPath(reportDir, cycleID string) string {
	return filepath.Join(reportDir, cycleID, "agent-tool-bind.json")
}
