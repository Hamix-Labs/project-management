import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { createTimelineFixtures } from "./timelineFixtures";
import { TaskTimelineSection } from "./TaskTimelineSection";

const NOW = new Date("2026-07-25T18:00:00.000Z");

function renderSection() {
  return render(
    <MemoryRouter future={ROUTER_FUTURE_FLAGS}>
      <TaskTimelineSection
        events={createTimelineFixtures(NOW)}
        now={NOW}
      />
    </MemoryRouter>,
  );
}

describe("TaskTimelineSection", () => {
  it("renders timeline heading, filters, and fixture events", () => {
    renderSection();
    expect(
      screen.getByRole("heading", { name: "Timeline" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("tabpanel", { name: "Timeline" }),
    ).toHaveAttribute("id", "task-timeline-panel");
    expect(
      screen.getByRole("button", { name: "All events" }),
    ).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByText("Verification passed")).toBeInTheDocument();
    expect(screen.getByText("Task created")).toBeInTheDocument();
  });

  it("filters to verification events", async () => {
    const user = userEvent.setup();
    renderSection();
    await user.click(screen.getByRole("button", { name: "Verification" }));
    expect(screen.getByText("Verification passed")).toBeInTheDocument();
    expect(screen.getByText("Verification failed")).toBeInTheDocument();
    expect(screen.queryByText("Task created")).not.toBeInTheDocument();
    expect(screen.queryByText("Agent run started")).not.toBeInTheDocument();
  });

  it("shows empty state when range excludes all events", async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter future={ROUTER_FUTURE_FLAGS}>
        <TaskTimelineSection
          events={[
            {
              id: "old",
              kind: "task-created",
              category: "tasks",
              at: "2020-01-01T12:00:00.000Z",
              title: "Task created",
              highlight: "ancient",
              detail: "Too old.",
            },
          ]}
          now={NOW}
        />
      </MemoryRouter>,
    );
    await user.click(screen.getByRole("button", { name: /Last 7 days/i }));
    await user.click(screen.getByRole("option", { name: "Last 24 hours" }));
    expect(
      screen.getByRole("status"),
    ).toHaveTextContent(/No activity over the last 24 hours/i);
  });
});
