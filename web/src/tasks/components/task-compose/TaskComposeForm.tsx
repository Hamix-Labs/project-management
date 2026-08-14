import { useCallback, useRef, type RefObject } from "react";
import { TaskCreateModalActionFooter } from "../task-create-modal/TaskCreateModalActionFooter";
import { TaskCreateModalFormBody } from "../task-create-modal/TaskCreateModalFormBody";
import { TaskCreateModalMutationErrors } from "../task-create-modal/TaskCreateModalMutationErrors";
import { TestScenariosPopover } from "../task-create-modal/TestScenariosPopover";
import { TestScenariosTrigger } from "../task-create-modal/TestScenariosTrigger";
import type { TaskCreateModalProps } from "../task-create-modal/taskCreateModalProps";
import type { TaskCreateModalPresentation } from "../task-create-modal/taskCreateModalPresentation";
import type { TestScenario } from "@/tasks/test-scenarios";
import {
  DraftAssistProvider,
  DraftAssistThread,
  useDraftAssistContext,
} from "../draft-assist";
import type { DraftAssistSnapshot } from "@/types/draftAssist";
import { TaskComposeLayout } from "./TaskComposeLayout";

type Props = TaskCreateModalProps & {
  presentation: TaskCreateModalPresentation;
  backTo: string;
  backLabel?: string;
  scenariosOpen: boolean;
  scenariosTriggerRef: RefObject<HTMLButtonElement>;
  onToggleScenarios: () => void;
  onScenarioPicked: (scenario: TestScenario) => void;
  onCloseScenarios: () => void;
};

/** Routed compose form (ADR-0100) — same fields as the former modal, page chrome. */
export function TaskComposeForm({
  presentation,
  session,
  essentials,
  prompt,
  criteria,
  git,
  execution,
  actions,
  appTimezone,
  backTo,
  backLabel,
  scenariosOpen,
  scenariosTriggerRef,
  onToggleScenarios,
  onScenarioPicked,
  onCloseScenarios,
}: Props) {
  const editingTaskId = session.editingTaskId ?? null;
  const editingTemplateId = session.editingTemplateId ?? null;
  const editingTaskRunner = session.editingTaskRunner ?? "";
  const createError = session.createError ?? null;
  const createFormError = session.createFormError ?? null;
  const patchError = session.patchError ?? null;
  const formError = session.formError ?? null;

  const createHint =
    !presentation.isEdit && !presentation.isTemplateMode
      ? "Define the work, then hand it off to your agent."
      : null;
  const draftSubtitle =
    presentation.showDraftStatus && session.draftSaveLabel
      ? session.draftSaveLabel
      : session.draftSaveError
        ? "Draft autosave failed"
        : createHint;

  // Refs let the DraftAssistProvider read the latest form state without
  // re-instantiating provider callbacks on every keystroke.
  const promptRef = useRef(prompt.prompt);
  promptRef.current = prompt.prompt;
  const titleRef = useRef(essentials.title);
  titleRef.current = essentials.title;
  const priorityRef = useRef(essentials.priority);
  priorityRef.current = essentials.priority;
  const projectIdRef = useRef(git.projectId);
  projectIdRef.current = git.projectId;
  const tagsCsvRef = useRef(criteria.tagsCsv);
  tagsCsvRef.current = criteria.tagsCsv;
  const cursorModelRef = useRef(execution.taskCursorModel);
  cursorModelRef.current = execution.taskCursorModel;

  const getSnapshot = useCallback((): DraftAssistSnapshot => {
    const snap: DraftAssistSnapshot = {
      title: titleRef.current,
      prompt: promptRef.current,
      priority: String(priorityRef.current ?? ""),
    };
    if (projectIdRef.current) snap.project_id = projectIdRef.current;
    const csv = tagsCsvRef.current.trim();
    if (csv !== "") {
      snap.tags = csv
        .split(",")
        .map((t) => t.trim())
        .filter((t) => t !== "");
    }
    if (cursorModelRef.current) snap.cursor_model = cursorModelRef.current;
    return snap;
  }, []);

  const onPromptChangeRef = useRef(prompt.onPromptChange);
  onPromptChangeRef.current = prompt.onPromptChange;
  const applyPromptPatch = useCallback((next: string) => {
    onPromptChangeRef.current(next);
    promptRef.current = next;
  }, []);
  const getPromptSnapshot = useCallback(() => promptRef.current, []);

  const worktreeIdForAssist = git.worktreeId.trim() || undefined;

  return (
    <DraftAssistProvider
      getSnapshot={getSnapshot}
      worktreeId={worktreeIdForAssist}
      onApplyPromptPatch={applyPromptPatch}
      getPromptSnapshot={getPromptSnapshot}
    >
      <TaskComposeFormShell
        presentation={presentation}
        draftSubtitle={draftSubtitle}
        backTo={backTo}
        backLabel={backLabel}
        scenariosOpen={scenariosOpen}
        scenariosTriggerRef={scenariosTriggerRef}
        onToggleScenarios={onToggleScenarios}
        onScenarioPicked={onScenarioPicked}
        onCloseScenarios={onCloseScenarios}
        editingTaskId={editingTaskId}
        editingTemplateId={editingTemplateId}
        editingTaskRunner={editingTaskRunner}
        session={session}
        essentials={essentials}
        prompt={prompt}
        criteria={criteria}
        git={git}
        execution={execution}
        actions={actions}
        appTimezone={appTimezone}
        createError={createError}
        createFormError={createFormError}
        patchError={patchError}
        formError={formError}
      />
    </DraftAssistProvider>
  );
}

