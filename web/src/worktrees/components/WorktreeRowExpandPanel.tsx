import type { GitBranch } from "@/types/git";
import { worktreeGitCopy } from "../worktreeGitCopy";
import { BranchPill } from "./BranchPill";
import { WorktreesFolderIcon } from "./WorktreesIcons";

type Props = {
  panelId: string;
  path: string;
  branch?: GitBranch;
  branchId?: string;
};

export function WorktreeRowExpandPanel({ panelId, path, branch, branchId }: Props) {
  return (
    <div
      id={panelId}
      className="worktree-row__expand-panel"
      role="region"
      aria-label="Worktree details"
    >
      <dl className="worktree-row__expand-grid">
        <div className="worktree-row__expand-item">
          <dt className="worktree-row__expand-label">{worktreeGitCopy.locationLabel}</dt>
          <dd className="worktree-row__expand-value">
            <WorktreesFolderIcon className="worktree-row__expand-path-icon" aria-hidden />
            <span className="worktree-row__expand-path" title={path}>
              {path}
            </span>
          </dd>
        </div>
        <div className="worktree-row__expand-item">
          <dt className="worktree-row__expand-label">{worktreeGitCopy.listColumnBranch}</dt>
          <dd className="worktree-row__expand-value">
            {branch ? (
              <BranchPill branch={branch} />
            ) : branchId ? (
              <span className="worktree-row__branch-empty">{branchId}</span>
            ) : (
              <span className="worktree-row__branch-empty">{worktreeGitCopy.detachedHead}</span>
            )}
          </dd>
        </div>
      </dl>
    </div>
  );
}
