import { useEffect, useState } from "react";
import { Button } from "@/components/ui";
import { Modal } from "@/shared/Modal";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";
import { WorkspaceDirPickerModal } from "@/components/workspace-picker";
import {
  gitRegisterErrorMessage,
  isDuplicateRegisterError,
} from "../gitRegisterErrors";
import {
  isRepositoryAlreadyRegistered,
  type RegisteredRepositoryPaths,
} from "../isRepositoryAlreadyRegistered";
import {
  RegisterAlertIcon,
  RegisterCheckIcon,
  RegisterCloseIcon,
  RegisterFolderSearchIcon,
  RegisterHeaderGitIcon,
  RegisterStatusCheckIcon,
} from "./RegisterRepositoryIcons";
import "./register-repository-modal.css";

type Props = {
  open: boolean;
  pending: boolean;
  error: unknown;
  registeredRepositories: readonly RegisteredRepositoryPaths[];
  onClose: () => void;
  onSubmit: (input: { path: string }) => void;
  onClearError?: () => void;
};

function repoBasename(path: string): string {
  const normalized = path.replace(/\\/g, "/");
  const parts = normalized.split("/").filter(Boolean);
  return parts[parts.length - 1] ?? path;
}

export function RegisterRepositoryModal({
  open,
  pending,
  error,
  registeredRepositories,
  onClose,
  onSubmit,
  onClearError,
}: Props) {
  const [path, setPath] = useState("");
  const [pickerOpen, setPickerOpen] = useState(false);

  // Parent keeps this mounted while closed; clear selection so reopen starts empty.
  useEffect(() => {
    if (!open) {
      setPath("");
      setPickerOpen(false);
    }
  }, [open]);

  if (!open) return null;

  const trimmed = path.trim();
  const hasSelection = trimmed !== "";
  const alreadyRegistered =
    hasSelection && isRepositoryAlreadyRegistered(trimmed, registeredRepositories);
  const duplicateError = isDuplicateRegisterError(error);
  const blocked = alreadyRegistered || duplicateError;
  const errorMessage =
    error != null && !alreadyRegistered && !duplicateError
      ? gitRegisterErrorMessage(error)
      : null;
  const repoName = hasSelection ? repoBasename(trimmed) : null;

  return (
    <>
      <Modal
        onClose={onClose}
        labelledBy="register-repo-title"
        busy={pending}
        dismissibleWhileBusy={false}
      >
        <form
          className="panel modal-sheet register-repo-modal"
          onSubmit={(e) => {
            e.preventDefault();
            if (!trimmed || blocked) return;
            onSubmit({ path: trimmed });
          }}
        >
          <header className="register-repo-modal__header">
            <div className="register-repo-modal__header-icon" aria-hidden="true">
              <RegisterHeaderGitIcon />
            </div>
            <div className="register-repo-modal__header-copy">
              <h2 id="register-repo-title" className="register-repo-modal__title">
                Register repository
              </h2>
            </div>
            <button
              type="button"
              className="register-repo-modal__close"
              aria-label="Close"
              disabled={pending}
              onClick={onClose}
            >
              <RegisterCloseIcon />
            </button>
          </header>

          <div className="register-repo-modal__body">
            <p className="register-repo-modal__label">Repository path</p>

            <button
              id="choose-repo-button"
              type="button"
              disabled={pending}
              aria-label={hasSelection ? "Change repository" : "Choose repository"}
              onClick={() => setPickerOpen(true)}
              className={
                hasSelection
                  ? "register-repo-modal__chooser register-repo-modal__chooser--selected"
                  : "register-repo-modal__chooser register-repo-modal__chooser--empty"
              }
            >
              <span
                className={
                  hasSelection
                    ? "register-repo-modal__chooser-icon register-repo-modal__chooser-icon--selected"
                    : "register-repo-modal__chooser-icon"
                }
              >
                {hasSelection ? <RegisterCheckIcon /> : <RegisterFolderSearchIcon />}
              </span>
              <span className="register-repo-modal__chooser-copy">
                <span className="register-repo-modal__chooser-title">
                  {hasSelection ? repoName : "Choose repository"}
                </span>
                <span className="register-repo-modal__chooser-hint">
                  {hasSelection
                    ? trimmed
                    : "Browse your folders to select a Git repository"}
                </span>
              </span>
              <span className="register-repo-modal__chooser-action">
                {hasSelection ? "Change" : "Browse"}
              </span>
            </button>

            <div
              className={
                blocked
                  ? "register-repo-modal__status register-repo-modal__status--blocked"
                  : "register-repo-modal__status"
              }
            >
              {hasSelection ? (
                blocked ? (
                  <>
                    <RegisterAlertIcon className="register-repo-modal__status-icon register-repo-modal__status-icon--blocked" />
                    <span>This repository is already registered.</span>
                  </>
                ) : (
                  <>
                    <RegisterStatusCheckIcon className="register-repo-modal__status-icon register-repo-modal__status-icon--ready" />
                    <span>Ready to register this repository.</span>
                  </>
                )
              ) : (
                <>
                  <RegisterAlertIcon className="register-repo-modal__status-icon" />
                  <span>No repository selected yet.</span>
                </>
              )}
            </div>
          </div>

          {errorMessage ? (
            <MutationErrorBanner error={errorMessage} className="register-repo-modal__error" />
          ) : null}

          <footer className="register-repo-modal__footer">
            <Button type="button" variant="secondary" disabled={pending} onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              disabled={pending || !hasSelection || blocked}
              loading={pending}
            >
              Register
            </Button>
          </footer>
        </form>
      </Modal>
      <WorkspaceDirPickerModal
        open={pickerOpen}
        nested
        requireGitRepository
        rootsScope="expanded"
        title="Choose a repository"
        lead="Select any worktree of a Git repository. Repositories are identified by their primary checkout, so all linked worktrees count as one."
        selectionFooterLabel="Repository to register"
        confirmLabel="Use this repository"
        currentPath={path}
        onClose={() => setPickerOpen(false)}
        onSelect={(next) => {
          setPath(next);
          setPickerOpen(false);
          onClearError?.();
        }}
      />
    </>
  );
}
