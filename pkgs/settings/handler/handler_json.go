package handler

import (
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/repo"
)

// settingsResponse is the on-the-wire shape of GET /settings and the
// PATCH /settings response.
type settingsResponse struct {
	AgentPaused                 bool   `json:"agent_paused"`
	Runner                      string `json:"runner"`
	CursorBin                   string `json:"cursor_bin"`
	CursorModel                 string `json:"cursor_model"`
	VerifyModel                 string `json:"verify_model"`
	MaxRunDurationSeconds       int    `json:"max_run_duration_seconds"`
	StreamIdleStuckSeconds      int    `json:"stream_idle_stuck_seconds"`
	AgentPickupDelaySeconds     int    `json:"agent_pickup_delay_seconds"`
	DisplayTimezone             string `json:"display_timezone"`
	OptimisticMutationsEnabled  bool   `json:"optimistic_mutations_enabled"`
	SSEReplayEnabled            bool   `json:"sse_replay_enabled"`
	VerifyMaxRetries            int    `json:"verify_max_retries"`
	VerifyCommandTimeoutSeconds int    `json:"verify_command_timeout_seconds"`
	CursorSessionResumeEnabled  bool   `json:"cursor_session_resume_enabled"`
	UpdatedAt                   string `json:"updated_at,omitempty"`
}

type settingsPatchBody struct {
	AgentPaused                 *bool   `json:"agent_paused,omitempty"`
	Runner                      *string `json:"runner,omitempty"`
	CursorBin                   *string `json:"cursor_bin,omitempty"`
	CursorModel                 *string `json:"cursor_model,omitempty"`
	VerifyModel                 *string `json:"verify_model,omitempty"`
	MaxRunDurationSeconds       *int    `json:"max_run_duration_seconds,omitempty"`
	StreamIdleStuckSeconds      *int    `json:"stream_idle_stuck_seconds,omitempty"`
	AgentPickupDelaySeconds     *int    `json:"agent_pickup_delay_seconds,omitempty"`
	DisplayTimezone             *string `json:"display_timezone,omitempty"`
	OptimisticMutationsEnabled  *bool   `json:"optimistic_mutations_enabled,omitempty"`
	SSEReplayEnabled            *bool   `json:"sse_replay_enabled,omitempty"`
	VerifyMaxRetries            *int    `json:"verify_max_retries,omitempty"`
	VerifyCommandTimeoutSeconds *int    `json:"verify_command_timeout_seconds,omitempty"`
	CursorSessionResumeEnabled  *bool   `json:"cursor_session_resume_enabled,omitempty"`
}

type probeRequest struct {
	Runner     string `json:"runner,omitempty"`
	BinaryPath string `json:"binary_path,omitempty"`
}

type probeResponse struct {
	OK         bool   `json:"ok"`
	Runner     string `json:"runner"`
	BinaryPath string `json:"binary_path,omitempty"`
	Version    string `json:"version,omitempty"`
	Error      string `json:"error,omitempty"`
}

type cancelRunResponse struct {
	Cancelled bool `json:"cancelled"`
}

type jsonErrorBody struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

type listCursorModelsRequest struct {
	Runner     string `json:"runner,omitempty"`
	BinaryPath string `json:"binary_path,omitempty"`
}

type cursorModelWire struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type listCursorModelsResponse struct {
	OK         bool              `json:"ok"`
	Runner     string            `json:"runner"`
	BinaryPath string            `json:"binary_path,omitempty"`
	Models     []cursorModelWire `json:"models,omitempty"`
	Error      string            `json:"error,omitempty"`
}

type workspaceRootsResponse struct {
	Roots       []repo.BrowseRoot      `json:"roots"`
	Environment repo.BrowseEnvironment `json:"environment"`
}

type browseDirsResponse struct {
	Path       string                `json:"path,omitempty"`
	ParentPath string                `json:"parent_path,omitempty"`
	IsGitRepo  bool                  `json:"is_git_repo,omitempty"`
	Entries    []repo.BrowseDirEntry `json:"entries"`
}

type gitLiveBranchJSON struct {
	Name    string `json:"name"`
	HeadSHA string `json:"head_sha"`
}

type gitRepositoryProbeResponse struct {
	Path            string              `json:"path"`
	MainPath        string              `json:"main_path,omitempty"`
	IsMain          bool                `json:"is_main,omitempty"`
	IsGitRepository bool                `json:"is_git_repository"`
	CurrentBranch   string              `json:"current_branch,omitempty"`
	Branches        []gitLiveBranchJSON `json:"branches"`
}

const settingsProbeTimeout = 5 * time.Second

const maxRepoRelPathQueryBytes = 4096

const maxHTTPLogTitleRunes = 160
