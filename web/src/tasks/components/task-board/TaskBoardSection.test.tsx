import type { ReactElement } from "react";
import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { TaskBoardSection } from "./TaskBoardSection";
import { makeTask } from "@/test/taskDefaults";
import type { TaskWithDepth } from "../../task-tree";

function withDepth(t: ReturnType<typeof makeTask>): TaskWithDepth {
  return { ...t, depth: 0 };
}

function renderBoard(ui: ReactElement) {
  return render(
    <MemoryRouter future={ROUTER_FUTURE_FLAGS}>{ui}</MemoryRouter>,
  );
}

describe("TaskBoardSection", () => {
  it("renders columns and cards for active tasks", () => {
    renderBoard(
      <TaskBoardSection
        tasks={[
          withDepth(makeTask({ id: "1", title: "Ready one", status: "ready" })),
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
});