// Kept as its own inner component so `useDraftAssistContext` reads the
// provider we just mounted. The wiring split also keeps the outer
// component's snapshot refs from re-triggering child renders.

type ShellProps = Omit<
  Props,
  "presentation" | "backTo" | "backLabel"
> & {
  presentation: TaskCreateModalPresentation;
  draftSubtitle: string | null;
  backTo: string;
  backLabel?: string;
  editingTaskId: string | null;
  editingTemplateId: string | null;
  editingTaskRunner: string;
  createError: Error | null;
  createFormError: string | null;
  patchError: string | null;
  formError: string | null;
};

function TaskComposeFormShell({
  presentation,
  draftSubtitle,
  backTo,
  backLabel,
  scenariosOpen,
  scenariosTriggerRef,
  onToggleScenarios,
  onScenarioPicked,
  onCloseScenarios,
  editingTaskId,
  editingTemplateId,
  editingTaskRunner,
  session,
  essentials,
  prompt,
  criteria,
  git,
  execution,
  actions,
  appTimezone,
  createError,
  createFormError,
  patchError,
  formError,
}: ShellProps) {
  const assist = useDraftAssistContext();
  const assistNode = assist.active ? <DraftAssistThread /> : null;
  const autonomyDisabled = execution.autonomyDisabled ?? false;

  return (
    <>
      <TaskComposeLayout
        title={presentation.modalTitle}
        subtitle={draftSubtitle}
        backTo={backTo}
        backLabel={backLabel}
        topActions={
          presentation.showTestScenarios ? (
            <TestScenariosTrigger
              ref={scenariosTriggerRef}
              open={scenariosOpen}
              disabled={presentation.disabled}
              onToggle={onToggleScenarios}
            />
          ) : null
        }
        assist={assistNode}
      >
        <section
          className="panel modal-sheet modal-sheet--edit task-create-modal-sheet task-create"
          aria-labelledby="task-compose-page-title"
        >
          {presentation.isTaskEdit && editingTaskId ? (
            <p
              className="muted stack-tight-zero task-create-modal-task-id"
              id="task-edit-modal-description"
            >
              <code>{editingTaskId}</code>
            </p>
          ) : null}
          <form
            className="task-create-modal-form task-create-form"
            onSubmit={actions.onSubmit}
          >
            <TaskCreateModalFormBody
              presentation={presentation}
              editingTaskId={editingTaskId}
              editingTemplateId={editingTemplateId}
              editingTaskRunner={editingTaskRunner}
              onComposeStatusChange={session.onComposeStatusChange}
              essentials={essentials}
              prompt={prompt}
              criteria={criteria}
              git={git}
              execution={{ ...execution, autonomyDisabled }}
              appTimezone={appTimezone}
            />

            <div className="task-compose-page__errors">
              <TaskCreateModalMutationErrors
                isTaskEdit={presentation.isTaskEdit}
                createFormError={createFormError}
                createError={createError}
                formError={formError}
                patchError={patchError}
              />
            </div>

            <footer className="task-compose-page__footer task-create-modal-footer">
              <TaskCreateModalActionFooter
                presentation={presentation}
                title={essentials.title}
                priority={essentials.priority}
                checklistItems={criteria.checklistItems}
                repositoryId={git.repositoryId}
                draftSaving={session.draftSaving}
                onClose={actions.onClose}
                onSaveDraft={actions.onSaveDraft}
              />
            </footer>
          </form>
        </section>
      </TaskComposeLayout>

      {scenariosOpen && presentation.showTestScenarios ? (
        <TestScenariosPopover
          anchor={scenariosTriggerRef.current}
          onPick={onScenarioPicked}
          onClose={onCloseScenarios}
        />
      ) : null}
    </>
  );
}
