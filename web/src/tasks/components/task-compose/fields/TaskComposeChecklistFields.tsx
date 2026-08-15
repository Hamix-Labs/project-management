import { FieldRequirementBadge } from "@/shared/FieldLabel";
import type { ChecklistItemDraft } from "@/types";
import { ChecklistVerifyBadge } from "@/tasks/components/task-detail/checklist/ChecklistVerifyBadge";
import { CREATE_CHECKLIST_REQUIRED_MSG } from "@/tasks/task-compose/checklistRequirement";

type Props = {
  checklistHeadingId: string;
  checklistItems: ChecklistItemDraft[];
  /** When `required`, shows the required badge and create-time helper copy. */
  checklistRequirement?: "optional" | "required";
  disabled: boolean;
  /** Section header owns the title when true (create modal). */
  hideSectionHeading?: boolean;
  /** Hide inline New criterion when rendered in section action slot. */
  hideNewCriterionButton?: boolean;
  onOpenNewCriterion: () => void;
  onOpenEditCriterion: (index: number, item: ChecklistItemDraft) => void;
  onRemoveRow: (index: number) => void;
};

export function TaskComposeChecklistFields({
  checklistHeadingId,
  checklistItems,
  checklistRequirement = "optional",
  disabled,
  hideSectionHeading = false,
  hideNewCriterionButton = false,
  onOpenNewCriterion,
  onOpenEditCriterion,
  onRemoveRow,
}: Props) {
  const isEmpty = checklistItems.length === 0;

  return (
    <div className="task-create-checklist">
      {hideSectionHeading ? null : (
        <div className="task-create-checklist-head">
          <div className="field-heading-with-req task-create-checklist-title-row">
            <h3 className="task-create-checklist-heading" id={checklistHeadingId}>
              Done criteria
            </h3>
            <FieldRequirementBadge requirement={checklistRequirement} />
          </div>
          {hideNewCriterionButton ? null : (
            <button
              type="button"
              className="task-detail-add-checklist-btn"
              disabled={disabled}
              onClick={onOpenNewCriterion}
            >
              New criterion
            </button>
          )}
        </div>
      )}

      {isEmpty ? (
        <div
          className="task-create-checklist-empty"
          aria-labelledby={checklistHeadingId}
        >
          <p className="task-create-checklist-empty__text">
            {checklistRequirement === "required"
              ? CREATE_CHECKLIST_REQUIRED_MSG
              : "No criteria yet."}
          </p>
        </div>
      ) : (
        <div className="task-checklist-surface">
          <ul
            className="task-checklist-list task-checklist-list--grouped"
            aria-labelledby={checklistHeadingId}
          >
            {checklistItems.map((item, index) => {
              const commandCount = item.verify_commands?.length ?? 0;
              const canEditRow = !disabled;
              return (
                <li
                  key={`${index}-${item.text}`}
                  className={
                    canEditRow
                      ? "task-checklist-row task-checklist-row--interactive"
                      : "task-checklist-row"
                  }
                  onClick={(event) => {
                    if (!canEditRow) return;
                    if ((event.target as HTMLElement).closest("button")) return;
                    onOpenEditCriterion(index, item);
                  }}
                >
                  <div className="task-checklist-row-primary">
                    <CriterionMarkerIcon />
                    <span className="task-checklist-text" title={item.text}>
                      {item.text}
                    </span>
                  </div>
                  <div className="task-checklist-row-trailing">
                    {commandCount > 0 ? (
                      <ChecklistVerifyBadge count={commandCount} />
                    ) : null}
                    <div className="task-checklist-row-actions">
                      <button
                        type="button"
                        className="task-detail-checklist-edit"
                        disabled={disabled}
                        onClick={() => onOpenEditCriterion(index, item)}
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        className="task-detail-checklist-remove"
                        disabled={disabled}
                        aria-label="Remove"
                        onClick={() => onRemoveRow(index)}
                      >
                        <CriterionTrashIcon />
                      </button>
                    </div>
                  </div>
                </li>
              );
            })}
          </ul>
        </div>
      )}
    </div>
  );
}

function CriterionTrashIcon() {
  return (
    <svg
      width={16}
      height={16}
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <path
        d="M3 6h18"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinecap="round"
      />
      <path
        d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M10 11v6M14 11v6"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinecap="round"
      />
    </svg>
  );
}

/** Decorative circle-check — compose criteria are not yet satisfied. */
function CriterionMarkerIcon() {
  return (
    <span className="compose-criteria__check" aria-hidden="true">
      <svg
        width={20}
        height={20}
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <circle
          cx={12}
          cy={12}
          r={10}
          stroke="currentColor"
          strokeWidth={1.75}
        />
        <path
          d="M8 12.5 10.8 15.2 16 9.8"
          stroke="currentColor"
          strokeWidth={1.75}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </span>
  );
}
