import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { PRIORITIES } from "@/types";
import { PriorityBadge } from "./PriorityBadge";
import { PRIORITY_META } from "./priorityMeta";
import { priorityListLabel } from "./priorityListLabel";

describe("PriorityBadge", () => {
  it.each(PRIORITIES)("renders label for %s", (priority) => {
    render(<PriorityBadge priority={priority} />);
    expect(screen.getByText(priorityListLabel(priority))).toBeInTheDocument();
  });

  it("renders four meter bars", () => {
    const { container } = render(<PriorityBadge priority="high" />);
    expect(container.querySelectorAll(".task-priority-badge__bar")).toHaveLength(4);
  });

  it("fills bars up to priority weight", () => {
    const { container } = render(<PriorityBadge priority="critical" />);
    const filled = container.querySelectorAll(".task-priority-badge__bar--filled");
    expect(filled).toHaveLength(PRIORITY_META.critical.weight);
  });

  it("maps every priority to a tone class", () => {
    for (const priority of PRIORITIES) {
      const { container } = render(<PriorityBadge priority={priority} />);
      const tone = PRIORITY_META[priority].tone;
      expect(container.firstChild).toHaveClass(
        `task-priority-badge--tone-${tone}`,
      );
    }
  });
});
