import { useId, useState } from "react";
import { RichPromptEditor } from "@/components/rich-prompt";
import { useProjectContextPromptBinding } from "@/hooks/useProjectContextPromptBinding";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import { promptHasVisibleContent } from "@/lib/promptFormat";
import { Modal } from "@/shared/Modal";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";
import { TaskPolishAddCriteria } from "./TaskPolishAddCriteria";
import { TaskPolishCriteriaList } from "./TaskPolishCriteriaList";

export type PolishCriterionOption = {
  id: string;
  text: string;
};

export type PolishConfirmPayload = {
  instructions: string;
  flaggedCriterionIds: string[];
  newCriteria: string[];
};

type Props = {
  worktreeId?: string;
  projectId?: string;
  projectContextItemIds?: string[];
  criteria?: PolishCriterionOption[];
  saving: boolean;
  pending: boolean;
  error?: string | null;
  onCancel: () => void;
  onConfirm: (payload: PolishConfirmPayload) => void;
};

function SparkleGlyph({ size = 16 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.4"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M8 2.25L9.35 6.4 13.5 7.75 9.35 9.1 8 13.25 6.65 9.1 2.5 7.75 6.65 6.4z" />
      <path d="M12.5 2.25v2" />
      <path d="M11.5 3.25h2" />
    </svg>
  );
}

function CloseGlyph() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      aria-hidden="true"
    >
      <path d="M4 4l8 8M12 4l-8 8" />
    </svg>
  );
}

export function TaskPolishDialog({
  worktreeId,
  projectId = "",
  projectContextItemIds = [],
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
  const [newDraft, setNewDraft] = useState("");
  const [newCriteria, setNewCriteria] = useState<string[]>([]);
  const [selectedContextIds, setSelectedContextIds] = useState(
    projectContextItemIds,
  );
  const projectsUiEnabled = !isUiFeatureOmitted("projects");
  const promptProjectContext = useProjectContextPromptBinding({
    projectId: projectsUiEnabled ? projectId : "",
    selectedIds: selectedContextIds,
    onSelectedIdsChange: setSelectedContextIds,
  });
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

  function addNewCriterion() {
    const text = newDraft.trim();
    if (!text) return;
    setNewCriteria((prev) => [...prev, text]);
    setNewDraft("");
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
            newCriteria={newCriteria}
            draft={newDraft}
            disabled={controlsDisabled}
            onDraftChange={setNewDraft}
            onAdd={addNewCriterion}
            onRemove={(index) =>
              setNewCriteria((prev) => prev.filter((_, i) => i !== index))
            }
          />

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
              projectContext={promptProjectContext ?? undefined}
            />
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
                  newCriteria,
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
