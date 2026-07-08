import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it } from "vitest";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { server } from "@/test/server";
import { globalGitApiHandlers } from "@/test/handlers/gitMsw";
import { FACTORY_GIT_WORKTREE_ID } from "@/test/factories/git";
import { TaskDetailGitBinding } from "./TaskDetailGitBinding";

function createWrapper(qc: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
  };
}

describe("TaskDetailGitBinding", () => {
  beforeEach(() => {
    server.use(...globalGitApiHandlers());
  });

  it("renders branch and worktree context for a bound task", async () => {
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    render(
      <TaskDetailGitBinding worktreeId={FACTORY_GIT_WORKTREE_ID} />,
      { wrapper: createWrapper(qc) },
    );

    await waitFor(() => {
      expect(screen.getByTestId("task-detail-git-binding")).toBeInTheDocument();
    });

    expect(screen.getByText("Branch")).toBeInTheDocument();
    expect(screen.getByText("Worktree")).toBeInTheDocument();
    expect(screen.getByTestId("task-commits-context")).toHaveTextContent("main");
  });

  it("renders nothing when the task has no worktree binding", () => {
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    render(<TaskDetailGitBinding />, { wrapper: createWrapper(qc) });

    expect(screen.queryByTestId("task-detail-git-binding")).not.toBeInTheDocument();
  });

  it("renders nothing when the worktree cannot be resolved", async () => {
    const missingWorktreeId = "00000000-0000-4000-8000-000000000099";
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    render(
      <TaskDetailGitBinding worktreeId={missingWorktreeId} />,
      { wrapper: createWrapper(qc) },
    );

    await waitFor(() => {
      expect(
        qc.getQueryState(gitQueryKeys.taskBinding(missingWorktreeId))?.status,
      ).toBe("success");
    });

    expect(screen.queryByTestId("task-detail-git-binding")).not.toBeInTheDocument();
  });
});
