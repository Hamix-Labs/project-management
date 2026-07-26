import { useNavigate } from "react-router-dom";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { useTasksAppContext } from "../app/TasksAppProvider";
import { TemplateBatchBar } from "./components/TemplateBatchBar";
import { TemplateFunctionBindModal } from "./components/TemplateFunctionBindModal";
import { TemplatePageBody } from "./components/TemplatePageBody";
import { TemplatePageHeader } from "./components/TemplatePageHeader";
import { TemplateTagFilters } from "./components/TemplateTagFilters";
import { TemplateToolbar } from "./components/TemplateToolbar";
import { clampInstanceCount } from "./templateUtils";
import { useTaskTemplatesPageModel } from "./useTaskTemplatesPageModel";

export function TaskTemplatesPage() {
  const app = useTasksAppContext();
  const navigate = useNavigate();
  const model = useTaskTemplatesPageModel(app, navigate);
  useDocumentTitle("Task templates");

  const rowDisabled =
    app.loadTemplatePending || app.deleteTemplatePending || app.instantiateTemplatesPending;

  return (
    <section className="templates-page-card task-detail-content--enter">
      <TemplatePageHeader onNewTemplate={() => app.openTemplateCreateModal()} />
      <div className="templates-page-toolbar-section">
        <TemplateToolbar
          searchInput={model.searchInput}
          sort={model.sort}
          onSearchChange={model.setSearchInput}
          onSortChange={model.setSort}
        />
        <TemplateTagFilters
          activeTag={model.activeTag}
          dynamicTags={model.dynamicTags}
          onTagChange={model.setActiveTag}
        />
      </div>
      {model.batchError ? (
        <div className="err templates-page-error" role="alert">
          <p>{model.batchError}</p>
        </div>
      ) : null}
      <TemplatePageBody
        loading={model.loading}
        showSkeleton={model.showSkeleton}
        error={model.error}
        onRetry={() => void model.templatesQuery.refetch()}
        templates={model.templates}
        hasFilters={model.hasFilters}
        onClearFilters={model.clearFilters}
        onNewTemplate={() => app.openTemplateCreateModal()}
        selectedIds={model.selectedIds}
        instanceCounts={model.instanceCounts}
        allSelected={model.allSelected}
        someSelected={model.someSelected}
        deletingTemplateId={model.deletingTemplateId}
        exitingTemplateIds={model.exitingTemplateIds}
        rowDisabled={rowDisabled}
        renderNow={model.renderNow}
        selectedCount={model.selectedCount}
        onToggleSelectAll={model.toggleSelectAll}
        onToggleSelected={model.toggleSelected}
        onInstanceCountChange={model.setInstanceCountForTemplate}
        onEdit={(id) => void app.editTemplateByID(id)}
        onDelete={(id) => void model.deleteTemplate(id)}
      />
      <TemplateBatchBar
        selectedCount={model.selectedCount}
        totalTaskCount={model.totalTaskCount}
        batchDefaultCount={model.batchDefaultCount}
        instantiatePending={app.instantiateTemplatesPending}
        onClear={model.clearSelection}
        onBatchDefaultCountChange={(count) => model.setBatchDefaultCount(clampInstanceCount(count))}
        onApplyToAll={model.applyBatchDefaultToSelected}
        onCreate={() => void model.runBatchCreate()}
      />
      {model.bindDrafts ? (
        <TemplateFunctionBindModal
          drafts={model.bindDrafts}
          pending={app.instantiateTemplatesPending}
          error={model.bindError}
          onChange={model.setBindDrafts}
          onCancel={model.closeBindModal}
          onConfirm={() => void model.confirmBindAndCreate()}
        />
      ) : null}
    </section>
  );
}
