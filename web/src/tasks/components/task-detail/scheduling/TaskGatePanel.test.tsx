import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { TaskGatePanel } from "./TaskGatePanel";

describe("TaskGatePanel", () => {
  it("shows empty state when gate is absent", () => {
    render(<TaskGatePanel gate={null} />);
    expect(screen.getByTestId("task-gate-empty")).toBeInTheDocument();
  });

  it("shows active gate status", () => {
    render(
      <TaskGatePanel
        gate={{
          kind: "manual_approval",
          status: "active",
          hold: false,
        }}
      />,
    );
    expect(screen.getByTestId("task-gate-meta")).toHaveTextContent("Active");
  });

  it("shows pending release with hold chip", () => {
    render(
      <TaskGatePanel
        gate={{
          kind: "manual_approval",
          status: "pending_release",
          hold: true,
        }}
      />,
    );
    expect(screen.getByText("Pending release")).toBeInTheDocument();
    expect(screen.getByText("Paused")).toBeInTheDocument();
  });
});
