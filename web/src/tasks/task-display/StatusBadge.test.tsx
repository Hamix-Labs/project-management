import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { STATUSES } from "@/types";
import { StatusBadge } from "./StatusBadge";
import { STATUS_META } from "./statusMeta";
import { statusListLabel } from "./statusListLabel";

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

  it("maps every status to a tone class", () => {
    for (const status of STATUSES) {
      const { container } = render(<StatusBadge status={status} />);
      const tone = STATUS_META[status].tone;
      expect(container.firstChild).toHaveClass(`task-status-badge--tone-${tone}`);
    }
  });
});
