import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { CYCLE_STATUSES } from "@/types/cycle";
import { CycleStatusBadge, CYCLE_STATUS_META } from "./index";

describe("CycleStatusBadge", () => {
  it.each(CYCLE_STATUSES)("renders label for %s", (status) => {
    render(<CycleStatusBadge status={status} />);
    expect(screen.getByText(CYCLE_STATUS_META[status].label)).toBeInTheDocument();
  });

  it("applies pulse class for running status", () => {
    const { container } = render(<CycleStatusBadge status="running" />);
    expect(container.firstChild).toHaveClass("task-status-badge--pulse");
    expect(container.firstChild).toHaveClass("task-status-badge--tone-info");
  });

  it("maps succeeded to success tone", () => {
    const { container } = render(<CycleStatusBadge status="succeeded" />);
    expect(container.firstChild).toHaveClass("task-status-badge--tone-success");
  });
});
