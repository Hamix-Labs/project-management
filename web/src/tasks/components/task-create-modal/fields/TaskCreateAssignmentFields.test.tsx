import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import type { ComponentProps } from "react";
import { useState } from "react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import {
  FACTORY_GIT_REPO_ID,
  gitBranchFactory,
  gitRepositoryFactory,
  gitWorktreeFactory,
} from "@/test/factories/git";
import { FACTORY_REPO_DEFAULT_PROJECT_ID, repoDefaultProjectFactory } from "@/test/factories/project";
import { TaskCreateAssignmentFields } from "./TaskCreateAssignmentFields";

const WT_TEST_1 = "00000000-0000-4000-8000-000000000021";
const WT_TEST_2 = "00000000-0000-4000-8000-000000000022";
const BRANCH_TEST_1 = "00000000-0000-4000-8000-000000000031";
const BRANCH_TEST_2 = "00000000-0000-4000-8000-000000000032";

function jsonResponse(body: unknown, init: ResponseInit = { status: 200 }): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

function mockGitFetch(onRequest?: (url: string) => void) {
  const repo = gitRepositoryFactory();
  const worktrees = [
    gitWorktreeFactory({
      id: WT_TEST_2,
      name: "test_2",
      path: "/repo/test_2",
      is_main: true,
      branch_id: BRANCH_TEST_2,
    }),
    gitWorktreeFactory({
      id: WT_TEST_1,
      name: "test_1",
      path: "/repo/test_1",
      is_main: false,
      branch_id: BRANCH_TEST_1,
    }),
  ];
  const branches = [
    gitBranchFactory({ id: BRANCH_TEST_2, name: "test_2" }),
    gitBranchFactory({ id: BRANCH_TEST_1, name: "test_1" }),
  ];

  return vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
    const url = String(input);
    onRequest?.(url);
    if (url.endsWith("/git/repositories")) {
      return jsonResponse({ repositories: [repo] });
    }
    if (url.includes(`/git/repositories/${FACTORY_GIT_REPO_ID}/worktrees`)) {
      return jsonResponse({ worktrees });
    }
    if (url.includes(`/git/repositories/${FACTORY_GIT_REPO_ID}/branches`)) {
      return jsonResponse({ branches });
    }
    if (url.includes(`/git/repositories/${FACTORY_GIT_REPO_ID}/projects`)) {
      return jsonResponse({
        projects: [repoDefaultProjectFactory()],
        limit: 100,
      });
    }
    return new Response("not found", { status: 404 });
  });
}

function renderFields(
  props: Partial<ComponentProps<typeof TaskCreateAssignmentFields>> = {},
) {
  const onAssignmentChange = vi.fn();
  const onProjectContextClear = vi.fn();
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });

  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <TaskCreateAssignmentFields
          idsPrefix="task-template-edit"
          repositoryId=""
          projectId={FACTORY_REPO_DEFAULT_PROJECT_ID}
          worktreeId={WT_TEST_1}
          onAssignmentChange={onAssignmentChange}
          onProjectContextClear={onProjectContextClear}
          {...props}
        />
      </MemoryRouter>
    </QueryClientProvider>,
  );

  return { onAssignmentChange, onProjectContextClear };
}

describe("TaskCreateAssignmentFields", () => {
  it("does not auto-select repository when project and worktree are already hydrated", async () => {
    const fetchMock = mockGitFetch();
    const { onAssignmentChange } = renderFields();

    await waitFor(() => {
      expect(screen.getByRole("combobox", { name: /repository/i })).toBeInTheDocument();
    });

    expect(onAssignmentChange).not.toHaveBeenCalled();

    fetchMock.mockRestore();
  });

  it("keeps a saved non-default worktree when repository id is already hydrated", async () => {
    const fetchMock = mockGitFetch();
    const { onAssignmentChange } = renderFields({
      repositoryId: FACTORY_GIT_REPO_ID,
    });

    await waitFor(() => {
      expect(screen.getByRole("combobox", { name: /worktree/i })).toHaveTextContent(
        "test_1 (test_1)",
      );
    });

    await waitFor(() => {
      expect(onAssignmentChange).not.toHaveBeenCalled();
    });

    fetchMock.mockRestore();
  });

  it("propagates single-repo auto-select on a fresh form without refetch loops", async () => {
    let worktreeListCalls = 0;
    const fetchMock = mockGitFetch((url) => {
      if (url.includes("/worktrees") && !url.includes("/worktrees/live")) {
        worktreeListCalls += 1;
      }
    });

    function FreshFormHarness() {
      const [repositoryId, setRepositoryId] = useState("");
      const [projectId, setProjectId] = useState("");
      const [worktreeId, setWorktreeId] = useState("");
      const client = new QueryClient({
        defaultOptions: { queries: { retry: false, gcTime: 0 } },
      });

      return (
        <QueryClientProvider client={client}>
          <MemoryRouter>
            <TaskCreateAssignmentFields
              idsPrefix="task-template-new"
              repositoryId={repositoryId}
              projectId={projectId}
              worktreeId={worktreeId}
              onAssignmentChange={(next) => {
                setRepositoryId(next.repositoryId);
                setProjectId(next.projectId);
                setWorktreeId(next.worktreeId);
              }}
              onProjectContextClear={() => {}}
            />
          </MemoryRouter>
        </QueryClientProvider>
      );
    }

    render(<FreshFormHarness />);

    await waitFor(() => {
      expect(screen.getByRole("combobox", { name: /repository/i })).not.toHaveTextContent(
        /^▾$/,
      );
    });

    await waitFor(() => {
      expect(screen.getByRole("combobox", { name: /worktree/i })).toHaveTextContent(
        "test_2 (test_2)",
      );
    });

    expect(worktreeListCalls).toBeLessThanOrEqual(2);

    fetchMock.mockRestore();
  });
});
