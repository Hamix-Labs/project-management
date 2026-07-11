import { FieldLabel } from "@/shared/FieldLabel";
import { Modal } from "@/shared/Modal";
import { RichPromptEditor } from "@/components/rich-prompt";
import { ProjectContextKindPicker } from "./ProjectContextKindPicker";
import type { FormEvent } from "react";

type Props = {
  open: boolean;
  onClose: () => void;
  isPending: boolean;
  newNodeBody: string;
  newNodeEditorKey: number;
  onBodyChange: (body: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

export function ProjectContextAddNodeModal({
  open,
  onClose,
  isPending,
  newNodeBody,
  newNodeEditorKey,
  onBodyChange,
  onSubmit,
}: Props) {
  if (!open) return null;

  return (
    <Modal
      onClose={onClose}
      labelledBy="project-context-add-node-title"
      describedBy="project-context-add-node-desc"
      size="wide"
      busy={isPending}
      busyLabel="Adding node..."
    >
      <form
        className="panel modal-sheet modal-sheet--edit project-context-form project-context-node-modal"
        onSubmit={onSubmit}
      >
        <div className="project-context-form__heading">
          <div>
            <h2 id="project-context-add-node-title">Add memory node</h2>
            <p id="project-context-add-node-desc" className="muted">
              Nodes are project-owned facts, decisions, constraints, or custom
              context. All fields are required.
            </p>
          </div>
        </div>
        <ProjectContextKindPicker
          idPrefix="project-context-kind"
          disabled={isPending}
        />
        <div className="field grow">
          <FieldLabel htmlFor="project-context-title" requirement="required">
            Title
          </FieldLabel>
          <input
            id="project-context-title"
            name="title"
            required
            aria-required="true"
          />
        </div>
        <div className="field grow">
          <FieldLabel
            id="project-context-body-label"
            htmlFor="project-context-body"
            requirement="required"
          >
            Body
          </FieldLabel>
          <div className="project-context-editor-shell">
            <RichPromptEditor
              key={newNodeEditorKey}
              id="project-context-body"
              value={newNodeBody}
              onChange={onBodyChange}
              disabled={isPending}
              placeholder="Write markdown-style context. Type @ to reference a repo file."
            />
          </div>
        </div>
        <div className="row stack-row-actions">
          <button type="submit" disabled={isPending}>
            {isPending ? "Adding..." : "Add node"}
          </button>
          <button
            type="button"
            className="secondary"
            disabled={isPending}
            onClick={onClose}
          >
            Cancel
          </button>
        </div>
      </form>
    </Modal>
  );
}
