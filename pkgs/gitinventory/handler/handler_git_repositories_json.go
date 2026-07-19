package handler

type gitRepositoryCreateJSON struct {
	Path          string `json:"path"`
	HostPath      string `json:"host_path"`
	DefaultBranch string `json:"default_branch"`
}

type gitRepositoriesListResponse struct {
	Repositories []gitRepositoryJSON `json:"repositories"`
}

type gitRepositoryJSON struct {
	ID                  string `json:"id"`
	Path                string `json:"path"`
	GitCommonDir        string `json:"git_common_dir"`
	HostPath            string `json:"host_path"`
	DefaultBranch       string `json:"default_branch"`
	MainBranchName      string `json:"main_branch_name"`
	LinkedWorktreeCount int    `json:"linked_worktree_count"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type gitWorktreesListResponse struct {
	Worktrees []gitWorktreeJSON `json:"worktrees"`
}

type gitWorktreeJSON struct {
	ID           string `json:"id"`
	RepositoryID string `json:"repository_id"`
	Path         string `json:"path"`
	HostPath     string `json:"host_path"`
	Name         string `json:"name"`
	IsMain       bool   `json:"is_main"`
	BranchID     string `json:"branch_id"`
	CreatedAt    string `json:"created_at"`
	Stale        bool   `json:"stale,omitempty"`
}

type gitBranchesListResponse struct {
	Branches []gitBranchJSON `json:"branches"`
}

type gitBranchJSON struct {
	ID           string `json:"id"`
	RepositoryID string `json:"repository_id"`
	Name         string `json:"name"`
	HeadSHA      string `json:"head_sha"`
	CreatedAt    string `json:"created_at"`
}

type gitWorktreeCheckoutStatusJSON struct {
	WorktreeID   string `json:"worktree_id"`
	Available    bool   `json:"available"`
	Reason       string `json:"reason,omitempty"`
	Dirty        bool   `json:"dirty,omitempty"`
	Detached     bool   `json:"detached,omitempty"`
	HeadCommitAt string `json:"head_commit_at,omitempty"`
	HasUpstream  bool   `json:"has_upstream,omitempty"`
	Ahead        *int   `json:"ahead,omitempty"`
	Behind       *int   `json:"behind,omitempty"`
	Upstream     string `json:"upstream,omitempty"`
}

type gitWorktreeCheckoutStatusListResponse struct {
	Worktrees []gitWorktreeCheckoutStatusJSON `json:"worktrees"`
}
