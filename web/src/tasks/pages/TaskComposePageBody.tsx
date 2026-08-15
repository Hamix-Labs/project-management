import { Suspense, useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { getTask } from "@/api/tasks.read";
import { RegisterRepositoryFirstPrompt } from "@/components/RegisterRepositoryFirstPrompt";
import { useAppTimezone } from "@/shared/time/appTimezone";
import { useTasksAppContext } from "../app/TasksAppProvider";
import { DraftResumeModal } from "../components/draft-resume";
import { TaskComposeForm } from "../components/task-compose/TaskComposeForm";
import { TaskComposeLayout } from "../components/task-compose/TaskComposeLayout";
import type { TestScenario } from "../test-scenarios";
import { buildTaskComposeFormBundle } from "./buildTaskComposeFormBundle";
import { CreateModalChunkFallback } from "./CreateModalChunkFallback";
import { composeBackTo, type ComposeMode } from "./composeMode";

export function TaskComposePageBody({ mode }: { mode: ComposeMode }) {
  const app = useTasksAppContext();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const appTimezone = useAppTimezone();
  /** True after this mount has shown compose/draft-picker/repo-setup. */
  const sawComposeSurface = useRef(false);
  const [seedError, setSeedError] = useState<string | null>(null);
  const [scenariosOpen, setScenariosOpen] = useState(false);
  const scenariosTriggerRef = useRef<HTMLButtonElement>(null);
  const backTo = composeBackTo(mode);

  /**
   * Seed compose from the route. Must re-run after React Strict Mode's
   * setup→cleanup→setup: the cleanup closes the form and bumps the entry
   * request id, so a one-shot "seeded" ref would leave the page on Loading…
   * forever.
   */
  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      try {
        if (mode.kind === "task-create") {
          const project = searchParams.get("project")?.trim() || undefined;
          const repository =
            searchParams.get("repository")?.trim() || undefined;
          const worktree = searchParams.get("worktree")?.trim() || undefined;
          const draft = searchParams.get("draft")?.trim();
          const lockGit = searchParams.get("lock_git") === "1";
          const lockProject = searchParams.get("lock_project") === "1";
          if (draft) {
            await app.resumeDraftByID(draft);
          } else if (project) {
            await app.openCreateModal({
              projectID: project,
              repositoryID: repository,
              worktreeID: worktree,
              lockGitAssignment: lockGit,
              lockProjectAssignment: lockProject,
            });
          } else {
            await app.openCreateModal();
          }
        } else if (mode.kind === "template-create") {
          app.openTemplateCreateModal();
        } else if (mode.kind === "template-edit") {
          await app.editTemplateByID(mode.templateId);
        } else if (mode.kind === "task-edit") {
          const task = await getTask(mode.taskId);
          if (cancelled) return;
          app.openEdit(task);
        }
      } catch (err) {
        if (!cancelled) {
          setSeedError(
            err instanceof Error ? err.message : "Could not open compose form.",
          );
        }
      }
    };

    void run();
    return () => {
      cancelled = true;
      // Strict Mode runs cleanup between double-invokes; don't treat that
      // close as "user left compose" or the leave-effect will bounce home.
      sawComposeSurface.current = false;
      app.closeEdit();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [mode.kind, "taskId" in mode ? mode.taskId : "", "templateId" in mode ? mode.templateId : "", searchParams]);

  // Leave the route once compose has been shown and then closed (create
  // success, Cancel, etc.). Ignore the Strict Mode cleanup flicker by only
  // arming after a surface was visible on this mount.
  useEffect(() => {
    if (
      app.createModalOpen ||
      app.draftPickerOpen ||
      app.repositorySetupPromptOpen
    ) {
      sawComposeSurface.current = true;
      return;
    }
    if (!sawComposeSurface.current) return;
    sawComposeSurface.current = false;
    navigate(backTo, { replace: true });
  }, [
    app.createModalOpen,
    app.draftPickerOpen,
    app.repositorySetupPromptOpen,
    navigate,
    backTo,
  ]);

  const leave = () => {
    app.closeEdit();
    navigate(backTo);
  };

  if (seedError) {
    return (
      <TaskComposeLayout title="Compose" backTo={backTo}>
        <div className="err error-banner" role="alert">
          {seedError}
        </div>
      </TaskComposeLayout>
    );
  }

  if (app.repositorySetupPromptOpen) {
    return (
      <RegisterRepositoryFirstPrompt
        open
        onClose={leave}
        onRegister={() => {
          app.setRepositorySetupPromptOpen(false);
          navigate("/repositories?register=1");
        }}
      />
    );
  }

  // Retry-from-entry-hint can still open the picker here; New task on the
  // list is the primary gate and /tasks/new no longer auto-opens it.
  if (app.draftPickerOpen) {
    return (
      <TaskComposeLayout title="Resume a draft?" backTo={backTo}>
        <DraftResumeModal
          drafts={app.taskDrafts}
          onClose={leave}
          onStartFresh={() => void app.startFreshDraft()}
          onResume={(id) => {
            void app.resumeDraftByID(id).catch(() => {});
          }}
          loading={app.draftListLoading}
          loadError={app.draftListError}
          onRetryLoad={() => {
            void app.retryDraftList();
          }}
          resumePending={app.resumeDraftPending}
          resumeError={app.resumeDraftError}
        />
      </TaskComposeLayout>
    );
  }

  if (!app.createModalOpen) {
    return (
      <TaskComposeLayout title="Loading…" backTo={backTo}>
        <CreateModalChunkFallback onClose={leave} />
      </TaskComposeLayout>
    );
  }

  const { presentation, props } = buildTaskComposeFormBundle(app, {
    leave,
    appTimezone,
  });

  return (
    <Suspense fallback={<CreateModalChunkFallback onClose={leave} />}>
      <>
        {app.createEntryDraftErrorHint ? (
          <div className="err error-banner" role="alert">
            <span className="error-banner__text">
              Saved drafts are unavailable right now, so a fresh task form was
              opened.
            </span>
            <button
              type="button"
              className="secondary"
              onClick={() => {
                void app.retryCreateEntryDraftLoad();
              }}
            >
              Retry loading drafts
            </button>
          </div>
        ) : null}
        <TaskComposeForm
          {...props}
          presentation={presentation}
          backTo={backTo}
          scenariosOpen={scenariosOpen}
          scenariosTriggerRef={scenariosTriggerRef}
          onToggleScenarios={() => setScenariosOpen((o) => !o)}
          onScenarioPicked={(scenario: TestScenario) => {
            app.applyTestScenario?.(scenario);
            setScenariosOpen(false);
            scenariosTriggerRef.current?.focus();
          }}
          onCloseScenarios={() => setScenariosOpen(false)}
        />
      </>
    </Suspense>
  );
}
