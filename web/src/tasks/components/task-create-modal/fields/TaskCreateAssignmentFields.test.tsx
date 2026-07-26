import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import type { ComponentProps } from "react";
import { useState } from "react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import {
  FACTORY_GIT_REPO_ID,
  gitRepositoryFactory,
} from "@/test/factories/git";
import { FACTORY_REPO_DEFAULT_PROJECT_ID, repoDefaultProjectFactory } from "@/test/factories/project";
import { TaskCreateAssignmentFields } from "./TaskCreateAssignmentFields";

function jsonResponse(body: unknown, init: ResponseInit = { status: 200 }): Response {
  return new Response(JSON.stringify(body), {
    ...init,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

function mockGitFetch(onRequest?: (url: string) => void) {
  const repo = gitRepositoryFactory();

  return vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
    const url = String(input);
    onRequest?.(url);
    if (url.endsWith("/git/repositories")) {
      return jsonResponse({ repositories: [repo] });
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
          worktreeId=""
          onAssignmentChange={onAssignmentChange}
          {...props}
        />
      </MemoryRouter>
    </QueryClientProvider>,
  );

  return { onAssignmentChange };
}

describe("TaskCreateAssignmentFields", () => {
  it("does not auto-select repository when project is already hydrated", async () => {
    const fetchMock = mockGitFetch();
    const { onAssignmentChange } = renderFields();

    await waitFor(() => {
      expect(screen.getByRole("combobox", { name: /repository/i })).toBeInTheDocument();
    });

    expect(onAssignmentChange).not.toHaveBeenCalled();
    expect(screen.queryByRole("combobox", { name: /worktree/i })).not.toBeInTheDocument();
    expect(screen.getByText(/Hamix allocates a worktree/i)).toBeInTheDocument();

    fetchMock.mockRestore();
  });

  it("keeps a hydrated repository without emitting assignment changes", async () => {
    const fetchMock = mockGitFetch();
    const { onAssignmentChange } = renderFields({
      repositoryId: FACTORY_GIT_REPO_ID,
    });

    await waitFor(() => {
      expect(screen.getByRole("combobox", { name: /repository/i })).not.toHaveTextContent(/^▾$/);
    });

    await waitFor(() => {
      expect(onAssignmentChange).not.toHaveBeenCalled();
    });

    fetchMock.mockRestore();
  });

  it("propagates single-repo auto-select on a fresh form without worktree UI", async () => {
    let projectListCalls = 0;
    const fetchMock = mockGitFetch((url) => {
      if (url.includes("/projects")) {
        projectListCalls += 1;
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
      expect(screen.getByRole("combobox", { name: /project/i })).toBeInTheDocument();
    });

    expect(screen.queryByRole("combobox", { name: /worktree/i })).not.toBeInTheDocument();
    expect(projectListCalls).toBeGreaterThan(0);

    fetchMock.mockRestore();
  });
});
