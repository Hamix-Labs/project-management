import { useId, useState } from "react";
import { RichPromptEditor } from "@/components/rich-prompt";
import { promptHasVisibleContent } from "@/lib/promptFormat";
import { Modal } from "@/shared/Modal";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";
import type { ChecklistItemDraft } from "@/types";
import { TaskPolishAddCriteria } from "./TaskPolishAddCriteria";
import { TaskPolishCriteriaList } from "./TaskPolishCriteriaList";
import { CloseGlyph, SparkleGlyph } from "./TaskPolishGlyphs";
import { usePolishCriterionDrafts } from "./usePolishCriterionDrafts";

export type PolishCriterionOption = {
  id: string;
  text: string;
};

export type PolishConfirmPayload = {
  instructions: string;
  flaggedCriterionIds: string[];
  newCriteria: ChecklistItemDraft[];
};

type Props = {
  worktreeId?: string;
  criteria?: PolishCriterionOption[];
  saving: boolean;
  pending: boolean;
  error?: string | null;
  onCancel: () => void;
  onConfirm: (payload: PolishConfirmPayload) => void;
};

export function TaskPolishDialog({
  worktreeId,
  criteria = [],
  saving,
  pending,
  error = null,
  onCancel,
  onConfirm,
}: Props) {
  const titleId = useId();
  const descriptionId = useId();
  const instructionsId = useId();
  const instructionsLabelId = `${instructionsId}-label`;
  const flaggedHeadingId = useId();
  const addHeadingId = useId();
  const [instructions, setInstructions] = useState("");
  const [flaggedIds, setFlaggedIds] = useState<Set<string>>(() => new Set());
  const drafts = usePolishCriterionDrafts();
  const canSubmit =
    promptHasVisibleContent(instructions) && !saving && !pending;
  const controlsDisabled = saving || pending;

  function toggleFlagged(id: string) {
    setFlaggedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  return (
    <Modal
      onClose={onCancel}
      labelledBy={titleId}
      describedBy={descriptionId}
      size="wide"
      busy={pending}
      busyLabel="Queueing polish…"
      dismissibleWhileBusy
    >
      <section className="panel modal-sheet task-polish-dialog">
        <header className="task-polish-dialog__header">
          <div className="task-polish-dialog__header-main">
            <span className="task-polish-dialog__badge" aria-hidden="true">
              <SparkleGlyph size={20} />
            </span>
            <div className="task-polish-dialog__title-block">
              <h2 id={titleId} className="task-polish-dialog__title">
                Polish this task
              </h2>
              <p className="task-polish-dialog__blurb" id={descriptionId}>
                Resume the existing agent conversation with new instructions.
                The task returns to awaiting review when the polish finishes.
              </p>
            </div>
          </div>
          <button
            type="button"
            className="task-polish-dialog__close"
            aria-label="Close"
            disabled={saving}
            onClick={onCancel}
          >
            <CloseGlyph />
          </button>
        </header>

        <div className="task-polish-dialog__body">
          <TaskPolishCriteriaList
            criteria={criteria}
            flaggedIds={flaggedIds}
            disabled={controlsDisabled}
            headingId={flaggedHeadingId}
            onToggle={toggleFlagged}
          />
          <TaskPolishAddCriteria
            headingId={addHeadingId}
            newCriteria={drafts.newCriteria}
            disabled={controlsDisabled}
            modalOpen={drafts.modalOpen}
            modalText={drafts.modalText}
            modalCommands={drafts.modalCommands}
            onOpenModal={drafts.openModal}
            onCloseModal={drafts.closeModal}
            onModalTextChange={drafts.setModalText}
            onModalCommandsChange={drafts.setModalCommands}
            onSubmitModal={drafts.submitModal}
            onRemove={drafts.removeAt}
          />

          <div className="task-polish-dialog__section">
            <div className="task-polish-dialog__label-row">
              <label
                id={instructionsLabelId}
                htmlFor={instructionsId}
                className="task-polish-dialog__label"
              >
                Instructions
              </label>
              <span className="task-polish-dialog__hint">
                Type <kbd>@</kbd> to reference files
              </span>
            </div>
            <div className="task-create-editor-shell">
              <RichPromptEditor
                id={instructionsId}
                value={instructions}
                onChange={setInstructions}
                disabled={controlsDisabled}
                placeholder="Describe what should change in this polish pass…"
                worktreeId={worktreeId?.trim() || undefined}
              />
            </div>
          </div>
        </div>

        <MutationErrorBanner error={error} className="task-polish-dialog__err" />

        <footer className="task-polish-dialog__footer">
          <span className="task-polish-dialog__esc-hint" aria-hidden="true">
            <kbd>Esc</kbd> to cancel
          </span>
          <div className="task-polish-dialog__actions">
            <button
              type="button"
              className="secondary task-polish-dialog__cancel"
              onClick={onCancel}
              disabled={saving}
            >
              Cancel
            </button>
            <button
              type="button"
              className="primary task-polish-dialog__submit"
              disabled={!canSubmit}
              onClick={() =>
                onConfirm({
                  instructions,
                  flaggedCriterionIds: Array.from(flaggedIds),
                  newCriteria: drafts.newCriteria,
                })
              }
            >
              {pending ? (
                "Queueing…"
              ) : (
                <>
                  <SparkleGlyph size={16} />
                  Polish
                </>
              )}
            </button>
          </div>
        </footer>
      </section>
    </Modal>
  );
}
