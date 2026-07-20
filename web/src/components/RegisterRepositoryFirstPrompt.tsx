import { Modal } from "@/shared/Modal";
import { Button } from "@/components/ui";

type Props = {
  open: boolean;
  onClose: () => void;
  /** Navigate or open register flow — owned by the caller (vertical / app). */
  onRegister: () => void;
};

/**
 * Presentational “register a repository first” confirm. Callers supply
 * `onRegister` so this shell stays vertical-free (no `/worktrees` coupling).
 */
export function RegisterRepositoryFirstPrompt({
  open,
  onClose,
  onRegister,
}: Props) {
  if (!open) return null;

  return (
    <Modal
      onClose={onClose}
      labelledBy="repository-setup-prompt-title"
      describedBy="repository-setup-prompt-lead"
    >
      <section className="panel modal-sheet worktrees-form-modal">
        <header className="worktrees-form-modal__header">
          <h2 id="repository-setup-prompt-title">Register a repository first</h2>
          <p id="repository-setup-prompt-lead" className="worktrees-form-modal__lead">
            Hamix needs a registered git checkout before you can create tasks with worktrees
            and branches.
          </p>
        </header>
        <div className="row stack-row-actions">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="button" variant="primary" onClick={onRegister}>
            Register repository
          </Button>
        </div>
      </section>
    </Modal>
  );
}
