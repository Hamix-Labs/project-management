import type { ReactElement } from "react";
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { TaskBoardSection } from "./TaskBoardSection";
import { TaskBoardCard } from "./TaskBoardCard";
import { makeTask } from "@/test/taskDefaults";
import type { TaskWithDepth } from "../../task-tree";

function withDepth(t: ReturnType<typeof makeTask>): TaskWithDepth {
  return { ...t, depth: 0 };
}

function renderBoard(ui: ReactElement) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter future={ROUTER_FUTURE_FLAGS}>{ui}</MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("TaskBoardCard", () => {
  it("renders id, title, summary, tags, and created time without assignee", () => {
    renderBoard(
      <TaskBoardCard
        task={withDepth(
          makeTask({
            id: "abcdef12-3456-7890-abcd-ef1234567890",
            title: "Migrate sessions",
            initial_prompt: "<p>Move auth to Redis for scale.</p>",
            status: "running",
            tags: ["backend", "infra", "auth", "extra"],
            created_at: new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString(),
            project_id: "p1",
          }),
        )}
        showProject
        showTags
        projectName="core-api"
        prefetchTaskDetail={() => {}}
      />,
    );
    expect(screen.getByText("abcdef12")).toBeInTheDocument();
    expect(screen.getByText("Migrate sessions")).toBeInTheDocument();
    expect(screen.getByText(/Move auth to Redis/)).toBeInTheDocument();
    expect(screen.getByText("hamix/task-abcdef12")).toBeInTheDocument();
    expect(screen.getByText("core-api")).toBeInTheDocument();
    expect(
      document.querySelector(".task-board-card__project-chip svg"),
    ).toBeTruthy();
    expect(screen.getByText("backend")).toBeInTheDocument();
    expect(
      document.querySelector(".task-board-card__tag-chip svg"),
    ).toBeTruthy();
    expect(screen.getByText("+1")).toBeInTheDocument();
    expect(screen.getByText(/h ago|min ago|just now/)).toBeInTheDocument();
    expect(screen.queryByText(/assigned/i)).not.toBeInTheDocument();
  });
});

describe("TaskBoardSection", () => {
  it("renders columns, cards, active pill, and subtitle", () => {
    renderBoard(
      <TaskBoardSection
        tasks={[
          withDepth(
            makeTask({ id: "1", title: "Ready one", status: "ready" }),
          ),
          withDepth(
            makeTask({ id: "2", title: "Running one", status: "running" }),
          ),
        ]}
        loading={false}
        refreshing={false}
        error={null}
        truncated={false}
        onRetry={vi.fn()}
        smoothTransitions={false}
      />,
    );
    expect(screen.getByRole("heading", { name: "Board" })).toBeInTheDocument();
    expect(screen.getByText("2 active")).toBeInTheDocument();
    expect(
      screen.getByText("Track engineering work across every stage."),
    ).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Backlog" })).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "In Progress" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Ready one")).toBeInTheDocument();
    expect(screen.getByText("Running one")).toBeInTheDocument();
  });

  it("shows empty state when there are no active tasks", () => {
    renderBoard(
      <TaskBoardSection
        tasks={[]}
        loading={false}
        refreshing={false}
        error={null}
        truncated={false}
        onRetry={vi.fn()}
        smoothTransitions={false}
      />,
    );
    expect(screen.getByText("No active tasks")).toBeInTheDocument();
  });

  it("shows error with retry", () => {
    const onRetry = vi.fn();
    renderBoard(
      <TaskBoardSection
        tasks={[]}
        loading={false}
        refreshing={false}
        error="Network down"
        truncated={false}
        onRetry={onRetry}
        smoothTransitions={false}
      />,
    );
    expect(screen.getByRole("alert")).toHaveTextContent("Network down");
    screen.getByRole("button", { name: "Retry" }).click();
    expect(onRetry).toHaveBeenCalled();
  });

  it("shows truncation banner when truncated", () => {
    renderBoard(
      <TaskBoardSection
        tasks={[
          withDepth(makeTask({ id: "1", title: "Only", status: "ready" })),
        ]}
        loading={false}
        refreshing={false}
        error={null}
        truncated
        onRetry={vi.fn()}
        smoothTransitions={false}
      />,
    );
    expect(screen.getByRole("status")).toHaveTextContent(
      /first 500 active tasks/i,
    );
  });

  it("shows dashed empty column copy for unused columns", () => {
    renderBoard(
      <TaskBoardSection
        tasks={[
          withDepth(makeTask({ id: "1", title: "Only ready", status: "ready" })),
        ]}
        loading={false}
        refreshing={false}
        error={null}
        truncated={false}
        onRetry={vi.fn()}
        smoothTransitions={false}
      />,
    );
    expect(screen.getAllByText("No tasks").length).toBeGreaterThanOrEqual(1);
  });
});
