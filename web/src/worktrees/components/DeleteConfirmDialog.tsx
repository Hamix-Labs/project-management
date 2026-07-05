import { useEffect, useState, type ReactNode } from "react";
import { ConfirmDialog } from "@/components/feedback/ConfirmDialog";
import type { GitDeleteTarget } from "../gitDeleteErrors";
import {
  gitDeleteBlocked,
  gitDeleteErrorMessage,
  gitDeleteNeedsForce,
} from "../gitDeleteErrors";
import { worktreeGitCopy } from "../worktreeGitCopy";

type Props = {
  target: GitDeleteTarget | null;
  pending: boolean;
  error: unknown;
  onClose: () => void;
  onConfirm: (options?: { force?: boolean }) => void;
};

function targetNoun(kind: GitDeleteTarget["kind"]): string {
  switch (kind) {
    case "repository":
      return "repository";
    case "worktree":
      return "worktree";
  }
}

function dialogTitle(target: GitDeleteTarget): string {
  if (target.kind === "worktree") {
    return target.mode === "remove_from_disk"
      ? worktreeGitCopy.deleteWorktreeConfirmTitle
      : worktreeGitCopy.unregisterWorktreeConfirmTitle;
  }
  return `Delete ${targetNoun(target.kind)}?`;
}

function dialogDescription(
  target: GitDeleteTarget,
  showForceOption: boolean,
  forceRemove: boolean,
  pending: boolean,
  onForceChange: (checked: boolean) => void,
): ReactNode {
  if (target.kind === "worktree" && target.mode === "remove_from_disk") {
    return (
      <>
        <strong>{target.label}</strong> {worktreeGitCopy.deleteWorktreeConfirmDescription}
        {showForceOption ? (
          <label className="worktrees-form-modal__checkbox worktrees-delete-dialog__force">
            <input
              type="checkbox"
              checked={forceRemove}
              disabled={pending}
              onChange={(e) => onForceChange(e.target.checked)}
            />
            {worktreeGitCopy.deleteWorktreeForceLabel}
          </label>
        ) : null}
      </>
    );
  }
  if (target.kind === "worktree") {
    return (
      <>
        <strong>{target.label}</strong> {worktreeGitCopy.unregisterWorktreeConfirmDescription}
      </>
    );
  }
  return (
    <>
      <strong>{target.label}</strong> will be removed from Hamix. Repository files on disk are not
      deleted.
    </>
  );
}

function confirmLabel(target: GitDeleteTarget, pending: boolean): string {
  if (target.kind === "worktree") {
    if (target.mode === "remove_from_disk") {
      return pending ? "Deleting…" : worktreeGitCopy.deleteWorktree;
    }
    return pending ? "Unregistering…" : "Unregister";
  }
  return pending ? "Deleting…" : "Delete";
}

export function DeleteConfirmDialog({
  target,
  pending,
  error,
  onClose,
  onConfirm,
}: Props) {
  const [forceRemove, setForceRemove] = useState(false);

  useEffect(() => {
    if (target == null) {
      setForceRemove(false);
    }
  }, [target]);

  if (!target) return null;

  const blocked = error != null && gitDeleteBlocked(error);
  const needsForce = error != null && gitDeleteNeedsForce(error);
  const errorMessage = error != null ? gitDeleteErrorMessage(error) : null;
  const showForceOption =
    target.kind === "worktree" && target.mode === "remove_from_disk" && needsForce;

  return (
    <ConfirmDialog
      title={dialogTitle(target)}
      description={dialogDescription(
        target,
        showForceOption,
        forceRemove,
        pending,
        setForceRemove,
      )}
      footnote={
        target.kind === "worktree" && target.mode === "remove_from_disk"
          ? worktreeGitCopy.deleteWorktreeConfirmFootnote
          : target.kind === "worktree"
            ? worktreeGitCopy.unregisterWorktreeConfirmFootnote
            : "This action cannot be undone."
      }
      confirmLabel={confirmLabel(target, pending)}
      confirmVariant="danger"
      busy={pending}
      cancelDisabled={pending}
      confirmDisabled={pending || blocked || (needsForce && !forceRemove)}
      error={errorMessage}
      onCancel={onClose}
      onConfirm={() =>
        onConfirm(
          target.kind === "worktree" && target.mode === "remove_from_disk" && forceRemove
            ? { force: true }
            : undefined,
        )
      }
      titleId="git-delete-dialog-title"
      descriptionId="git-delete-dialog-description"
      sectionClassName="worktrees-delete-dialog"
      focusCancelOnOpen={false}
    />
  );
}
