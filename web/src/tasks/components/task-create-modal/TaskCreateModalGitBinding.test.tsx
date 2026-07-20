import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { useState } from "react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import type { AppSettings, ListCursorModelsResult } from "@/api/settings";
import { settingsQueryKeys } from "@/settings/settingsQueryKeys";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import { respondGlobalGitApi } from "@/test/handlers/gitGlobal";
import { APP_SETTINGS_DEFAULTS } from "@/test/settingsDefaults";
import { buildTaskCreateModalProps } from "./buildTaskCreateModalProps";
import { TaskCreateModal } from "./TaskCreateModal";
import type { TaskCreateModalFlatInput } from "./taskCreateModalProps";

const testAppSettings: AppSettings = {
  ...APP_SETTINGS_DEFAULTS,
  ...TASK_TEST_DEFAULTS,
  optimistic_mutations_enabled: false,
};

const testCursorModelsEmpty: ListCursorModelsResult = {
  ok: true,
  runner: TASK_TEST_DEFAULTS.runner,
  models: [],
};

const flatDefaults = {
  pending: false,
  saving: false,
  draftSaving: false,
  draftSaveLabel: null,
  draftSaveError: false,
  onClose: vi.fn(),
  title: "Draft title",
  prompt: "Draft prompt",
  priority: "medium" as const,
  checklistItems: [{ text: "Criterion" }],
  onTitleChange: vi.fn(),
  onPromptChange: vi.fn(),
  onPriorityChange: vi.fn(),
  onAppendChecklistCriterion: vi.fn(),
  onUpdateChecklistRow: vi.fn(),
  onRemoveChecklistRow: vi.fn(),
  taskRunner: TASK_TEST_DEFAULTS.runner,
  taskCursorModel: TASK_TEST_DEFAULTS.cursor_model,
  onTaskRunnerChange: vi.fn(),
  onTaskCursorModelChange: vi.fn(),
  schedule: null,
  onScheduleChange: vi.fn(),
  autonomyEnabled: true,
  onAutonomyChange: vi.fn(),
  tagsCsv: "",
  milestone: "",
  repositoryId: "",
  projectId: "",
  worktreeId: "",
  onRepositoryChange: vi.fn(),
  onProjectChange: vi.fn(),
  onWorktreeChange: vi.fn(),
  onProjectContextClear: vi.fn(),
  dependsOn: [],
  onTagsCsvChange: vi.fn(),
  onMilestoneChange: vi.fn(),
  onDependsOnChange: vi.fn(),
  appTimezone: "UTC",
  onSaveDraft: vi.fn(),
  onSubmit: vi.fn(),
} satisfies TaskCreateModalFlatInput;

function renderModal(overrides?: Partial<TaskCreateModalFlatInput>) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  });
  client.setQueryData(settingsQueryKeys.app(), testAppSettings);
  client.setQueryData(
    settingsQueryKeys.cursorModels("cursor", ""),
    testCursorModelsEmpty,
  );
  return render(
    <MemoryRouter>
      <QueryClientProvider client={client}>
        <TaskCreateModal
          {...buildTaskCreateModalProps({ ...flatDefaults, ...overrides })}
        />
      </QueryClientProvider>
    </MemoryRouter>,
  );
}

function stubGitFetch() {
  vi.spyOn(globalThis, "fetch").mockImplementation(async (input: RequestInfo | URL) => {
    const url = typeof input === "string" ? input : input.toString();
    const git = respondGlobalGitApi(url);
    if (git) return git;
    return new Response("not found", { status: 404 });
  });
}

describe("TaskCreateModal git binding", () => {
  it("disables Create task until repository is selected", async () => {
    stubGitFetch();
    renderModal({ repositoryId: "", worktreeId: "" });
    expect(screen.getByRole("button", { name: /Create task/i })).toBeDisabled();
  });

  it("preselects when only one repository exists", async () => {
    stubGitFetch();

    function Harness() {
      const [repositoryId, setRepositoryId] = useState("");
      const [projectId, setProjectId] = useState("");
      const [worktreeId, setWorktreeId] = useState("");
      return (
        <TaskCreateModal
          {...buildTaskCreateModalProps({
            ...flatDefaults,
            repositoryId,
            projectId,
            worktreeId,
            onRepositoryChange: setRepositoryId,
            onProjectChange: setProjectId,
            onWorktreeChange: setWorktreeId,
          })}
        />
      );
    }

    const client = new QueryClient({
      defaultOptions: { queries: { retry: false, staleTime: Infinity } },
    });
    client.setQueryData(settingsQueryKeys.app(), testAppSettings);
    client.setQueryData(
      settingsQueryKeys.cursorModels("cursor", ""),
      testCursorModelsEmpty,
    );

    render(
      <MemoryRouter>
        <QueryClientProvider client={client}>
          <Harness />
        </QueryClientProvider>
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByRole("combobox", { name: /repository/i })).not.toHaveTextContent(
        /^▾$/,
      );
    });
    expect(screen.queryByRole("combobox", { name: /worktree/i })).not.toBeInTheDocument();
  });
});
