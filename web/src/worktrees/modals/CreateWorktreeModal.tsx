import { useEffect, useState } from "react";
import { Modal } from "@/shared/Modal";
import { FieldLabel } from "@/shared/FieldLabel";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";
import { CustomSelect } from "@/components/custom-select";
import { WorkspaceDirPickerModal } from "@/components/workspace-picker";
import { gitDeleteErrorMessage } from "../gitDeleteErrors";
import { useGlobalLiveWorktrees } from "../hooks/useGlobalLiveWorktrees";
import { useAutoReconcileInventory } from "../hooks/useAutoReconcileInventory";
import { useLiveInventoryUnreachable } from "../hooks/useLiveInventoryUnreachable";
import { worktreeGitCopy } from "../worktreeGitCopy";
import { liveWorktreeOptionLabel } from "../worktreeGitCopy";
import { joinWorktreeCreatePath } from "../worktreeCreatePath";
import { parentBrowsePath } from "../parentBrowsePath";
import {
  WorktreeBranchBindFields,
  branchBindPayload,
  type BranchBindValue,
} from "../components/WorktreeBranchBindFields";
import { WorktreeInventoryReconcilePrompt } from "../components/WorktreeInventoryReconcilePrompt";
import { WorktreeInventorySyncStatus } from "../components/WorktreeInventorySyncStatus";

type StartFromMode = "main" | "reference";

type Props = {
  open: boolean;
  pending: boolean;
  error: unknown;
  repositoryId: string;
  storedPath: string;
  reconcilePending?: boolean;
  inventoryRefreshPending?: boolean;
  reconcileError?: unknown;
  reconcileBlocked?: boolean;
  onReconcile: () => void;
  onClose: () => void;
  onSubmit: (input: {
    path: string;
    name?: string;
    branch: string;
    create_branch?: boolean;
    start_point?: string;
  }) => void;
};

