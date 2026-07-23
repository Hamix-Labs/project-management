import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it } from "vitest";
import { gitQueryKeys } from "@/lib/gitQueryKeys";
import { server } from "@/test/server";
import { globalGitApiHandlers } from "@/test/handlers/gitMsw";
import { FACTORY_GIT_WORKTREE_ID } from "@/test/factories/git";
import { TaskDetailGitBinding } from "./TaskDetailGitBinding";

const SAMPLE_TASK_ID = "0acaf529-9adf-4333-8992-29aa308eadba";

function createWrapper(qc: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
  };
}

describe("TaskDetailGitBinding", () => {
  beforeEach(() => {
    localStorage.clear();
    server.use(...globalGitApiHandlers());
  });

  it("renders branch and worktree context for a bound task", async () => {
    const user = userEvent.setup();
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    render(
      <TaskDetailGitBinding
        taskId={SAMPLE_TASK_ID}
        worktreeId={FACTORY_GIT_WORKTREE_ID}
      />,
      { wrapper: createWrapper(qc) },
    );

    await waitFor(() => {
      expect(screen.getByTestId("task-detail-git-binding")).toBeInTheDocument();
    });

    expect(screen.getByText("Branch")).toBeInTheDocument();
    expect(screen.getByText("Worktree")).toBeInTheDocument();
    expect(screen.getByTestId("task-commits-context")).toHaveTextContent("main");
    expect(screen.getByTestId("task-detail-git-binding-actions")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /copy worktree path/i })).toBeInTheDocument();

    await user.click(screen.getByTestId("task-detail-open-in-trigger"));
    expect(
      screen.getByRole("menuitem", { name: /open worktree in cursor/i }),
    ).toHaveAttribute("href", "cursor://file/repo/main/?windowId=_blank");
    expect(
      screen.getByRole("menuitem", { name: /open worktree in vs code/i }),
    ).toHaveAttribute("href", "vscode://file/repo/main/?windowId=_blank");
  });

  it("shows predicted branch and worktree names while provisioning", () => {
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    render(<TaskDetailGitBinding taskId={SAMPLE_TASK_ID} />, {
      wrapper: createWrapper(qc),
    });

    expect(screen.getByTestId("task-detail-git-binding")).toBeInTheDocument();
    expect(screen.getByTestId("task-detail-git-binding")).toHaveAttribute(
      "data-provisioning",
      "true",
    );
    expect(screen.getByTestId("task-commits-context")).toHaveTextContent(
      "hamix/task-0acaf529",
    );
    expect(screen.getByTestId("task-commits-context")).toHaveTextContent(
      "hamix-task-0acaf529",
    );
    expect(
      screen.getByTestId("task-detail-git-binding-pending"),
    ).toHaveTextContent(/preparing workspace/i);
    expect(
      screen.queryByTestId("task-detail-git-binding-actions"),
    ).not.toBeInTheDocument();
  });

  it("renders nothing when the worktree cannot be resolved", async () => {
    const missingWorktreeId = "00000000-0000-4000-8000-000000000099";
    const qc = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    render(
      <TaskDetailGitBinding
        taskId={SAMPLE_TASK_ID}
        worktreeId={missingWorktreeId}
      />,
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
