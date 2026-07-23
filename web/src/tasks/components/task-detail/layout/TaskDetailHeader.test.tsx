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
  it("renders title, priority pill, and back link without a status badge", () => {
    renderHeader({
      title: "My task",
      priority: "high",
      ...TASK_TEST_DEFAULTS,
    });

    expect(screen.getByRole("heading", { name: /^my task$/i })).toBeInTheDocument();
    expect(screen.getByText("High priority")).toBeInTheDocument();
    expect(screen.queryByText("Ready", { selector: ".task-status-badge" })).toBeNull();
    expect(screen.getByRole("link", { name: /^all tasks$/i })).toHaveAttribute(
      "href",
      "/",
    );
  });

  it("does not render a status badge for needs-user statuses", () => {
    renderHeader({
      title: "Blocked",
      priority: "medium",
      ...TASK_TEST_DEFAULTS,
      cursor_model: "opus",
    });

    expect(document.querySelector(".task-status-badge")).toBeNull();
    // The old standalone stance line is gone — guard against its return.
    expect(screen.queryByText("Agent needs input")).not.toBeInTheDocument();
    expect(screen.queryByText("Informational")).not.toBeInTheDocument();
  });

  it("renders the runtime chip with runner and model intent (Phase 4a of plan)", () => {
    renderHeader({
      title: "Has model",
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
      priority: "medium",
      ...TASK_TEST_DEFAULTS,
    });

    expect(
      screen.queryByRole("button", { name: /change model/i }),
    ).not.toBeInTheDocument();
  });
});
