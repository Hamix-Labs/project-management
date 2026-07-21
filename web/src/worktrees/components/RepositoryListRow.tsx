import type { MouseEvent } from "react";
import { useState } from "react";
import type { GitRepository } from "@/types/git";
import {
  repositoryDisplayName,
  repositoryPathsEquivalent,
} from "../repositoryDisplay";
import { worktreeGitCopy } from "../worktreeGitCopy";
import {
  WorktreesCheckIcon,
  WorktreesCopyIcon,
  WorktreesFolderGitIcon,
  WorktreesTrashIcon,
} from "./WorktreesIcons";

type Props = {
  repository: GitRepository;
  onDelete: (repository: GitRepository) => void;
};

export function RepositoryListRow({ repository, onDelete }: Props) {
  const [copied, setCopied] = useState(false);
  const repoName = repositoryDisplayName(repository.path);
  const showHostPath =
    repository.host_path.trim() !== "" &&
    !repositoryPathsEquivalent(repository.path, repository.host_path);

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
    <li className="repositories-list-row" role="row" aria-label={repoName}>
      <div className="repositories-list-row__main" role="gridcell">
        <span className="repositories-list-row__icon-wrap" aria-hidden>
          <WorktreesFolderGitIcon className="repositories-list-row__icon" />
        </span>
        <div className="repositories-list-row__details">
          <div className="repositories-list-row__title-row">
            <span className="repositories-list-row__name" title={repoName}>
              {repoName}
            </span>
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
        <div className="repositories-list-row__actions">
          <button
            type="button"
            className="task-list-icon-btn task-list-icon-btn--delete"
            aria-label={`${worktreeGitCopy.deleteRepository} ${repoName}`}
            onClick={() => onDelete(repository)}
          >
            <WorktreesTrashIcon className="repositories-list-row__action-icon" />
          </button>
        </div>
      </div>
    </li>
  );
}
