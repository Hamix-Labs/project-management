import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { STATUSES } from "@/types";
import { StatusBadge, STATUS_META, statusListLabel } from "./index";

describe("StatusBadge", () => {
  it.each(STATUSES)("renders label for %s", (status) => {
    render(<StatusBadge status={status} />);
    expect(screen.getByText(statusListLabel(status))).toBeInTheDocument();
  });

  it("applies pulse class for running status", () => {
    const { container } = render(<StatusBadge status="running" />);
    expect(container.firstChild).toHaveClass("task-status-badge--pulse");
  });

  it("does not apply pulse class for ready status", () => {
    const { container } = render(<StatusBadge status="ready" />);
    expect(container.firstChild).not.toHaveClass("task-status-badge--pulse");
  });

  it("maps review to the review tone (matches status filter pills)", () => {
    const { container } = render(<StatusBadge status="review" />);
    expect(container.firstChild).toHaveClass("task-status-badge--tone-review");
    expect(container.firstChild).not.toHaveClass(
      "task-status-badge--tone-warning",
    );
  });

  it("maps every status to a tone class", () => {
    for (const status of STATUSES) {
      const { container } = render(<StatusBadge status={status} />);
      const tone = STATUS_META[status].tone;
      expect(container.firstChild).toHaveClass(`task-status-badge--tone-${tone}`);
    }
  });
});
