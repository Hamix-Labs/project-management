import type { KeyboardEvent, MouseEvent } from "react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import type { GitRepository } from "@/types/git";
import {
  repositoryDisplayName,
  repositoryPathsEquivalent,
} from "../repositoryDisplay";
import {
  worktreeGitCopy,
  repositoryListWorktreeCountDisplay,
  worktreeCountLabel,
} from "../worktreeGitCopy";
import {
  WorktreesBranchIcon,
  WorktreesCheckIcon,
  WorktreesChevronRightIcon,
  WorktreesCopyIcon,
  WorktreesFolderGitIcon,
} from "./WorktreesIcons";

type Props = {
  repository: GitRepository;
};

function isRowActionExcluded(target: EventTarget | null): boolean {
  if (!(target instanceof Element)) return true;
  return Boolean(target.closest("button, a, input, label"));
}

export function RepositoryListRow({ repository }: Props) {
  const navigate = useNavigate();
  const [copied, setCopied] = useState(false);
  const repoName = repositoryDisplayName(repository.path);
  const showHostPath =
    repository.host_path.trim() !== "" &&
    !repositoryPathsEquivalent(repository.path, repository.host_path);
  const branchName = repository.main_branch_name.trim();
  const worktreeCount = repository.linked_worktree_count;
  const worktreeCountText = worktreeCountLabel(worktreeCount);

  const openDetail = () => {
    navigate(`/worktrees/${repository.id}`);
  };

  const onKeyDown = (event: KeyboardEvent<HTMLLIElement>) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      openDetail();
    }
  };

  const onCopyPath = async (event: MouseEvent<HTMLButtonElement>) => {
    event.stopPropagation();
    try {
      await navigator.clipboard.writeText(repository.path);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // clipboard unavailable
    }
  };

  return (
    <li
      className="repositories-list-row"
      role="row"
      tabIndex={0}
      aria-label={`${repoName}, ${worktreeCountText}`}
      onClick={(event) => {
        if (isRowActionExcluded(event.target)) return;
        openDetail();
      }}
      onKeyDown={onKeyDown}
    >
      <div className="repositories-list-row__main" role="gridcell">
        <span className="repositories-list-row__icon-wrap" aria-hidden>
          <WorktreesFolderGitIcon className="repositories-list-row__icon" />
        </span>
        <div className="repositories-list-row__details">
          <div className="repositories-list-row__title-row">
            <span className="repositories-list-row__name" title={repoName}>
              {repoName}
            </span>
            {branchName ? (
              <span className="repositories-list-row__branch">
                <WorktreesBranchIcon className="repositories-list-row__branch-icon" />
                {branchName}
              </span>
            ) : null}
          </div>
          <div className="repositories-list-row__path-row">
            <code className="repositories-list-row__path" title={repository.path}>
              {repository.path}
            </code>
            <button
              type="button"
              className="repositories-list-row__copy"
              aria-label={
                copied
                  ? worktreeGitCopy.pathCopied
                  : `${worktreeGitCopy.copyPath} for ${repoName}`
              }
              onClick={onCopyPath}
            >
              {copied ? (
                <WorktreesCheckIcon className="repositories-list-row__copy-icon repositories-list-row__copy-icon--success" />
              ) : (
                <WorktreesCopyIcon className="repositories-list-row__copy-icon" />
              )}
            </button>
          </div>
          {showHostPath ? (
            <span className="repositories-list-row__host-path">
              <span className="worktrees-repo-row__meta-label">
                {worktreeGitCopy.hostPathLabel}
              </span>
              <code>{repository.host_path}</code>
            </span>
          ) : null}
        </div>
      </div>
      <div className="repositories-list-row__aside" role="gridcell">
        <span className="repositories-list-row__count-pill" aria-label={worktreeCountText}>
          {repositoryListWorktreeCountDisplay(worktreeCount)}
        </span>
        <span className="repositories-list-row__chevron" aria-hidden="true">
          <WorktreesChevronRightIcon />
        </span>
      </div>
    </li>
  );
}
