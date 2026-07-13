import type { CursorModelOption } from "@/api/cursorModels";
import type { CustomSelectOption } from "@/components/custom-select";

export const RUNNERS = [{ id: "cursor", label: "Cursor CLI" }] as const;

export const AGENT_HEADING_ID = "task-create-agent-heading";

export const RUNNER_OPTIONS: CustomSelectOption[] = RUNNERS.map((r) => ({
  value: r.id,
  label: r.label,
}));

export function runnerDisplayLabel(runnerId: string): string {
  const row = RUNNERS.find((r) => r.id === runnerId);
  return row?.label ?? runnerId;
}

export type TaskCreateModalAgentSectionVariant = "default" | "modelDialog" | "createModal";

export type TaskCreateModalAgentSectionProps = {
  disabled: boolean;
  lockRunner?: boolean;
  variant?: TaskCreateModalAgentSectionVariant;
  runner: string;
  cursorModel: string;
  modelIds: Set<string>;
  modelsForSelect: CursorModelOption[];
  modelSelectBusy: boolean;
  modelFetchError: string | null;
  modelServerError: string | null;
  onRunnerChange: (runner: string) => void;
  onCursorModelChange: (v: string) => void;
};

export function AlertGlyph() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.6"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <circle cx="8" cy="8" r="6.25" />
      <path d="M8 5v3.5" />
      <path d="M8 10.75v0.25" />
    </svg>
  );
}
