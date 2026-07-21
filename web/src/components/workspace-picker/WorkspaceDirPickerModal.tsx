import { Modal } from "@/shared/Modal";
import { PickerHeaderGitIcon, PickerCloseIcon } from "./WorkspacePickerIcons";
import { WorkspacePickerReadyBody } from "./WorkspacePickerReadyBody";
import {
  useWorkspaceDirPickerState,
  type WorkspaceDirPickerModalProps,
} from "./useWorkspaceDirPickerState";
import "./workspace-picker.css";

export function WorkspaceDirPickerModal({
  open,
  onClose,
  nested = false,
  title,
  ...pickerProps
}: WorkspaceDirPickerModalProps) {
  const picker = useWorkspaceDirPickerState({
    open,
    onClose,
    ...pickerProps,
  });

  const resolvedTitle =
    title ??
    (picker.requireGitRepository ? "Choose a repository" : "Choose folder");

  if (!open) return null;

  return (
    <Modal
      labelledBy="workspace-dir-picker-title"
      describedBy="workspace-dir-picker-lead"
      size="wide"
      stack={nested ? "nested" : "default"}
      lockBodyScroll={!nested}
      onClose={onClose}
    >
      <div className="panel modal-sheet workspace-picker-modal">
        <header className="workspace-picker-header">
          <div className="workspace-picker-header-icon" aria-hidden="true">
            <PickerHeaderGitIcon />
          </div>
          <div className="workspace-picker-header-copy">
            <h2 id="workspace-dir-picker-title" className="workspace-picker-title">
              {resolvedTitle}
            </h2>
            <p id="workspace-dir-picker-lead" className="workspace-picker-lead">
              {picker.resolvedLead}
            </p>
          </div>
          <button
            type="button"
            className="workspace-picker-close"
            onClick={onClose}
            aria-label="Close"
          >
            <PickerCloseIcon />
          </button>
        </header>

        {picker.loadState.kind === "loading" ? (
          <p className="workspace-picker-status">Loading locations…</p>
        ) : null}

        {picker.loadState.kind === "error" ? (
          <p
            className="workspace-picker-status workspace-picker-status--error"
            role="alert"
          >
            {picker.loadState.message}
          </p>
        ) : null}

        {picker.loadState.kind === "ready" ? (
          <WorkspacePickerReadyBody picker={picker} onClose={onClose} />
        ) : null}
      </div>
    </Modal>
  );
}
