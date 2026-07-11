import { FieldLabel } from "@/shared/FieldLabel";
import { Modal } from "@/shared/Modal";
import { CustomSelect, type CustomSelectOption } from "@/components/custom-select";
import { RichPromptEditor } from "@/components/rich-prompt";
import type { ProjectContextRelation } from "@/types";
import type { FormEvent } from "react";

type Props = {
  open: boolean;
  onClose: () => void;
  isPending: boolean;
  memoryOptions: CustomSelectOption[];
  relationOptions: CustomSelectOption[];
  strengthOptions: CustomSelectOption[];
  newEdgeSourceID: string;
  newEdgeTargetID: string;
  newEdgeRelation: ProjectContextRelation;
  newEdgeStrength: string;
  newEdgeNote: string;
  newEdgeEditorKey: number;
  onSourceChange: (id: string) => void;
  onTargetChange: (id: string) => void;
  onRelationChange: (relation: ProjectContextRelation) => void;
  onStrengthChange: (strength: string) => void;
  onNoteChange: (note: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

export function ProjectContextAddEdgeModal({
  open,
  onClose,
  isPending,
  memoryOptions,
  relationOptions,
  strengthOptions,
  newEdgeSourceID,
  newEdgeTargetID,
  newEdgeRelation,
  newEdgeStrength,
  newEdgeNote,
  newEdgeEditorKey,
  onSourceChange,
  onTargetChange,
  onRelationChange,
  onStrengthChange,
  onNoteChange,
  onSubmit,
}: Props) {
  if (!open) return null;

  return (
    <Modal
      onClose={onClose}
      labelledBy="project-context-add-edge-title"
      describedBy="project-context-add-edge-desc"
      size="wide"
      busy={isPending}
      busyLabel="Adding connection..."
    >
      <form
        className="panel modal-sheet modal-sheet--edit project-context-form project-context-edge-modal"
        onSubmit={onSubmit}
      >
        <div className="project-context-form__heading">
          <div>
            <h2 id="project-context-add-edge-title">Add connection</h2>
            <p id="project-context-add-edge-desc" className="muted">
              Link two memory nodes when the relationship helps future work.
            </p>
          </div>
        </div>
        <div className="project-context-edge-grid">
          <CustomSelect
            id="project-context-edge-source"
            label="From"
            value={newEdgeSourceID}
            options={memoryOptions}
            onChange={onSourceChange}
          />
          <CustomSelect
            id="project-context-edge-target"
            label="To"
            value={newEdgeTargetID}
            options={memoryOptions}
            onChange={onTargetChange}
          />
          <CustomSelect
            id="project-context-edge-relation"
            label="Relation"
            value={newEdgeRelation}
            options={relationOptions}
            onChange={(value) => onRelationChange(value as ProjectContextRelation)}
          />
          <CustomSelect
            id="project-context-edge-strength"
            label="Strength"
            value={newEdgeStrength}
            options={strengthOptions}
            onChange={onStrengthChange}
          />
          <div className="field grow project-context-edge-note">
            <FieldLabel
              id="project-context-edge-note-label"
              htmlFor="project-context-edge-note"
            >
              Note
            </FieldLabel>
            <div className="project-context-editor-shell">
              <RichPromptEditor
                key={newEdgeEditorKey}
                id="project-context-edge-note"
                value={newEdgeNote}
                onChange={onNoteChange}
                disabled={isPending}
                placeholder="Why does this link matter? Type @ for files."
              />
            </div>
          </div>
        </div>
        <div className="row stack-row-actions">
          <button type="submit" disabled={isPending}>
            {isPending ? "Connecting..." : "Add connection"}
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
