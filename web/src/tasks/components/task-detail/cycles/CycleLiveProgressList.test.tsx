import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { CycleLiveProgressList } from "./CycleLiveProgressList";
import type { AgentRunProgressItem } from "@/tasks/hooks/useAgentRunProgress";

const NOW = 1_000_000;

function frame(
  overrides: Partial<AgentRunProgressItem> & Pick<AgentRunProgressItem, "progress">,
  receivedAt: number,
): AgentRunProgressItem {
  return {
    taskId: "task-1",
    cycleId: "cycle-1",
    phaseSeq: 1,
    receivedAt,
    ...overrides,
  };
}

describe("CycleLiveProgressList", () => {
  it("renders empty message when there are no items", () => {
    render(
      <CycleLiveProgressList
        items={[]}
        now={NOW}
        emptyMessage="Preparing execute…"
      />,
    );
    expect(screen.getByTestId("task-cycle-progress-empty")).toHaveTextContent(
      /Preparing execute/,
    );
  });

  it("renders nothing when empty without emptyMessage", () => {
    const { container } = render(<CycleLiveProgressList items={[]} now={NOW} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("marks the newest item as latest and caps at maxItems", () => {
    const items = [
      frame({ progress: { kind: "system", message: "older" } }, NOW - 20_000),
      frame({ progress: { kind: "assistant", message: "newest" } }, NOW - 1_000),
      frame({ progress: { kind: "tool_call", message: "mid" } }, NOW - 10_000),
      frame({ progress: { kind: "system", message: "oldest" } }, NOW - 30_000),
    ];
    const { container } = render(
      <CycleLiveProgressList items={items} now={NOW} maxItems={2} showPendingRow={false} />,
    );
    const rows = container.querySelectorAll(".task-cycle-progress-item");
    expect(rows).toHaveLength(2);
    expect(rows[0]).toHaveClass("task-cycle-progress-item--latest");
    expect(rows[0]).toHaveTextContent(/newest/);
    expect(rows[1]).toHaveTextContent(/mid/);
  });

  it("exposes list aria-label for screen readers", () => {
    const items = [
      frame({ progress: { kind: "assistant", message: "Hi" } }, NOW - 2_000),
    ];
    render(<CycleLiveProgressList items={items} now={NOW} showPendingRow={false} />);
    expect(screen.getByRole("list", { name: /recent agent progress/i })).toBeInTheDocument();
  });

  it("renders pending Working row with relative time (no Last prefix)", () => {
    const items = [
      frame({ progress: { kind: "assistant", message: "Hi" } }, NOW - 2_000),
    ];
    render(<CycleLiveProgressList items={items} now={NOW} showPendingRow />);
    const pending = screen.getByLabelText(/agent working/i);
    expect(pending).toBeInTheDocument();
    expect(pending).toHaveTextContent(/Working/i);
    expect(pending).toHaveTextContent("2s ago");
    expect(pending).not.toHaveTextContent(/Last 2s ago/);
  });
});
