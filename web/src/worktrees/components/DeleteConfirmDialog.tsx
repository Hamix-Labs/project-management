import { ConfirmDialog } from "@/components/feedback/ConfirmDialog";
import type { GitDeleteTarget } from "../gitDeleteErrors";
import {
  gitDeleteBlocked,
  gitDeleteErrorMessage,
} from "../gitDeleteErrors";

type Props = {
  target: GitDeleteTarget | null;
  pending: boolean;
  error: unknown;
  onClose: () => void;
  onConfirm: () => void;
};

export function DeleteConfirmDialog({
  target,
  pending,
  error,
  onClose,
  onConfirm,
}: Props) {
  if (!target) return null;

  const blocked = error != null && gitDeleteBlocked(error);
  const errorMessage = error != null ? gitDeleteErrorMessage(error) : null;

  return (
    <ConfirmDialog
      title="Delete repository?"
      description={
        <>
          <strong>{target.label}</strong> will be removed from Hamix. Repository
          files on disk are not deleted.
        </>
      }
      footnote="This action cannot be undone."
      confirmLabel={pending ? "Deleting…" : "Delete"}
      confirmVariant="danger"
      busy={pending}
      cancelDisabled={pending}
      confirmDisabled={pending || blocked}
      error={errorMessage}
      onCancel={onClose}
      onConfirm={onConfirm}
      titleId="git-delete-dialog-title"
      descriptionId="git-delete-dialog-description"
      sectionClassName="worktrees-delete-dialog"
      focusCancelOnOpen={false}
    />
  );
}
