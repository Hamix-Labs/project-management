import { useEffect, useState, type FormEvent } from "react";
import { FieldLabel } from "@/shared/FieldLabel";
import { Modal } from "@/shared/Modal";
import { RichPromptEditor } from "@/components/rich-prompt";
import { promptHasVisibleContent } from "@/lib/promptFormat";
import { type ProjectContextItem } from "@/types";
import {
  MAX_PROJECT_CONTEXT_DESCRIPTION_CHARS,
  validateProjectContextDescription,
  validateProjectContextTag,
} from "./projectContextLimits";
import { ProjectContextTagPicker } from "./ProjectContextTagPicker";

type Props = {
  item: ProjectContextItem;
  existingTags: string[];
  saving: boolean;
  deleting: boolean;
  onSave: (
    id: string,
    patch: {
      tag: string;
      title: string;
      description: string;
      body: string;
      pinned: boolean;
    },
  ) => void;
  onDelete: (id: string) => void;
};

export function ProjectContextItemEditor({
  item,
  existingTags,
  saving,
  deleting,
  onSave,
  onDelete,
}: Props) {
  const [body, setBody] = useState(item.body);
  const [description, setDescription] = useState(item.description);
  const [tag, setTag] = useState(item.tag);
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    setBody(item.body);
  }, [item.body]);

  useEffect(() => {
    setDescription(item.description);
  }, [item.description]);

  useEffect(() => {
    setTag(item.tag);
  }, [item.tag]);

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const nextBody = body.trim();
    if (!promptHasVisibleContent(nextBody)) return;
    const nextDescription = description.trim();
    if (validateProjectContextDescription(nextDescription)) return;
    const nextTag = tag.trim();
    if (validateProjectContextTag(nextTag)) return;
    onSave(item.id, {
      tag: nextTag,
      title: String(form.get("title") ?? "").trim(),
      description: nextDescription,
      body: nextBody,
      pinned: false,
    });
  }

  const descriptionError = validateProjectContextDescription(description);
  const tagError = validateProjectContextTag(tag);

  return (
    <>
      <button
        type="button"
        className="project-context-node-card__action-button"
        onClick={() => setIsOpen(true)}
      >
        Edit node
      </button>
      {isOpen ? (
        <Modal
          onClose={() => setIsOpen(false)}
          labelledBy={`context-edit-title-${item.id}`}
          describedBy={`context-edit-description-${item.id}`}
          size="wide"
          busy={saving || deleting}
          busyLabel={deleting ? "Deleting node..." : "Saving node..."}
          dismissibleWhileBusy
        >
          <form
            className="panel modal-sheet modal-sheet--edit project-context-item-form project-context-item-modal"
            onSubmit={submit}
          >
            <div className="project-context-form__heading">
              <div>
                <h2 id={`context-edit-title-${item.id}`}>Edit node</h2>
                <p id={`context-edit-description-${item.id}`} className="muted">
                  Update this project memory node.
                </p>
              </div>
            </div>
            <ProjectContextTagPicker
              id={`context-tag-${item.id}`}
              value={tag}
              onChange={setTag}
              existingTags={existingTags}
              disabled={saving || deleting}
            />
            <div className="field grow">
              <FieldLabel
                htmlFor={`context-title-${item.id}`}
                requirement="required"
              >
                Title
              </FieldLabel>
              <input
                id={`context-title-${item.id}`}
                name="title"
                defaultValue={item.title}
                required
                aria-required="true"
              />
            </div>
            <div className="field grow">
              <FieldLabel htmlFor={`context-desc-${item.id}`}>
                Short description
              </FieldLabel>
              <textarea
                id={`context-desc-${item.id}`}
                name="description"
                value={description}
                aria-invalid={Boolean(descriptionError)}
                disabled={saving || deleting}
                maxLength={MAX_PROJECT_CONTEXT_DESCRIPTION_CHARS}
                rows={2}
                placeholder="Optional blurb for when to use this memory"
                onChange={(event) => setDescription(event.target.value)}
              />
              {descriptionError ? (
                <p className="pd__inline-error" role="alert">
                  {descriptionError}
                </p>
              ) : null}
            </div>
            <div className="field grow">
              <FieldLabel
                id={`context-body-${item.id}-label`}
                htmlFor={`context-body-${item.id}`}
                requirement="required"
              >
                Body
              </FieldLabel>
              <div className="project-context-editor-shell">
                <RichPromptEditor
                  id={`context-body-${item.id}`}
                  value={body}
                  onChange={setBody}
                  disabled={saving || deleting}
                  placeholder="Write markdown-style context. Type @ to reference a repo file."
                />
              </div>
            </div>
            <div className="row stack-row-actions">
              <button
                type="submit"
                disabled={
                  saving || Boolean(descriptionError) || Boolean(tagError)
                }
              >
                Save item
              </button>
              <button
                type="button"
                className="secondary"
                disabled={deleting}
                onClick={() => onDelete(item.id)}
              >
                Delete
              </button>
              <button
                type="button"
                className="secondary"
                disabled={saving || deleting}
                onClick={() => setIsOpen(false)}
              >
                Cancel
              </button>
            </div>
          </form>
        </Modal>
      ) : null}
    </>
  );
}
