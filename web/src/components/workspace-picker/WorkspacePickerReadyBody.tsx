import {
  FolderRow,
  PickerBreadcrumb,
  RootGroup,
} from "./WorkspacePickerListParts";
import { PickerInfoIcon, PickerSearchIcon } from "./WorkspacePickerIcons";
import type { useWorkspaceDirPickerState } from "./useWorkspaceDirPickerState";

type Picker = ReturnType<typeof useWorkspaceDirPickerState>;

type Props = {
  picker: Picker;
  onClose: () => void;
};

/** Toolbar, folder list, and confirm footer once roots have loaded. */
export function WorkspacePickerReadyBody({ picker, onClose }: Props) {
  return (
    <>
      <div className="workspace-picker-toolbar">
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
        <div className="workspace-picker-filter">
          <PickerSearchIcon className="workspace-picker-filter-icon" />
          <input
            type="search"
            value={picker.folderFilter}
            onChange={(e) => picker.setFolderFilter(e.target.value)}
            placeholder="Filter folders"
            aria-label="Filter folders"
            className="workspace-picker-filter-input"
          />
        </div>
      </div>

      {picker.listingError ? (
        <p
          className="workspace-picker-status workspace-picker-status--error"
          role="alert"
        >
          {picker.listingError}
        </p>
      ) : null}

      <WorkspacePickerEntryList picker={picker} />
      <WorkspacePickerFooter picker={picker} onClose={onClose} />
    </>
  );
}

function WorkspacePickerEntryList({ picker }: { picker: Picker }) {
  return (
    <ul
      className="workspace-picker-list"
      aria-busy={picker.listingPending || picker.probePending}
    >
      {picker.filterEmpty ? (
        <li className="workspace-picker-empty">
          <p className="workspace-picker-empty-title">
            No folders match &ldquo;{picker.folderFilter.trim()}&rdquo;.
          </p>
        </li>
      ) : null}
      {!picker.filterEmpty && picker.atRoots && picker.rootGroups ? (
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
      {!picker.filterEmpty && !picker.atRoots
        ? picker.entries.map((entry) => {
            const selectGit =
              picker.requireGitRepository && entry.is_git_repo === true;
            const selected =
              selectGit &&
              picker.probedPath !== "" &&
              (picker.probedPath === entry.path ||
                picker.footerPath === entry.path);
            return (
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
                  disabled={picker.listingPending || picker.probePending}
                  selected={selected}
                  onClick={() => {
                    if (selectGit) {
                      void picker.resolveGitSelection(entry.path);
                      return;
                    }
                    void picker.loadListing(entry.path);
                  }}
                  onOpen={
                    selectGit
                      ? () => void picker.loadListing(entry.path)
                      : undefined
                  }
                />
              </li>
            );
          })
        : null}
      {!picker.filterEmpty &&
      !picker.atRoots &&
      !picker.listingPending &&
      picker.entries.length === 0 ? (
        <li className="workspace-picker-empty">
          <p className="workspace-picker-empty-title">
            No subfolders inside this folder.
          </p>
          <p className="workspace-picker-empty-hint">
            {picker.requireGitRepository && !picker.hasResolvedRepo
              ? "This folder is not a git repository. Go back and select a folder with a git badge."
              : picker.requireGitRepository
                ? "Use the button below to register this repository, or go back to pick a different folder."
                : "Use the button below to register this folder, or go back to pick a different one."}
          </p>
        </li>
      ) : null}
    </ul>
  );
}

function WorkspacePickerFooter({
  picker,
  onClose,
}: {
  picker: Picker;
  onClose: () => void;
}) {
  return (
    <footer className="workspace-picker-footer">
      <div className="workspace-picker-summary" aria-live="polite">
        <div className="workspace-picker-summary-label">
          <PickerInfoIcon />
          <span>{picker.selectionLabel}</span>
        </div>
        <code
          className="workspace-picker-selection-path"
          data-empty={!picker.footerPath}
        >
          {picker.footerPath || picker.footerEmptyHint}
        </code>
        {picker.probePending ? (
          <p className="workspace-picker-validation">Resolving repository…</p>
        ) : null}
        {!picker.probePending && picker.probeError ? (
          <p
            className="workspace-picker-validation workspace-picker-validation--error"
            role="alert"
          >
            {picker.probeError}
          </p>
        ) : null}
        {picker.validatingPath ? (
          <p className="workspace-picker-validation">Checking folder…</p>
        ) : null}
        {!picker.validatingPath &&
        picker.pathValidation &&
        !picker.pathValidation.ok &&
        picker.pathValidation.message ? (
          <p
            className="workspace-picker-validation workspace-picker-validation--error"
            role="alert"
          >
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
  );
}
