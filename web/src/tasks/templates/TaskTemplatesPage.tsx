import { useNavigate } from "react-router-dom";
import { useMemo } from "react";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { useGlobalRepositories } from "@/hooks/useGlobalRepositories";
import { useProjects } from "@/hooks/useProjects";
import { useTasksAppContext } from "../app/TasksAppProvider";
import { TemplateBatchBar } from "./components/TemplateBatchBar";
import { TemplateFunctionBindModal } from "./components/TemplateFunctionBindModal";
import { TemplatePageBody } from "./components/TemplatePageBody";
import { TemplatePageHeader } from "./components/TemplatePageHeader";
import { TemplateToolbar } from "./components/TemplateToolbar";
import { repositoryDisplayName } from "@/lib/repositoryDisplayName";
import { useTaskTemplatesPageModel } from "./useTaskTemplatesPageModel";

export function TaskTemplatesPage() {
  const app = useTasksAppContext();
  const navigate = useNavigate();
  const model = useTaskTemplatesPageModel(app, navigate);
  const projectsQuery = useProjects();
  const repositoriesQuery = useGlobalRepositories();
  useDocumentTitle("Task templates");

  const projectNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const project of projectsQuery.data?.projects ?? []) {
      map.set(project.id, project.name);
    }
    return map;
  }, [projectsQuery.data]);

  const repositoryNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const repo of repositoriesQuery.data ?? []) {
      map.set(repo.id, repositoryDisplayName(repo.path) || repo.path);
    }
    return map;
  }, [repositoriesQuery.data]);

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
        projectNameById={projectNameById}
        repositoryNameById={repositoryNameById}
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
        onBatchDefaultCountChange={model.setBatchDefaultCountAndApply}
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
