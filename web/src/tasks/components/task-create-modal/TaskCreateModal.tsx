import { useRef, useState } from "react";
import type { TestScenario } from "@/tasks/test-scenarios";
import { taskCreateModalBusyLabel } from "./taskCreateModalBusyLabel";
import { resolveTaskCreateModalPresentation } from "./taskCreateModalPresentation";
import type { TaskCreateModalProps } from "./taskCreateModalProps";
import { TaskCreateModalShell } from "./TaskCreateModalShell";

export type { TaskCreateModalProps };

export function TaskCreateModal(props: TaskCreateModalProps) {
  const { session, actions } = props;
  const editingTaskId = session.editingTaskId ?? null;
  const composeTarget = session.composeTarget ?? "task";
  const composeOperation = session.composeOperation ?? "create";
  const { onApplyTestScenario } = actions;

  const presentation = resolveTaskCreateModalPresentation({
    editingTaskId,
    composeTarget,
    composeOperation,
    composeStatus: session.composeStatus,
    pending: session.pending,
    saving: session.saving,
    patchPending: session.patchPending ?? false,
    draftSaveLabel: session.draftSaveLabel,
    onApplyTestScenario,
  });

  const [scenariosOpen, setScenariosOpen] = useState(false);
  const scenariosTriggerRef = useRef<HTMLButtonElement>(null);

  const handleScenarioPicked = (scenario: TestScenario) => {
    onApplyTestScenario?.(scenario);
    setScenariosOpen(false);
    scenariosTriggerRef.current?.focus();
  };

  return (
    <TaskCreateModalShell
      {...props}
      presentation={presentation}
      busyLabel={taskCreateModalBusyLabel()}
      scenariosOpen={scenariosOpen}
      scenariosTriggerRef={scenariosTriggerRef}
      onToggleScenarios={() => setScenariosOpen((open) => !open)}
      onScenarioPicked={handleScenarioPicked}
      onCloseScenarios={() => setScenariosOpen(false)}
    />
  );
}
