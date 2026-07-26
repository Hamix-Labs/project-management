import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { TaskTimelineSection } from "./TaskTimelineSection";
import type { TimelineEvent } from "./timelineTypes";

vi.mock("../../hooks/useTasksActivity", () => ({
  useTasksActivity: () => ({
    events: [],
    total: 0,
    hasMore: false,
    truncated: false,
    loading: false,
    error: null,
    refetch: vi.fn(),
  }),
}));

const NOW = new Date("2026-07-25T18:00:00.000Z");

const FIXTURE_EVENTS: TimelineEvent[] = [
  {
    id: "ev-1",
    kind: "status-changed",
    category: "tasks",
    at: new Date(NOW.getTime() - 60 * 60 * 1000).toISOString(),
    title: "Status changed",
    highlight: "",
    detail: "Ready → Running",
    taskId: "f0000131-0000-4000-8000-000000000131",
    taskRef: "f0000131",
    seq: 3,
    taskTitle: "My task",
    taskPriority: "high",
    taskTags: ["api"],
  },
  {
    id: "ev-2",
    kind: "review-approved",
    category: "tasks",
    at: new Date(NOW.getTime() - 2 * 60 * 60 * 1000).toISOString(),
    title: "Review approved",
    highlight: "",
    detail: "Approval granted.",
    taskId: "f0000142-0000-4000-8000-000000000142",
    taskRef: "f0000142",
    seq: 5,
    taskTitle: "Another task",
    taskPriority: "low",
  },
];

function buildQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function renderSection(events?: TimelineEvent[]) {
  return render(
    <QueryClientProvider client={buildQueryClient()}>
      <MemoryRouter future={ROUTER_FUTURE_FLAGS}>
        <TaskTimelineSection
          view="timeline"
          events={events}
          now={NOW}
          projectFilterOptions={[{ id: "proj-1", name: "Alpha" }]}
          showProjectColumn
        />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("TaskTimelineSection", () => {
  it("renders timeline heading and range dropdown", () => {
    renderSection(FIXTURE_EVENTS);
    expect(
      screen.getByRole("heading", { name: "Timeline" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("tabpanel", { name: "Timeline" }),
    ).toHaveAttribute("id", "task-timeline-panel");
    expect(
      screen.getByRole("button", { name: /Last 7 days/i }),
    ).toBeInTheDocument();
  });

  it("does NOT render category filter pills", () => {
    renderSection(FIXTURE_EVENTS);
    expect(
      screen.queryByRole("button", { name: "All events" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Tasks" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Verification" }),
    ).not.toBeInTheDocument();
  });

  it("renders board-parity filters", () => {
    renderSection(FIXTURE_EVENTS);
    expect(screen.getByLabelText("Priority")).toBeInTheDocument();
    expect(screen.getByLabelText("Search titles")).toBeInTheDocument();
  });

  it("renders events from the override prop", () => {
    renderSection(FIXTURE_EVENTS);
    expect(screen.getByText("Status changed")).toBeInTheDocument();
    expect(screen.getByText("Review approved")).toBeInTheDocument();
  });

  it("filters events by title search", async () => {
    const user = userEvent.setup();
    renderSection(FIXTURE_EVENTS);
    await user.type(screen.getByLabelText("Search titles"), "Another");
    expect(screen.queryByText("Status changed")).not.toBeInTheDocument();
    expect(screen.getByText("Review approved")).toBeInTheDocument();
  });

  it("shows filter-empty copy when filters exclude everything", async () => {
    const user = userEvent.setup();
    renderSection(FIXTURE_EVENTS);
    await user.type(screen.getByLabelText("Search titles"), "zzzz-no-match");
    expect(
      screen.getByRole("status"),
    ).toHaveTextContent("No matching activity.");
  });

  it("shows empty state when no events", () => {
    renderSection([]);
    expect(
      screen.getByRole("status"),
    ).toHaveTextContent(/No activity over the last 7 days/i);
  });

  it("changes range via dropdown", async () => {
    const user = userEvent.setup();
    renderSection([]);
    await user.click(screen.getByRole("button", { name: /Last 7 days/i }));
    await user.click(screen.getByRole("option", { name: "Last 24 hours" }));
    expect(
      screen.getByRole("status"),
    ).toHaveTextContent(/No activity over the last 24 hours/i);
  });

  it("shows empty state with 'No activity to show.' for all-time range", async () => {
    const user = userEvent.setup();
    renderSection([]);
    await user.click(screen.getByRole("button", { name: /Last 7 days/i }));
    await user.click(screen.getByRole("option", { name: "All time" }));
    expect(
      screen.getByRole("status"),
    ).toHaveTextContent("No activity to show.");
  });
});
