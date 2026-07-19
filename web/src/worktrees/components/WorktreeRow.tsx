import { useId, useState } from "react";
import type { GitBranch, GitWorktree, GitWorktreeCheckoutStatus } from "@/types/git";
import { worktreeAriaLabel, worktreeGitCopy } from "../worktreeGitCopy";
import { BranchPill } from "./BranchPill";
import { BranchSyncIndicator } from "./BranchSyncIndicator";
import { WorktreeRowExpandPanel } from "./WorktreeRowExpandPanel";
import { WorktreeRowStatus } from "./WorktreeRowStatus";
import { WorktreesChevronRightIcon, WorktreesMoreIcon } from "./WorktreesIcons";
import { WorktreesMenu } from "./WorktreesMenu";

type Props = {
  worktree: GitWorktree;
  branches: GitBranch[];
  checkoutStatus?: GitWorktreeCheckoutStatus;
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
  checkoutStatus,
  onUnregister,
  onDeleteFromDisk,
  deleteDisabled = false,
}: Props) {
  const [expanded, setExpanded] = useState(false);
  const panelId = useId();
  const displayName = worktree.name.trim() || worktree.path;
  const branchById = new Map(branches.map((b) => [b.id, b]));
  const branch = worktree.branch_id ? branchById.get(worktree.branch_id) : undefined;
  const deleteBlocked = deleteDisabled;

  const unregisterMenuItem = {
    id: "unregister-worktree",
    label: worktreeGitCopy.unregisterWorktree,
    onSelect: onUnregister,
    disabled: deleteBlocked,
    danger: true,
  };
  const deleteMenuItem = {
    id: "delete-worktree",
    label: worktree.stale
      ? worktreeGitCopy.removeStaleWorktree
      : worktreeGitCopy.deleteWorktree,
    onSelect: onDeleteFromDisk,
    disabled: deleteBlocked,
    danger: true,
  };
  // Primary checkout is never listed; if it appears, expose no destructive actions.
  const menuItems = worktree.is_main
    ? []
    : worktree.stale
      ? [deleteMenuItem]
      : [unregisterMenuItem, deleteMenuItem];

  const toggleExpanded = () => setExpanded((open) => !open);

  return (
    <li
      className={[
        "worktree-row",
        expanded ? "worktree-row--expanded" : "",
      ]
        .filter(Boolean)
        .join(" ")}
      data-main={worktree.is_main ? "true" : "false"}
      aria-label={worktreeAriaLabel(displayName)}
    >
      <div
        className="worktree-row__main"
        onClick={(event) => {
          if (isWorktreeRowActionExcluded(event.target)) return;
          toggleExpanded();
        }}
      >
        <button
          type="button"
          className="worktree-row__chevron-btn"
          aria-expanded={expanded}
          aria-controls={panelId}
          aria-label={expanded ? `Collapse ${displayName}` : `Expand ${displayName}`}
          onClick={(event) => {
            event.stopPropagation();
            toggleExpanded();
          }}
        >
          <WorktreesChevronRightIcon className="worktree-row__chevron" aria-hidden />
        </button>

        <div className="worktree-row__content">
          <div className="worktree-row__title-row">
            <span className="worktree-row__label" title={displayName}>
              {displayName}
            </span>
            {worktree.is_main ? (
              <span className="worktree-row__primary-badge" title={worktreeGitCopy.mainWorktreeHint}>
                {worktreeGitCopy.primaryWorktreeBadge}
              </span>
            ) : null}
            {worktree.stale ? (
              <span className="worktree-row__stale-hint" title={worktreeGitCopy.staleWorktreeHint}>
                {worktreeGitCopy.staleWorktreeHint}
              </span>
            ) : null}
          </div>
          <WorktreeRowStatus checkoutStatus={checkoutStatus} />
        </div>

        <div className="worktree-row__branch" aria-label="Branch">
          <BranchSyncIndicator checkoutStatus={checkoutStatus} />
          {branch ? (
            <BranchPill branch={branch} />
          ) : worktree.branch_id ? (
            <span className="worktree-row__branch-empty">{worktree.branch_id}</span>
          ) : (
            <span className="worktree-row__branch-empty">{worktreeGitCopy.detachedHead}</span>
          )}
        </div>

        <div className="worktree-row__actions">
          {menuItems.length > 0 ? (
            <WorktreesMenu
              triggerLabel={worktreeGitCopy.worktreeActions(displayName)}
              className="worktree-row__menu-btn"
              icon={<WorktreesMoreIcon />}
              iconOnly
              items={menuItems}
            />
          ) : null}
        </div>
      </div>

      {expanded ? (
        <WorktreeRowExpandPanel
          panelId={panelId}
          path={worktree.path}
          branch={branch}
          branchId={worktree.branch_id}
        />
      ) : null}
    </li>
  );
}
