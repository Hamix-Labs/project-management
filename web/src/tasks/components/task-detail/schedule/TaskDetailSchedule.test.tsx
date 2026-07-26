import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AppSettings } from "@/api/settings";
import { settingsQueryKeys } from "@/settings/settingsQueryKeys";
import type { Status } from "@/types";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import { APP_SETTINGS_DEFAULTS } from "@/test/settingsDefaults";
import { TaskDetailSchedule } from "./TaskDetailSchedule";

const isUiFeatureOmitted = vi.hoisted(() => vi.fn((_feature: string) => false));

vi.mock("@/launch/omittedFeatures", () => ({
  isUiFeatureOmitted: (feature: string) => isUiFeatureOmitted(feature),
}));

const NY_SETTINGS: AppSettings = {
  ...APP_SETTINGS_DEFAULTS,
  ...TASK_TEST_DEFAULTS,
  display_timezone: "America/New_York",
};

function createWrapper(settings: AppSettings = NY_SETTINGS) {
  const qc = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0, staleTime: Infinity },
    },
  });
  qc.setQueryData(settingsQueryKeys.app(), settings);
  function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
  }
  return { Wrapper };
}

function renderPanel(opts: {
  status?: Status;
  pickup?: string | null;
} = {}) {
  const { Wrapper } = createWrapper();
  const task = {
    status: opts.status ?? ("ready" as Status),
    pickup_not_before: opts.pickup ?? undefined,
  };
  return render(
    <Wrapper>
      <TaskDetailSchedule task={task} />
    </Wrapper>,
  );
}

describe("TaskDetailSchedule (read-only)", () => {
  beforeEach(() => {
    isUiFeatureOmitted.mockImplementation(() => false);
  });

  it("hides empty copy for terminal tasks without a pickup", () => {
    const { container } = renderPanel({ status: "done", pickup: null });
    expect(container).toBeEmptyDOMElement();
    expect(screen.queryByTestId("task-detail-status")).toBeNull();
    expect(screen.queryByText(/no pickup scheduled/i)).toBeNull();
  });

  it("renders a schedule row formatted in the app timezone", () => {
    renderPanel({
      status: "ready",
      pickup: "2026-04-22T13:00:00Z",
    });
    const row = screen.getByTestId("task-detail-schedule-badge");
    expect(row).toHaveClass("task-detail-schedule-row");
    expect(row).toHaveTextContent(/scheduled/i);
    expect(row).toHaveTextContent(/09:00/);
    expect(row).toHaveAttribute(
      "aria-label",
      expect.stringMatching(/scheduled for pickup/i),
    );
  });

  it("shows empty copy for an unscheduled non-terminal task without action buttons", () => {
    renderPanel({ status: "ready", pickup: null });
    expect(screen.getByText(/no pickup scheduled/i)).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /schedule/i }),
    ).not.toBeInTheDocument();
  });

  it("shows read-only schedule for a running task without edit controls", () => {
    renderPanel({
      status: "running",
      pickup: "2026-04-22T13:00:00Z",
    });
    expect(screen.queryByTestId("task-detail-status")).toBeNull();
    expect(screen.getByTestId("task-detail-schedule-badge")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("does not show phase-complete timing on the schedule strip", () => {
    renderPanel({ status: "running", pickup: null });
    expect(screen.queryByTestId("task-detail-phase-complete")).toBeNull();
    expect(screen.queryByTestId("task-detail-phase-complete-pending")).toBeNull();
    expect(screen.queryByTestId("task-detail-phase-duration")).toBeNull();
  });

  it("returns null when launch omits schedule", () => {
    isUiFeatureOmitted.mockImplementation((feature) => feature === "schedule");
    const { container } = renderPanel({
      status: "ready",
      pickup: "2026-04-22T13:00:00Z",
    });
    expect(container).toBeEmptyDOMElement();
  });

  it("renders a single schedule row when a pickup time is set", () => {
    const { container } = renderPanel({
      status: "ready",
      pickup: "2026-04-22T13:00:00Z",
    });
    const rows = container.querySelectorAll(".task-detail-schedule-row");
    expect(rows).toHaveLength(1);
    expect(screen.getByTestId("task-detail-schedule-badge")).toBeInTheDocument();
  });
});