export function CreateWorktreeModal({
  open,
  pending,
  error,
  repositoryId,
  storedPath,
  reconcilePending = false,
  inventoryRefreshPending = false,
  reconcileError,
  reconcileBlocked = false,
  onReconcile,
  onClose,
  onSubmit,
}: Props) {
  const [parentPath, setParentPath] = useState("");
  const [folderName, setFolderName] = useState("");
  const [name, setName] = useState("");
  const [startFrom, setStartFrom] = useState<StartFromMode>("main");
  const [referencePath, setReferencePath] = useState("");
  const [branchBind, setBranchBind] = useState<BranchBindValue>({
    selectedBranchName: "",
    newBranchName: "",
    createNew: false,
  });
  const [pickerOpen, setPickerOpen] = useState(false);

  useEffect(() => {
    if (!open) {
      setParentPath("");
      setFolderName("");
      setName("");
      setStartFrom("main");
      setReferencePath("");
      setBranchBind({ selectedBranchName: "", newBranchName: "", createNew: false });
      setPickerOpen(false);
      return;
    }
    setParentPath(parentBrowsePath(storedPath));
    setFolderName("");
    setName("");
    setStartFrom("main");
    setReferencePath("");
    setBranchBind({ selectedBranchName: "", newBranchName: "", createNew: false });
    setPickerOpen(false);
  }, [open, storedPath]);

  const liveWorktreesQuery = useGlobalLiveWorktrees(repositoryId, {
    enabled: open && repositoryId !== "",
  });
  const inventoryUnreachable = useLiveInventoryUnreachable(liveWorktreesQuery);
  useAutoReconcileInventory({
    enabled: open && repositoryId !== "",
    inventoryUnreachable,
    reconcilePending,
    reconcileBlocked,
    onReconcile,
  });
  const referenceOptions = (liveWorktreesQuery.data ?? [])
    .filter((wt) => wt.branch.trim() !== "")
    .map((wt) => ({
      value: wt.path,
      label: liveWorktreeOptionLabel(wt.path, wt.is_main),
    }));
  const referenceWorktree = liveWorktreesQuery.data?.find((wt) => wt.path === referencePath);
  const referenceDetached = startFrom === "reference" && referencePath !== "" && !referenceWorktree?.branch.trim();
  const defaultParentPath = parentBrowsePath(storedPath);
  const isDefaultParent =
    parentPath.trim() !== "" && parentPath === defaultParentPath;

  if (!open) return null;

  const errorMessage = error != null ? gitDeleteErrorMessage(error) : null;
  const branchPayload = branchBindPayload(branchBind);
  const referenceReady = startFrom === "main" || (referencePath !== "" && !referenceDetached);
  const fullPath = joinWorktreeCreatePath(parentPath, folderName);
  const canSubmit =
    !inventoryUnreachable &&
    !inventoryRefreshPending &&
    fullPath != null &&
    branchPayload != null &&
    referenceReady;

  const startPoint =
    startFrom === "reference" && branchPayload?.create_branch && referenceWorktree?.branch.trim()
      ? referenceWorktree.branch.trim()
      : undefined;

  return (
    <>
      <Modal
        onClose={onClose}
        labelledBy="create-worktree-title"
        busy={pending}
        dismissibleWhileBusy={false}
      >
        <form
          className="panel modal-sheet worktrees-form-modal"
          onSubmit={(e) => {
            e.preventDefault();
            if (!fullPath || !branchPayload) return;
            onSubmit({
              path: fullPath,
              name: name.trim() || undefined,
              branch: branchPayload.name,
              create_branch: branchPayload.create_branch,
              ...(startPoint ? { start_point: startPoint } : {}),
            });
          }}
        >
          <header className="worktrees-form-modal__header">
            <h2 id="create-worktree-title">{worktreeGitCopy.createModalTitle}</h2>
            <p className="worktrees-form-modal__lead">{worktreeGitCopy.createModalLead}</p>
          </header>

          <WorktreeInventorySyncStatus
            pending={inventoryRefreshPending && !inventoryUnreachable}
          />

          {inventoryUnreachable ? (
            <WorktreeInventoryReconcilePrompt
              storedPath={storedPath}
              pending={reconcilePending}
              reconcileError={reconcileError}
              onReconcile={onReconcile}
            />
          ) : null}
          {!inventoryUnreachable ? (
            <>
          <fieldset className="worktrees-form-modal__fieldset">
            <legend className="settings-field-label">{worktreeGitCopy.createModalStartFromLabel}</legend>
            <label className="worktrees-form-modal__radio">
              <input
                type="radio"
                name="create-worktree-start-from"
                checked={startFrom === "main"}
                disabled={pending || inventoryRefreshPending}
                onChange={() => setStartFrom("main")}
              />
              {worktreeGitCopy.createModalStartFromMain}
            </label>
            <label className="worktrees-form-modal__radio">
              <input
                type="radio"
                name="create-worktree-start-from"
                checked={startFrom === "reference"}
                disabled={pending || inventoryRefreshPending}
                onChange={() => setStartFrom("reference")}
              />
              {worktreeGitCopy.createModalStartFromReference}
            </label>
          </fieldset>
          {startFrom === "reference" ? (
            <CustomSelect
              id="create-worktree-reference-select"
              label={worktreeGitCopy.createModalReferenceLabel}
              value={referencePath}
              options={referenceOptions}
              disabled={
                pending ||
                inventoryRefreshPending ||
                liveWorktreesQuery.isLoading ||
                referenceOptions.length === 0
              }
              requirement="required"
              onChange={setReferencePath}
            />
          ) : null}
          {referenceDetached ? (
            <p className="worktrees-form-modal__picker-empty" role="alert">
              {worktreeGitCopy.createModalReferenceDetached}
            </p>
          ) : null}
          <div className="worktrees-form-modal__picker">
            <p className="worktrees-form-modal__picker-label">{worktreeGitCopy.createModalLocationLabel}</p>
            <p className="worktrees-form-modal__picker-hint">{worktreeGitCopy.createModalLocationHint}</p>
            {parentPath.trim() !== "" ? (
              <div
                className={[
                  "worktrees-form-modal__parent-choice",
                  isDefaultParent ? "worktrees-form-modal__parent-choice--default" : null,
                ]
                  .filter(Boolean)
                  .join(" ")}
              >
                <div className="worktrees-form-modal__parent-choice-header">
                  {isDefaultParent ? (
                    <span className="worktrees-form-modal__parent-default-badge">Default</span>
                  ) : (
                    <span className="worktrees-form-modal__parent-choice-label">
                      {worktreeGitCopy.createModalParentSelectedPrefix}
                    </span>
                  )}
                </div>
                <code className="worktrees-form-modal__parent-choice-path">{parentPath}</code>
              </div>
            ) : null}
            <button
              type="button"
              className="secondary"
              disabled={pending || inventoryRefreshPending}
              onClick={() => setPickerOpen(true)}
            >
              {isDefaultParent
                ? worktreeGitCopy.createModalChangeParentFolder
                : worktreeGitCopy.createModalChooseParentFolder}
            </button>
          </div>
          <div className="field">
            <FieldLabel htmlFor="create-worktree-folder-name" requirement="required">
              {worktreeGitCopy.createModalFolderNameLabel}
            </FieldLabel>
            <p className="worktrees-form-modal__field-hint">{worktreeGitCopy.createModalFolderNameHint}</p>
            <input
              id="create-worktree-folder-name"
              type="text"
              value={folderName}
              disabled={pending || inventoryRefreshPending}
              onChange={(e) => setFolderName(e.target.value)}
              placeholder={worktreeGitCopy.createModalFolderNamePlaceholder}
              required
            />
          </div>
          {fullPath ? (
            <p className="worktrees-form-modal__selected">
              {worktreeGitCopy.createModalFullPathPrefix} <code>{fullPath}</code>
            </p>
          ) : null}
          <div className="field">
            <span className="settings-field-label">{worktreeGitCopy.createModalDisplayNameLabel}</span>
            <p className="worktrees-form-modal__field-hint">{worktreeGitCopy.createModalDisplayNameHint}</p>
            <input
              id="create-worktree-display-name"
              type="text"
              value={name}
              disabled={pending || inventoryRefreshPending}
              onChange={(e) => setName(e.target.value)}
              placeholder={worktreeGitCopy.createModalDisplayNamePlaceholder}
            />
          </div>
          <WorktreeBranchBindFields
            repositoryId={repositoryId}
            enabled={open && repositoryId !== "" && !inventoryRefreshPending}
            pending={pending}
            value={branchBind}
            onChange={setBranchBind}
            branchSelectId="create-worktree-branch-select"
            newBranchInputId="create-worktree-branch-new-name"
          />
            </>
          ) : null}
          {errorMessage ? (
            <MutationErrorBanner error={errorMessage} className="worktrees-form-modal__error" />
          ) : null}
          <div className="row stack-row-actions">
            <button type="button" className="secondary" disabled={pending} onClick={onClose}>
              {worktreeGitCopy.cancel}
            </button>
            <button
              type="submit"
              className="btn-primary"
              disabled={pending || !canSubmit}
            >
              {pending ? worktreeGitCopy.createModalSubmitting : worktreeGitCopy.createModalSubmit}
            </button>
          </div>
        </form>
      </Modal>
      <WorkspaceDirPickerModal
        open={pickerOpen}
        nested
        currentPath={parentPath}
        rootsScope="expanded"
        title={worktreeGitCopy.createModalPickerTitle}
        lead={worktreeGitCopy.createModalPickerLead}
        selectionFooterLabel={worktreeGitCopy.createModalPickerSelectionLabel}
        confirmLabel={worktreeGitCopy.createModalPickerConfirmLabel}
        initialBrowsePath={
          parentBrowsePath(storedPath) !== "" ? parentBrowsePath(storedPath) : undefined
        }
        onClose={() => setPickerOpen(false)}
        onSelect={(next) => {
          setParentPath(next);
          setPickerOpen(false);
        }}
      />
    </>
  );
}
