import { describe, expect, it, vi } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach } from "vitest";
import { TaskHomeViewToggle } from "./TaskHomeViewToggle";

afterEach(() => {
  cleanup();
});

describe("TaskHomeViewToggle", () => {
  it("exposes tablist semantics and calls onChange", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TaskHomeViewToggle value="list" onChange={onChange} />);
    expect(screen.getByRole("tablist", { name: "Task view" })).toBeInTheDocument();
    const listTab = screen.getByRole("tab", { name: "List" });
    const boardTab = screen.getByRole("tab", { name: "Board" });
    const timelineTab = screen.getByRole("tab", { name: "Timeline" });
    expect(listTab).toHaveAttribute("aria-selected", "true");
    expect(boardTab).toHaveAttribute("aria-selected", "false");
    expect(timelineTab).toHaveAttribute("aria-selected", "false");
    expect(listTab).toHaveAttribute("aria-controls", "task-list-panel");
    expect(boardTab).toHaveAttribute("aria-controls", "task-board-panel");
    expect(timelineTab).toHaveAttribute("aria-controls", "task-timeline-panel");
    await user.click(boardTab);
    expect(onChange).toHaveBeenCalledWith("board");
    await user.click(timelineTab);
    expect(onChange).toHaveBeenCalledWith("timeline");
  });

  it("cycles selection with arrow keys across three views", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    const { rerender } = render(
      <TaskHomeViewToggle value="list" onChange={onChange} />,
    );
    screen.getByRole("tab", { name: "List" }).focus();
    await user.keyboard("{ArrowRight}");
    expect(onChange).toHaveBeenCalledWith("board");

    onChange.mockClear();
    rerender(<TaskHomeViewToggle value="board" onChange={onChange} />);
    screen.getByRole("tab", { name: "Board" }).focus();
    await user.keyboard("{ArrowRight}");
    expect(onChange).toHaveBeenCalledWith("timeline");

    onChange.mockClear();
    rerender(<TaskHomeViewToggle value="timeline" onChange={onChange} />);
    screen.getByRole("tab", { name: "Timeline" }).focus();
    await user.keyboard("{ArrowRight}");
    expect(onChange).toHaveBeenCalledWith("list");
  });
});
