import { useId, useState } from "react";
import { RichPromptEditor } from "@/components/rich-prompt";
import { useProjectContextPromptBinding } from "@/hooks/useProjectContextPromptBinding";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";
import { promptHasVisibleContent } from "@/lib/promptFormat";
import { FieldLabel } from "@/shared/FieldLabel";
import { Modal } from "@/shared/Modal";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";

type Props = {
  /** Scopes @ file mentions to the task worktree (same as create-task prompt). */
  worktreeId?: string;
  /** When set, enables # project-context mentions like create-task. */
  projectId?: string;
  projectContextItemIds?: string[];
  saving: boolean;
  pending: boolean;
  error?: string | null;
  onCancel: () => void;
  onConfirm: (instructions: string) => void;
};

export function TaskPolishDialog({
  worktreeId,
  projectId = "",
  projectContextItemIds = [],
  saving,
  pending,
  error = null,
  onCancel,
  onConfirm,
}: Props) {
  const titleId = useId();
  const descriptionId = useId();
  const instructionsId = useId();
  const [instructions, setInstructions] = useState("");
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
        <h2 id={titleId}>Polish this task?</h2>
        <p className="task-polish-dialog__footnote" id={descriptionId}>
          Starts a new attempt that resumes the existing agent conversation.
          The task returns to awaiting review when the new attempt finishes.
        </p>
        <div className="field">
          <FieldLabel
            id={`${instructionsId}-label`}
            htmlFor={instructionsId}
            requirement="required"
          >
            Instructions
          </FieldLabel>
          <div className="task-create-editor-shell">
            <RichPromptEditor
              id={instructionsId}
              value={instructions}
              onChange={setInstructions}
              disabled={controlsDisabled}
              placeholder={
                promptProjectContext
                  ? "What should the agent change? Type @ for a repo file, # for project context…"
                  : "What should the agent change? Type @ to mention a repo file…"
              }
              worktreeId={worktreeId?.trim() || undefined}
              projectContext={promptProjectContext ?? undefined}
            />
          </div>
        </div>
        <MutationErrorBanner error={error} className="task-polish-dialog__err" />
        <div className="row stack-row-actions">
          <button
            type="button"
            className="secondary"
            onClick={onCancel}
            disabled={saving}
          >
            Cancel
          </button>
          <button
            type="button"
            className="primary"
            disabled={!canSubmit}
            onClick={() => onConfirm(instructions)}
          >
            {pending ? "Queueing…" : "Polish"}
          </button>
        </div>
      </section>
    </Modal>
  );
}
