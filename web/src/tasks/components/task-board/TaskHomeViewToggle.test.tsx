import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaskHomeViewToggle } from "./TaskHomeViewToggle";

describe("TaskHomeViewToggle", () => {
  it("exposes tablist semantics and calls onChange", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TaskHomeViewToggle value="list" onChange={onChange} />);
    expect(screen.getByRole("tablist", { name: "Task view" })).toBeInTheDocument();
    const listTab = screen.getByRole("tab", { name: "List" });
    const boardTab = screen.getByRole("tab", { name: "Board" });
    expect(listTab).toHaveAttribute("aria-selected", "true");
    expect(boardTab).toHaveAttribute("aria-selected", "false");
    expect(listTab).toHaveAttribute("aria-controls", "task-list-panel");
    expect(boardTab).toHaveAttribute("aria-controls", "task-board-panel");
    await user.click(boardTab);
    expect(onChange).toHaveBeenCalledWith("board");
  });

  it("moves selection with arrow keys", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TaskHomeViewToggle value="list" onChange={onChange} />);
    const listTab = screen.getByRole("tab", { name: "List" });
    listTab.focus();
    await user.keyboard("{ArrowRight}");
    expect(onChange).toHaveBeenCalledWith("board");
  });
});
