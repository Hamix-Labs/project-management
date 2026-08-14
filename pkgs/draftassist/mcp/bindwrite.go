package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// WriteBind writes a draft-assist bind JSON file at path. The schema version
// is stamped to the current BindSchemaVersion regardless of the caller's
// value; callers should treat this as the sole way to emit bind files so
// the on-disk contract stays consistent for hamix-draft-mcp and the future
// SDK sidecar (Plan 5).
//
//funclogmeasure:skip category=hot-path reason="Pure I/O helper; harness call site emits the operation trace."
func WriteBind(path string, b BindFile) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("bind path is empty")
	}
	if strings.TrimSpace(b.SessionID) == "" || strings.TrimSpace(b.Nonce) == "" {
		return fmt.Errorf("session_id and nonce are required")
	}
	b.BindSchemaVersion = BindSchemaVersion
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal bind: %w", err)
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}
