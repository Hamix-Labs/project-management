import { useMemo } from "react";
import { normalizeCursorModelSelectValue } from "@/api/cursorModels";
import { CustomSelect, type CustomSelectOption } from "@/components/custom-select";
import {
  AlertGlyph,
  runnerDisplayLabel,
  type TaskCreateModalAgentSectionVariant,
} from "./taskCreateModalAgentShared";

type Props = {
  modelId: string;
  disabled: boolean;
  lockRunner: boolean;
  variant: TaskCreateModalAgentSectionVariant;
  runner: string;
  cursorModel: string;
  modelIds: Set<string>;
  modelsForSelect: import("@/api/cursorModels").CursorModelOption[];
  modelSelectBusy: boolean;
  modelFetchError: string | null;
  modelServerError: string | null;
  onCursorModelChange: (v: string) => void;
};

export function TaskCreateModalModelField({
  modelId,
  disabled,
  lockRunner,
  variant,
  runner,
  cursorModel,
  modelIds,
  modelsForSelect,
  modelSelectBusy,
  modelFetchError,
  modelServerError,
  onCursorModelChange,
}: Props) {
  const modelSelectDisabled = disabled || modelSelectBusy;
  const cursorModelSelectValue = normalizeCursorModelSelectValue(cursorModel);

  const modelOptions = useMemo((): CustomSelectOption[] => {
    const opts: CustomSelectOption[] = [{ value: "", label: "Auto" }];
    for (const m of modelsForSelect) {
      opts.push({ value: m.id, label: m.label });
    }
    if (
      cursorModelSelectValue !== "" &&
      !modelIds.has(cursorModelSelectValue)
    ) {
      opts.push({
        value: cursorModelSelectValue,
        label: `${cursorModelSelectValue} (saved - not in current list)`,
      });
    }
    return opts;
  }, [modelsForSelect, cursorModelSelectValue, modelIds]);

  const isModelDialog = variant === "modelDialog";
  const isCreateModal = variant === "createModal";
  const showModelHelp =
    !isCreateModal ||
    modelSelectBusy ||
    lockRunner ||
    Boolean(modelFetchError) ||
    Boolean(modelServerError);

  return (
    <div className="task-create-agent-field">
      {runner === "cursor" ? (
        <>
          <CustomSelect
            id={modelId}
            label="Model"
            value={cursorModelSelectValue}
            options={modelOptions}
            disabled={modelSelectDisabled}
            onChange={onCursorModelChange}
            triggerTestId="task-create-cursor-model-select"
            className="task-create-agent-custom-select"
          />
          {showModelHelp ? (
            <p className="task-create-agent-help">
              {modelSelectBusy
                ? "Loading available models."
                : lockRunner
                  ? isModelDialog
                    ? "Pick a model or Auto. This overrides the workspace default for this task only."
                    : "Per-task: pick a model or Auto. Auto lets cursor-agent choose for this task only."
                  : isCreateModal
                    ? "Auto uses the workspace default."
                    : "Auto uses Cursor's current default unless overridden."}
            </p>
          ) : null}
          {modelFetchError ? (
            <div role="alert" className="task-create-agent-model-err">
              <AlertGlyph />
              <span>
                Could not load models for {runnerDisplayLabel(runner)}: {modelFetchError}
              </span>
            </div>
          ) : null}
          {modelServerError ? (
            <div role="alert" className="task-create-agent-model-err">
              <AlertGlyph />
              <span>{modelServerError}</span>
            </div>
          ) : null}
        </>
      ) : (
        <>
          <div className="field task-create-agent-text-field">
            <label htmlFor={modelId}>Model</label>
            <input
              id={modelId}
              className="task-create-agent-input"
              type="text"
              value={cursorModel}
              disabled={disabled}
              onChange={(e) => onCursorModelChange(e.target.value)}
              placeholder="Model id (optional)"
              autoComplete="off"
            />
          </div>
          <p className="task-create-agent-help">
            Optional - leave blank to use the runner default.
          </p>
        </>
      )}
    </div>
  );
}
