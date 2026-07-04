import { useId, useState } from "react";
import type { GitBranch, GitWorktree } from "@/types/git";
import { worktreeAriaLabel, worktreeGitCopy } from "../worktreeGitCopy";
import { BranchPill } from "./BranchPill";
import {
  WorktreesChevronRightIcon,
  WorktreesFolderIcon,
  WorktreesMoreIcon,
} from "./WorktreesIcons";
import { WorktreesMenu } from "./WorktreesMenu";

type Props = {
  worktree: GitWorktree;
  branches: GitBranch[];
  onUnregister: () => void;
  onDeleteFromDisk: () => void;
  deleteDisabled?: boolean;
};

function isWorktreeRowActionExcluded(target: EventTarget | null): boolean {
  if (!(target instanceof Element)) return true;
  return Boolean(target.closest("button, [role='menu'], [role='menuitem']"));
}

export function WorktreeRow({
  worktree,
  branches,
  onUnregister,
  onDeleteFromDisk,
  deleteDisabled = false,
}: Props) {
  const [expanded, setExpanded] = useState(false);
  const pathPanelId = useId();
  const displayName = worktree.name.trim() || worktree.path;
  const branchById = new Map(branches.map((b) => [b.id, b]));
  const branch = worktree.branch_id ? branchById.get(worktree.branch_id) : undefined;
  const deleteBlocked = deleteDisabled;
  const kindLabel = worktree.is_main ? worktreeGitCopy.mainWorktreeShortLabel : null;

  const unregisterMenuItem = {
    id: "unregister-worktree",
    label: worktreeGitCopy.unregisterWorktree,
    onSelect: onUnregister,
    disabled: deleteBlocked,
    danger: true,
  };
  const deleteMenuItem = {
    id: "delete-worktree",
    label: worktreeGitCopy.deleteWorktree,
    onSelect: onDeleteFromDisk,
    disabled: deleteBlocked,
    danger: true,
  };
  const menuItems = worktree.is_main
    ? [unregisterMenuItem]
    : [unregisterMenuItem, deleteMenuItem];

  const toggleExpanded = () => setExpanded((open) => !open);

  return (
    <li
      className={[
        "worktrees-list-row",
        "worktrees-list-row--worktree",
        "draft-row--interactive",
        expanded ? "worktrees-list-row--expanded" : "",
      ]
        .filter(Boolean)
        .join(" ")}
      data-main={worktree.is_main ? "true" : "false"}
      aria-label={worktreeAriaLabel(displayName)}
      aria-expanded={expanded}
      aria-controls={pathPanelId}
      tabIndex={0}
      onClick={(event) => {
        if (isWorktreeRowActionExcluded(event.target)) return;
        toggleExpanded();
      }}
      onKeyDown={(event) => {
        if (isWorktreeRowActionExcluded(event.target)) return;
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          toggleExpanded();
        }
      }}
    >
      <div className="worktrees-list-row__name worktree-row__name">
        <div className="worktree-row__summary">
          <span className="worktree-row__chevron-wrap" aria-hidden>
            <WorktreesChevronRightIcon className="worktree-row__chevron" />
          </span>
          <span className="worktree-row__label" title={displayName}>
            {displayName}
          </span>
          {kindLabel ? (
            <span className="worktree-row__kind" title={worktreeGitCopy.mainWorktreeHint}>
              {kindLabel}
            </span>
          ) : null}
        </div>
        {expanded ? (
          <p id={pathPanelId} className="worktree-row__path" title={worktree.path}>
            <WorktreesFolderIcon className="worktree-row__path-icon" aria-hidden />
            <span className="worktree-row__path-text">{worktree.path}</span>
          </p>
        ) : null}
      </div>

      <div className="worktrees-list-row__branch worktree-row__branch" aria-label="Branch">
        {branch ? (
          <BranchPill branch={branch} />
        ) : worktree.branch_id ? (
          <span className="worktree-row__branch-empty">{worktree.branch_id}</span>
        ) : (
          <span className="worktree-row__branch-empty">{worktreeGitCopy.detachedHead}</span>
        )}
      </div>

      <div className="worktrees-list-row__actions worktree-row__actions">
        <div className="task-list-row-actions">
          <WorktreesMenu
            triggerLabel={worktreeGitCopy.worktreeActions(displayName)}
            className="task-list-icon-btn"
            icon={<WorktreesMoreIcon />}
            iconOnly
            items={menuItems}
          />
        </div>
      </div>
    </li>
  );
}
