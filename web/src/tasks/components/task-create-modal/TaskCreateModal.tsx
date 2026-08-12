import { useRef, useState } from "react";
import type { TestScenario } from "@/tasks/test-scenarios";
import { TaskComposeForm } from "../task-compose/TaskComposeForm";
import { resolveTaskCreateModalPresentation } from "./taskCreateModalPresentation";
import type { TaskCreateModalProps } from "./taskCreateModalProps";

export type { TaskCreateModalProps };

/**
 * Compose form entry used by unit tests and any non-route host.
 * Production create/edit uses {@link TaskComposePage} with the same fields.
 */
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
    <TaskComposeForm
      {...props}
      presentation={presentation}
      backTo="/"
      scenariosOpen={scenariosOpen}
      scenariosTriggerRef={scenariosTriggerRef}
      onToggleScenarios={() => setScenariosOpen((open) => !open)}
      onScenarioPicked={handleScenarioPicked}
      onCloseScenarios={() => setScenariosOpen(false)}
    />
  );
}
