import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { TaskCreateModalAutonomyToggle } from "../../task-create-modal/fields/TaskCreateModalAutonomyToggle";
import { TaskComposeAgentCard } from "./TaskComposeAgentCard";

vi.mock("@/tasks/create/hooks/useTaskCreateAgentOptions", () => ({
  useTaskCreateAgentOptions: () => ({
    modelIds: new Set<string>(),
    modelsForSelect: [],
    modelSelectBusy: false,
    modelFetchError: null,
    modelServerError: null,
  }),
}));

describe("TaskComposeAgentCard", () => {
  it("uses compact Autonomous copy and a switch control", () => {
    render(
      <TaskComposeAgentCard
        disabled={false}
        runner="cursor"
        cursorModel=""
        autonomyEnabled
        autonomyDisabled={false}
        onRunnerChange={vi.fn()}
        onCursorModelChange={vi.fn()}
        onAutonomyChange={vi.fn()}
      />,
    );

    expect(screen.getByText("Autonomous")).toBeInTheDocument();
    expect(
      screen.getByText("Picks it up when no other task is running."),
    ).toBeInTheDocument();
    expect(
      screen.queryByText("Autonomous execution"),
    ).not.toBeInTheDocument();
    expect(screen.getByRole("switch", { name: /autonomous/i })).toBeChecked();
  });
});

describe("TaskCreateModalAutonomyToggle", () => {
  it("keeps create-modal copy by default", () => {
    render(
      <TaskCreateModalAutonomyToggle
        enabled
        disabled={false}
        onChange={vi.fn()}
      />,
    );
    expect(screen.getByText("Autonomous execution")).toBeInTheDocument();
  });
});
