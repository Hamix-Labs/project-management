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

    const labels: Record<(typeof PRIORITIES)[number], string> = {
      low: "Low",
      medium: "Medium",
      high: "High",
      critical: "Urgent",
    };
    for (const priority of PRIORITIES) {
      const segment = screen.getByRole("radio", { name: labels[priority] });
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

    await user.click(screen.getByRole("radio", { name: "Urgent" }));
    expect(onChange).toHaveBeenCalledWith("critical");
  });
});
