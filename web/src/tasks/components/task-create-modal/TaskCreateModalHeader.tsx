import type { LegacyRef, RefObject } from "react";
import { TestScenariosTrigger } from "./TestScenariosTrigger";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";

type Props = {
  presentation: TaskCreateModalPresentation;
  editingTaskId: string | null;
  draftSaveLabel: string | null;
  draftSaveError: boolean;
  disabled: boolean;
  scenariosOpen: boolean;
  scenariosTriggerRef: RefObject<HTMLButtonElement | null>;
  onToggleScenarios: () => void;
  onClose: () => void;
};

function CloseIcon() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      aria-hidden="true"
    >
      <path d="M4 4l8 8M12 4l-8 8" />
    </svg>
  );
}

export function TaskCreateModalHeader({
  presentation,
  editingTaskId,
  draftSaveLabel,
  draftSaveError,
  disabled,
  scenariosOpen,
  scenariosTriggerRef,
  onToggleScenarios,
  onClose,
}: Props) {
  const showSubtitle =
    !presentation.isEdit && !presentation.isTemplateMode;

  return (
    <header className="task-create-modal-header">
      <div className="task-create-modal-header__top">
        <div className="task-create-modal-header__title-block">
          <h2 id={presentation.modalTitleId} className="task-create-modal-title">
            {presentation.modalTitle}
          </h2>
          {showSubtitle ? (
            <p className="task-create-modal-subtitle">
              Define the work, then hand it off to your agent.
            </p>
          ) : null}
        </div>
        <div className="task-create-modal-header__actions">
          {!presentation.showTestScenarios ? null : (
            <TestScenariosTrigger
              ref={scenariosTriggerRef as LegacyRef<HTMLButtonElement>}
              open={scenariosOpen}
              disabled={disabled}
              onToggle={onToggleScenarios}
            />
          )}
          <button
            type="button"
            className="task-create-modal-close"
            aria-label="Close"
            disabled={disabled}
            onClick={onClose}
          >
            <CloseIcon />
          </button>
        </div>
      </div>
      {presentation.isTaskEdit && editingTaskId ? (
        <p
          className="muted stack-tight-zero task-create-modal-task-id"
          id="task-edit-modal-description"
        >
          <code>{editingTaskId}</code>
        </p>
      ) : null}
      {presentation.showDraftStatus ? (
        <p
          className={[
            "task-create-draft-status",
            draftSaveError ? "task-create-draft-status--error" : "",
          ]
            .filter(Boolean)
            .join(" ")}
          aria-live={draftSaveError ? "assertive" : "polite"}
        >
          {draftSaveLabel}
        </p>
      ) : null}
    </header>
  );
}
