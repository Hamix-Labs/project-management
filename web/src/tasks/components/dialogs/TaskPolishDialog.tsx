import { useEffect, useId, useRef, useState } from "react";
import { Modal } from "@/shared/Modal";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";

type Props = {
  taskTitle: string;
  saving: boolean;
  pending: boolean;
  error?: string | null;
  onCancel: () => void;
  onConfirm: (instructions: string) => void;
};

export function TaskPolishDialog({
  taskTitle,
  saving,
  pending,
  error = null,
  onCancel,
  onConfirm,
}: Props) {
  const titleId = useId();
  const descriptionId = useId();
  const instructionsId = useId();
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [instructions, setInstructions] = useState("");
  const trimmed = instructions.trim();
  const canSubmit = trimmed.length > 0 && !saving && !pending;

  useEffect(() => {
    textareaRef.current?.focus();
  }, []);

  return (
    <Modal
      onClose={onCancel}
      labelledBy={titleId}
      describedBy={descriptionId}
      busy={pending}
      busyLabel="Queueing polish…"
      dismissibleWhileBusy
    >
      <section className="panel confirm-dialog modal-sheet task-polish-dialog">
        <h2 id={titleId}>Polish this task?</h2>
        <p className="confirm-dialog__statement" id={descriptionId}>
          <strong>{taskTitle}</strong>
        </p>
        <p className="confirm-dialog__footnote">
          Starts a new attempt that resumes the existing agent conversation.
          The task returns to awaiting review when polish finishes.
        </p>
        <div className="field">
          <label htmlFor={instructionsId}>Instructions</label>
          <textarea
            ref={textareaRef}
            id={instructionsId}
            rows={5}
            value={instructions}
            onChange={(e) => setInstructions(e.target.value)}
            disabled={saving || pending}
            placeholder="What should the agent change?"
            aria-required
          />
        </div>
        <MutationErrorBanner error={error} className="confirm-dialog__err" />
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
            onClick={() => onConfirm(trimmed)}
          >
            {pending ? "Queueing…" : "Polish"}
          </button>
        </div>
      </section>
    </Modal>
  );
}
