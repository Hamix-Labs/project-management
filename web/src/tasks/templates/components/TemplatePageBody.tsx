import { TaskDraftsListSkeleton } from "../../components/skeletons";
import type { TaskTemplateSummary } from "@/types";
import { TemplateEmptyState } from "./TemplateEmptyState";
import { TemplateListHeader } from "./TemplateListHeader";
import { TemplateRow } from "./TemplateRow";

type TemplatePageBodyProps = {
  loading: boolean;
  showSkeleton: boolean;
  error: string | null;
  onRetry: () => void;
  templates: TaskTemplateSummary[];
  hasFilters: boolean;
  onClearFilters: () => void;
  onNewTemplate: () => void;
  selectedIds: string[];
  instanceCounts: Record<string, number>;
  allSelected: boolean;
  someSelected: boolean;
  deletingTemplateId: string | null;
  exitingTemplateIds: string[];
  rowDisabled: boolean;
  renderNow: Date;
  selectedCount: number;
  onToggleSelectAll: () => void;
  onToggleSelected: (id: string) => void;
  onInstanceCountChange: (id: string, count: number) => void;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
};

export function TemplatePageBody(props: TemplatePageBodyProps) {
  if (props.loading && props.showSkeleton) return <TaskDraftsListSkeleton />;

  if (!props.loading && props.error) {
    return (
      <div className="err templates-page-error" role="alert">
        <p>{props.error}</p>
        <button type="button" className="secondary" onClick={props.onRetry}>
          Try again
        </button>
      </div>
    );
  }

  if (!props.loading && !props.error && props.templates.length === 0) {
    return (
      <TemplateEmptyState
        hasFilters={props.hasFilters}
        onClearFilters={props.onClearFilters}
        onNewTemplate={props.onNewTemplate}
      />
    );
  }

  if (!props.loading && !props.error) {
    return (
      <>
        <TemplateListHeader
          allSelected={props.allSelected}
          someSelected={props.someSelected}
          onToggleSelectAll={props.onToggleSelectAll}
        />
        <ul className="templates-list-rows" aria-label="Task templates">
          {props.templates.map((template) => (
            <TemplateRow
              key={template.id}
              template={template}
              isSelected={props.selectedIds.includes(template.id)}
              instanceCount={props.instanceCounts[template.id] ?? 1}
              isDeleting={props.deletingTemplateId === template.id}
              isExiting={props.exitingTemplateIds.includes(template.id)}
              rowDisabled={
                props.rowDisabled || props.exitingTemplateIds.includes(template.id)
              }
              renderNow={props.renderNow}
              onToggleSelected={props.onToggleSelected}
              onInstanceCountChange={props.onInstanceCountChange}
              onEdit={props.onEdit}
              onDelete={props.onDelete}
            />
          ))}
        </ul>
        {props.selectedCount > 0 ? (
          <div className="templates-batch-spacer" aria-hidden="true" />
        ) : null}
      </>
    );
  }

  return null;
}
