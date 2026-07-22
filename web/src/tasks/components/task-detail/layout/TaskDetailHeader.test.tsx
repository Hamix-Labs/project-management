import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "../../../../lib/routerFutureFlags";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import { TaskDetailHeader } from "./TaskDetailHeader";

function renderHeader(task: Parameters<typeof TaskDetailHeader>[0]["task"]) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={qc}>
        <MemoryRouter future={ROUTER_FUTURE_FLAGS}>{children}</MemoryRouter>
      </QueryClientProvider>
    );
  }
  return render(<TaskDetailHeader task={task} />, { wrapper: Wrapper });
}

describe("TaskDetailHeader", () => {
  it("renders title, status and priority pills, and back link", () => {
    renderHeader({
      title: "My task",
      status: "ready",
      priority: "high",
      ...TASK_TEST_DEFAULTS,
    });

    expect(screen.getByRole("heading", { name: /^my task$/i })).toBeInTheDocument();
    expect(screen.getByText("Ready")).toBeInTheDocument();
    expect(screen.getByText("High priority")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /^all tasks$/i })).toHaveAttribute(
      "href",
      "/",
    );
  });

  // The header is identity-only: the needs-user signal lives in the
  // attention callout, while the header surfaces it as a highlighted
  // status pill (`data-needs-user`). No separate stance line — that was
  // redundant with both the pill and the callout. (Redesign 2026-06-04.)
  it("highlights the status pill when status needs user input", () => {
    renderHeader({
      title: "Blocked",
      status: "blocked",
      priority: "medium",
      ...TASK_TEST_DEFAULTS,
      cursor_model: "opus",
    });

    expect(
      screen.getByText("Blocked", { selector: ".task-status-badge" }),
    ).toHaveAttribute("data-needs-user", "true");
    // The old standalone stance line is gone — guard against its return.
    expect(screen.queryByText("Agent needs input")).not.toBeInTheDocument();
    expect(screen.queryByText("Informational")).not.toBeInTheDocument();
  });

  it("renders the runtime chip with runner and model intent (Phase 4a of plan)", () => {
    renderHeader({
      title: "Has model",
      status: "ready",
      priority: "medium",
      ...TASK_TEST_DEFAULTS,
      cursor_model: "opus-4",
    });
    const chip = screen.getByTestId("task-detail-runtime");
    expect(chip).toHaveTextContent("Cursor CLI · opus-4");
    expect(chip.className).toContain("cell-pill--runtime");
  });

  it("renders 'default model' copy in the runtime chip when task has no cursor_model selected", () => {
    renderHeader({
      title: "No model",
      status: "ready",
      priority: "medium",
      ...TASK_TEST_DEFAULTS,
      cursor_model: "",
    });
    expect(screen.getByTestId("task-detail-runtime")).toHaveTextContent(
      "Cursor CLI · default model",
    );
  });

  it("does not render a header change-model control", () => {
    renderHeader({
      title: "T",
      status: "ready",
      priority: "medium",
      ...TASK_TEST_DEFAULTS,
    });

    expect(
      screen.queryByRole("button", { name: /change model/i }),
    ).not.toBeInTheDocument();
  });
});
