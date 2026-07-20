import { Modal } from "@/shared/Modal";
import {
  FolderRow,
  PickerBreadcrumb,
  RootGroup,
} from "./WorkspacePickerListParts";
import {
  useWorkspaceDirPickerState,
  type WorkspaceDirPickerModalProps,
} from "./useWorkspaceDirPickerState";
import "./workspace-picker.css";

export function WorkspaceDirPickerModal({
  open,
  onClose,
  nested = false,
  title = "Choose folder",
  ...pickerProps
}: WorkspaceDirPickerModalProps) {
  const picker = useWorkspaceDirPickerState({
    open,
    onClose,
    ...pickerProps,
  });

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
          <h2 id="workspace-dir-picker-title" className="workspace-picker-title">
            {title}
          </h2>
          <p id="workspace-dir-picker-lead" className="workspace-picker-lead">
            {picker.resolvedLead}
          </p>
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
          <>
            {picker.atRoots ? (
              <p className="workspace-picker-section-label">
                Choose a folder to browse from
              </p>
            ) : (
              <PickerBreadcrumb
                crumbs={picker.crumbs}
                listingPending={picker.listingPending}
                backToRoots={picker.backToRoots}
                onBack={picker.goBack}
                onJump={(path) => void picker.loadListing(path)}
              />
            )}

            {picker.listingError ? (
              <p
                className="workspace-picker-status workspace-picker-status--error"
                role="alert"
              >
                {picker.listingError}
              </p>
            ) : null}

            <ul className="workspace-picker-list" aria-busy={picker.listingPending}>
              {picker.atRoots && picker.rootGroups ? (
                <>
                  {picker.rootGroups.workspace.length > 0 ? (
                    <RootGroup
                      label="Workspace"
                      roots={picker.rootGroups.workspace}
                      listingPending={picker.listingPending}
                      onOpen={(path) => void picker.loadListing(path)}
                    />
                  ) : null}
                  {picker.rootGroups.userFolders.length > 0 ? (
                    <RootGroup
                      label="User folders"
                      roots={picker.rootGroups.userFolders}
                      listingPending={picker.listingPending}
                      onOpen={(path) => void picker.loadListing(path)}
                    />
                  ) : null}
                </>
              ) : null}
              {!picker.atRoots
                ? picker.entries.map((entry) => (
                    <li key={entry.path}>
                      <FolderRow
                        name={entry.name}
                        gitRepoStatus={
                          picker.requireGitRepository
                            ? entry.is_git_repo
                            : entry.is_git_repo
                              ? true
                              : undefined
                        }
                        disabled={picker.listingPending}
                        onClick={() => void picker.loadListing(entry.path)}
                      />
                    </li>
                  ))
                : null}
              {!picker.atRoots && !picker.listingPending && picker.entries.length === 0 ? (
                <li className="workspace-picker-empty">
                  <p className="workspace-picker-empty-title">
                    No subfolders inside this folder.
                  </p>
                  <p className="workspace-picker-empty-hint">
                    {picker.requireGitRepository && !picker.currentPathIsGitRepo
                      ? "This folder is not a git checkout. Go back and open a folder that contains a .git directory."
                      : picker.requireGitRepository
                        ? "Use the button below to register this git checkout, or go back to pick a different folder."
                        : "Use the button below to register this folder, or go back to pick a different one."}
                  </p>
                </li>
              ) : null}
            </ul>

            <footer className="workspace-picker-footer">
              <div className="workspace-picker-selection" aria-live="polite">
                <span className="workspace-picker-selection-label">
                  {picker.selectionLabel}
                </span>
                <code
                  className="workspace-picker-selection-path"
                  data-empty={!picker.hasOpenFolder}
                >
                  {picker.hasOpenFolder ? picker.currentBrowsePath : "Open a folder to register it"}
                </code>
                {picker.validatingPath ? (
                  <p className="workspace-picker-validation">Checking folder…</p>
                ) : null}
                {!picker.validatingPath && picker.pathValidation && !picker.pathValidation.ok && picker.pathValidation.message ? (
                  <p className="workspace-picker-validation workspace-picker-validation--error" role="alert">
                    {picker.pathValidation.message}
                  </p>
                ) : null}
              </div>
              <div className="workspace-picker-footer-actions">
                <button type="button" className="secondary" onClick={onClose}>
                  Cancel
                </button>
                <button
                  type="button"
                  disabled={!picker.canConfirm}
                  onClick={picker.confirmSelection}
                >
                  {picker.confirmButtonLabel}
                </button>
              </div>
            </footer>
          </>
        ) : null}
      </div>
    </Modal>
  );
}
