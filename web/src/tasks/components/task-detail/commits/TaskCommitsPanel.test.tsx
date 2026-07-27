import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { TaskCommitsPanel } from "./TaskCommitsPanel";

type FetchInput = Parameters<typeof fetch>[0];

function createWrapper() {
  const qc = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0, staleTime: 0 },
    },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={qc}>
        <MemoryRouter future={ROUTER_FUTURE_FLAGS}>{children}</MemoryRouter>
      </QueryClientProvider>
    );
  };
}

const okJSON = (body: unknown) =>
  new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });

const reqUrl = (input: FetchInput): string =>
  typeof input === "string"
    ? input
    : input instanceof URL
      ? input.toString()
      : (input as Request).url;

function renderPanel(taskId = "task-1") {
  const Wrapper = createWrapper();
  return render(
    <Wrapper>
      <TaskCommitsPanel taskId={taskId} />
    </Wrapper>,
  );
}

describe("TaskCommitsPanel", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders loading skeleton inside the empty well", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(
      () => new Promise(() => {}),
    );
    const { container } = renderPanel();

    expect(await screen.findByTestId("task-commits-panel")).toBeInTheDocument();
    const well = container.querySelector(".task-commits-empty-well");
    expect(well).not.toBeNull();
    expect(well?.getAttribute("aria-busy")).toBe("true");
  });

  it("renders error state with retry", async () => {
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValue(new Response("boom", { status: 500 }));

    renderPanel();

    expect(await screen.findByRole("alert")).toHaveTextContent(/boom/i);
    await userEvent.click(screen.getByRole("button", { name: /Try again/i }));
    expect(fetchMock.mock.calls.length).toBeGreaterThan(1);
  });

  it("renders centered empty well with git-oriented copy", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = reqUrl(input);
      if (url.endsWith("/tasks/task-1/commits")) {
        return okJSON({ task_id: "task-1", commits: [] });
      }
      return new Response("not found", { status: 404 });
    });

    renderPanel();

    expect(
      await screen.findByTestId("task-commits-empty-well"),
    ).toBeInTheDocument();
    expect(screen.getByText("No commits indexed yet")).toBeInTheDocument();
    expect(
      screen.getByText(/Recorded when an agent run commits to git/i),
    ).toBeInTheDocument();
  });

  it("renders populated commits list newest-first", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = reqUrl(input);
      if (url.endsWith("/tasks/task-1/commits")) {
        return okJSON({
          task_id: "task-1",
          commits: [
            {
              seq: 2,
              cycle_id: "cycle-2",
              attempt_seq: 2,
              repo: "/repo",
              worktree: "/repo/.worktrees/task-1",
              branch: "feature/newer",
              sha: "def123def4567890abcdef1234567890abcdef12",
              message: "Newer commit",
              committed_at: "2026-04-18T12:00:00.000Z",
            },
            {
              seq: 1,
              cycle_id: "cycle-1",
              attempt_seq: 1,
              repo: "/repo",
              worktree: "/repo/.worktrees/task-1",
              branch: "main",
              sha: "abc123def4567890abcdef1234567890abcdef12",
              message: "Older commit",
              committed_at: "2026-04-18T11:00:00.000Z",
            },
          ],
        });
      }
      return new Response("not found", { status: 404 });
    });

    renderPanel();

    const list = await screen.findByTestId("task-commits-list");
    const items = within(list).getAllByRole("listitem");
    expect(items[0]).toHaveTextContent(/Newer commit/i);
    expect(items[1]).toHaveTextContent(/Older commit/i);
    expect(screen.queryByTestId("task-commits-empty-well")).not.toBeInTheDocument();
    expect(screen.getByTestId("task-commits-branch-line")).toHaveTextContent(
      "feature/newer",
    );
    expect(screen.queryByTestId("task-commits-context")).not.toBeInTheDocument();

    const details = document.querySelector(
      ".task-commits-panel .task-detail-collapsible",
    );
    expect(details).toHaveAttribute("open");
  });

  it("collapses commits body when the section header is toggled", async () => {
    const user = userEvent.setup();
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input) => {
      const url = reqUrl(input);
      if (url.endsWith("/tasks/task-1/commits")) {
        return okJSON({
          task_id: "task-1",
          commits: [
            {
              seq: 1,
              cycle_id: "cycle-1",
              attempt_seq: 1,
              repo: "/repo",
              worktree: "/repo/.worktrees/task-1",
              branch: "main",
              sha: "abc123def4567890abcdef1234567890abcdef12",
              message: "Initial commit",
              committed_at: "2026-04-18T11:00:00.000Z",
            },
          ],
        });
      }
      return new Response("not found", { status: 404 });
    });

    renderPanel();

    expect(await screen.findByText(/Initial commit/i)).toBeVisible();
    const details = document.querySelector(
      ".task-commits-panel .task-detail-collapsible",
    );
    expect(details).toHaveAttribute("open");

    await user.click(screen.getByRole("heading", { name: /^commits$/i }));
    expect(details).not.toHaveAttribute("open");
    expect(screen.queryByText(/Initial commit/i)).not.toBeVisible();
  });
});
