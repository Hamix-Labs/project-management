import { useState, type ReactNode } from "react";
import type { TaskTemplateSummary } from "@/types";
import { FolderGitIcon } from "@/components/icons/FolderGitIcon";
import { ProjectsStackIcon } from "@/components/project/ProjectsStackIcon";
import { TaskListDeleteGlyph, TaskListEditGlyph } from "../../components/task-list/table/TaskListRowActionIcons";
import { QuantityStepper } from "../QuantityStepper";
import { formatTemplateRelativeTime, isTemplateRowActionExcluded } from "../templateUtils";

function ZapIcon() {
  return (
    <svg
      className="template-row__runs-icon"
      width="12"
      height="12"
      viewBox="0 0 12 12"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M6.5 1 3 7h3.5L5.5 11 9 5H5.5L6.5 1Z"
        stroke="currentColor"
        strokeWidth="1.1"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function TemplateBindingChip({ icon, label }: { icon: ReactNode; label: string }) {
  return (
    <span className="template-row__binding-chip">
      {icon}
      <span className="template-row__binding-label">{label}</span>
    </span>
  );
}

type TemplateRowProps = {
  template: TaskTemplateSummary;
  isSelected: boolean;
  instanceCount: number;
  isDeleting: boolean;
  isExiting: boolean;
  rowDisabled: boolean;
  renderNow: Date;
  projectLabel: string | null;
  repositoryLabel: string | null;
  onToggleSelected: (id: string) => void;
  onInstanceCountChange: (id: string, count: number) => void;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
};

export function TemplateRow({
  template,
  isSelected,
  instanceCount,
  isDeleting,
  isExiting,
  rowDisabled,
  renderNow,
  projectLabel,
  repositoryLabel,
  onToggleSelected,
  onInstanceCountChange,
  onEdit,
  onDelete,
}: TemplateRowProps) {
  const [confirmingDelete, setConfirmingDelete] = useState(false);
  const lastEdited = template.updated_at || template.created_at;
  const relative = formatTemplateRelativeTime(lastEdited, renderNow);
  const runsLabel = `${template.instantiate_count} run${template.instantiate_count === 1 ? "" : "s"}`;
  const bindingParts = [projectLabel, repositoryLabel].filter(Boolean) as string[];
  const hasSubline = Boolean(lastEdited && relative) || bindingParts.length > 0;

  return (
    <li
      className={[
        "template-row",
        isSelected ? "template-row--selected" : "",
        isExiting ? "template-row--exit" : "",
        rowDisabled ? "template-row--disabled" : "template-row--interactive",
      ]
        .filter(Boolean)
        .join(" ")}
      data-selected={isSelected || undefined}
      onClick={(e) => {
        if (rowDisabled || isTemplateRowActionExcluded(e.target)) return;
        onToggleSelected(template.id);
      }}
      onKeyDown={(e) => {
        if (rowDisabled || isTemplateRowActionExcluded(e.target)) return;
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onToggleSelected(template.id);
        }
      }}
      tabIndex={rowDisabled ? undefined : 0}
      aria-label={`Template: ${template.name}`}
      aria-selected={isSelected}
    >
      <div className="template-row__select">
        <input
          type="checkbox"
          className="template-row__checkbox"
          checked={isSelected}
          aria-label={`Select ${template.name}`}
          onChange={() => onToggleSelected(template.id)}
          onClick={(e) => e.stopPropagation()}
        />
      </div>

      <div className="template-row__meta">
        <div className="template-row__title-line">
          <span className="template-row__name" title={template.name}>
            {template.name}
          </span>
          {template.primary_tag ? (
            <span className="template-row__tag-pill">{template.primary_tag}</span>
          ) : null}
        </div>
        {hasSubline ? (
          <div className="template-row__subline">
            {bindingParts.length > 0 ? (
              <span className="template-row__binding" title={bindingParts.join(" · ")}>
                {projectLabel ? (
                  <TemplateBindingChip
                    icon={<ProjectsStackIcon className="template-row__binding-icon" />}
                    label={projectLabel}
                  />
                ) : null}
                {repositoryLabel ? (
                  <TemplateBindingChip
                    icon={<FolderGitIcon className="template-row__binding-icon" />}
                    label={repositoryLabel}
                  />
                ) : null}
              </span>
            ) : null}
            {bindingParts.length > 0 && lastEdited && relative ? (
              <span className="template-row__subline-sep" aria-hidden="true">
                •
              </span>
            ) : null}
            {lastEdited && relative ? (
              <>
                <time dateTime={lastEdited} title={lastEdited}>
                  Updated {relative}
                </time>
                <span className="template-row__subline-sep" aria-hidden="true">
                  •
                </span>
                <span className="template-row__runs">
                  <ZapIcon />
                  {runsLabel}
                </span>
              </>
            ) : null}
          </div>
        ) : null}
      </div>

      <div className="template-row__instances">
        <QuantityStepper
          size="sm"
          value={instanceCount}
          disabled={rowDisabled}
          ariaLabel={`Instances for ${template.name}`}
          onChange={(count) => onInstanceCountChange(template.id, count)}
        />
      </div>

      <div className="template-row__actions">
        {confirmingDelete ? (
          <div className="template-row__delete-confirm">
            <button
              type="button"
              className="secondary template-row__delete-cancel"
              onClick={(e) => {
                e.stopPropagation();
                setConfirmingDelete(false);
              }}
            >
              Cancel
            </button>
            <button
              type="button"
              className="template-row__delete-confirm-btn"
              aria-busy={isDeleting || undefined}
              disabled={rowDisabled}
              onClick={(e) => {
                e.stopPropagation();
                setConfirmingDelete(false);
                onDelete(template.id);
              }}
            >
              Delete
            </button>
          </div>
        ) : (
          <div className="template-row__action-icons">
            <button
              type="button"
              className="task-list-icon-btn task-list-icon-btn--edit"
              aria-label={`Edit template "${template.name}"`}
              disabled={rowDisabled}
              onClick={(e) => {
                e.stopPropagation();
                onEdit(template.id);
              }}
            >
              <TaskListEditGlyph />
            </button>
            <button
              type="button"
              className="task-list-icon-btn task-list-icon-btn--delete"
              aria-label={`Delete template "${template.name}"`}
              disabled={rowDisabled}
              onClick={(e) => {
                e.stopPropagation();
                setConfirmingDelete(true);
              }}
            >
              <TaskListDeleteGlyph />
            </button>
          </div>
        )}
      </div>
    </li>
  );
}
