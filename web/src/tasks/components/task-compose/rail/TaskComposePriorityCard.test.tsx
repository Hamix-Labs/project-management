import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { PRIORITIES } from "@/types";
import { TaskComposePriorityCard } from "./TaskComposePriorityCard";

describe("TaskComposePriorityCard", () => {
  it("exposes each level as data-priority instead of a shared data-warn", () => {
    render(
      <TaskComposePriorityCard value="high" onChange={vi.fn()} />,
    );

    for (const priority of PRIORITIES) {
      const label = priority[0].toUpperCase() + priority.slice(1);
      const segment = screen.getByRole("radio", { name: label });
      expect(segment).toHaveAttribute("data-priority", priority);
      expect(segment).not.toHaveAttribute("data-warn");
      expect(segment).toHaveAttribute(
        "data-active",
        priority === "high" ? "true" : "false",
      );
    }
  });

  it("notifies onChange with the chosen priority", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <TaskComposePriorityCard value="medium" onChange={onChange} />,
    );

    await user.click(screen.getByRole("radio", { name: "Critical" }));
    expect(onChange).toHaveBeenCalledWith("critical");
  });
});
