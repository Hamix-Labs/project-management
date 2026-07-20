import { useState } from "react";
import { Button } from "@/components/ui";
import { Modal } from "@/shared/Modal";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";
import { WorkspaceDirPickerModal } from "@/components/workspace-picker";
import { gitDeleteErrorMessage } from "../gitDeleteErrors";

type Props = {
  open: boolean;
  pending: boolean;
  error: unknown;
  onClose: () => void;
  onSubmit: (input: { path: string }) => void;
};

export function RegisterRepositoryModal({
  open,
  pending,
  error,
  onClose,
  onSubmit,
}: Props) {
  const [path, setPath] = useState("");
  const [pickerOpen, setPickerOpen] = useState(false);

  if (!open) return null;

  const errorMessage = error != null ? gitDeleteErrorMessage(error) : null;

  return (
    <>
      <Modal
        onClose={onClose}
        labelledBy="register-repo-title"
        describedBy="register-repo-lead"
        busy={pending}
        dismissibleWhileBusy={false}
      >
        <form
          className="panel modal-sheet worktrees-form-modal"
          onSubmit={(e) => {
            e.preventDefault();
            const trimmed = path.trim();
            if (!trimmed) return;
            onSubmit({ path: trimmed });
          }}
        >
          <header className="worktrees-form-modal__header">
            <h2 id="register-repo-title">Register repository</h2>
            <p id="register-repo-lead" className="worktrees-form-modal__lead">
              Point Hamix at a git repository on disk. Task workspaces are allocated later when you
              create a task on this repository.
            </p>
          </header>
          <div className="worktrees-form-modal__picker">
            <p className="worktrees-form-modal__picker-label">Repository path</p>
            <button
              type="button"
              className="secondary"
              disabled={pending}
              onClick={() => setPickerOpen(true)}
            >
              Choose repository
            </button>
            {path.trim() !== "" ? (
              <p className="worktrees-form-modal__selected">
                Selected: <code>{path}</code>
              </p>
            ) : (
              <p className="worktrees-form-modal__picker-empty">No repository selected yet.</p>
            )}
          </div>
          {errorMessage ? (
            <MutationErrorBanner error={errorMessage} className="worktrees-form-modal__error" />
          ) : null}
          <div className="row stack-row-actions">
            <Button type="button" variant="secondary" disabled={pending} onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              disabled={pending || !path.trim()}
              loading={pending}
            >
              Register
            </Button>
          </div>
        </form>
      </Modal>
      <WorkspaceDirPickerModal
        open={pickerOpen}
        nested
        requireGitRepository
        rootsScope="expanded"
        title="Choose repository"
        lead="Select any git folder of the repository. Linked folders resolve to one registration at the main repository path."
        selectionFooterLabel="Repository to register"
        confirmLabel="Use this repository"
        currentPath={path}
        onClose={() => setPickerOpen(false)}
        onSelect={(next) => {
          setPath(next);
          setPickerOpen(false);
        }}
      />
    </>
  );
}
