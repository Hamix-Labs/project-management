import { useCallback, useRef, type RefObject } from "react";
import type { TaskCreateModalProps } from "../task-create-modal/taskCreateModalProps";
import type { TaskCreateModalPresentation } from "../task-create-modal/taskCreateModalPresentation";
import type { TestScenario } from "@/tasks/test-scenarios";
import type { DraftAssistSnapshot } from "@/types/draftAssist";
import { DraftAssistProvider } from "../draft-assist";
import { TaskComposeFormShell } from "./TaskComposeFormShell";

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

/** Routed compose form (ADR-0100) — redesigned page chrome with draft-assist. */
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

  return (
    <DraftAssistProvider
      getSnapshot={getSnapshot}
      worktreeId={git.worktreeId.trim() || undefined}
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
        session={session}
        essentials={essentials}
        prompt={prompt}
        criteria={criteria}
        git={git}
        execution={execution}
        actions={actions}
        appTimezone={appTimezone}
      />
    </DraftAssistProvider>
  );
}
