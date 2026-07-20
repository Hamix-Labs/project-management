import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { beforeEach, describe, expect, it } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "../../lib/routerFutureFlags";
import { DEFAULT_DOCUMENT_TITLE } from "../../shared/useDocumentTitle";
import { setupAppTest } from "@/test/integration/appHarness";
import {
  taskEventGet,
  taskEventGetFlaky,
  taskEventUserResponsePatch,
  taskEventUserResponsePatchError,
} from "@/test/handlers/tasks";
import { server } from "@/test/server";
import { TaskEventDetailPage } from "./TaskEventDetailPage";

function renderEventPage(initialPath: string) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter
        future={ROUTER_FUTURE_FLAGS}
        initialEntries={[initialPath]}
      >
        <Routes>
          <Route
            path="/tasks/:taskId/events/:eventSeq"
            element={<TaskEventDetailPage />}
          />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("TaskEventDetailPage", () => {
  beforeEach(() => {
    setupAppTest();
  });

  it("shows event load error with retry and refetches successfully", async () => {
    const user = userEvent.setup();
    server.use(
      taskEventGetFlaky("t1", 2, {
        task_id: "t1",
        seq: 2,
        at: "2026-03-01T10:00:00.000Z",
        type: "message_added",
        by: "agent",
        data: {},
      }),
    );

    renderEventPage("/tasks/t1/events/2");

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /try again/i }));

    expect(
      await screen.findByRole("heading", { name: /event #2/i }),
    ).toBeInTheDocument();
  });

  it("loads one event and shows type, time, actor, overview fields, and raw JSON", async () => {
    const user = userEvent.setup();
    server.use(
      taskEventGet("t1", 2, {
        task_id: "t1",
        seq: 2,
        at: "2026-03-01T10:00:00.000Z",
        type: "message_added",
        by: "agent",
        data: { from: "a", to: "b" },
      }),
    );

    renderEventPage("/tasks/t1/events/2");

    expect(
      await screen.findByRole("heading", { name: /event #2/i }),
    ).toBeInTheDocument();
    expect(document.title).toBe(
      `Event #2: Title or message updated · ${DEFAULT_DOCUMENT_TITLE}`,
    );
    expect(screen.getByText("t1")).toBeInTheDocument();
    const pill = document.querySelector(
      "code.task-timeline-type-pill[data-event-type='message_added']",
    );
    expect(pill).not.toBeNull();
    expect(screen.getByText(/event data/i)).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /^overview$/i })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByText("From")).toBeInTheDocument();
    expect(screen.getByText("a")).toBeInTheDocument();
    expect(screen.getByText("To")).toBeInTheDocument();
    expect(screen.getByText("b")).toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: /^raw json$/i }));
    expect(screen.getByText(/"from"/)).toBeInTheDocument();
  });

  it("shows a response form when the event type needs user input", async () => {
    server.use(
      taskEventGet("t1", 2, {
        task_id: "t1",
        seq: 2,
        at: "2026-03-01T10:00:00.000Z",
        type: "approval_requested",
        by: "agent",
        data: {},
      }),
    );

    renderEventPage("/tasks/t1/events/2");

    expect(
      await screen.findByRole("heading", { name: /^add a message$/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("textbox", { name: /^add a message$/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("Agent needs input")).toHaveAttribute(
      "data-awaiting-response",
      "true",
    );
  });

  it("submits a user reply via PATCH and leaves awaiting state", async () => {
    const user = userEvent.setup();
    let patchBody: string | null = null;
    const baseEvent = {
      task_id: "t1",
      seq: 2,
      at: "2026-03-01T10:00:00.000Z",
      type: "approval_requested" as const,
      by: "agent" as const,
      data: {},
    };
    server.use(
      taskEventGet("t1", 2, baseEvent),
      taskEventUserResponsePatch("t1", 2, (body) => {
        patchBody = body;
      }, {
        ...baseEvent,
        user_response: "Looks good",
        user_response_at: "2026-03-01T10:05:00.000Z",
      }),
    );

    renderEventPage("/tasks/t1/events/2");

    const textbox = await screen.findByRole("textbox", {
      name: /^add a message$/i,
    });
    expect(screen.getByText("Agent needs input")).toHaveAttribute(
      "data-awaiting-response",
      "true",
    );

    await user.type(textbox, "Looks good");
    await user.click(screen.getByRole("button", { name: /^send$/i }));

    await waitFor(() => {
      expect(patchBody).toBe(JSON.stringify({ user_response: "Looks good" }));
    });
    expect(
      await screen.findByText("You replied — waiting on agent"),
    ).not.toHaveAttribute("data-awaiting-response");
    expect(textbox).toHaveValue("");
    expect(
      screen.getByRole("log", { name: /conversation on this event/i }),
    ).toHaveTextContent("Looks good");
  });

  it("shows an error alert when PATCH user reply fails", async () => {
    const user = userEvent.setup();
    server.use(
      taskEventGet("t1", 2, {
        task_id: "t1",
        seq: 2,
        at: "2026-03-01T10:00:00.000Z",
        type: "approval_requested",
        by: "agent",
        data: {},
      }),
      taskEventUserResponsePatchError("t1", 2, 500, "reply rejected"),
    );

    renderEventPage("/tasks/t1/events/2");

    const textbox = await screen.findByRole("textbox", {
      name: /^add a message$/i,
    });
    await user.type(textbox, "Looks good");
    await user.click(screen.getByRole("button", { name: /^send$/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/reply rejected/i);
    expect(screen.getByText("Agent needs input")).toHaveAttribute(
      "data-awaiting-response",
      "true",
    );
    expect(textbox).toHaveValue("Looks good");
  });

  it("does not mark awaiting when user has replied (legacy user_response)", async () => {
    server.use(
      taskEventGet("t1", 2, {
        task_id: "t1",
        seq: 2,
        at: "2026-03-01T10:00:00.000Z",
        type: "approval_requested",
        by: "agent",
        data: {},
        user_response: "Approved",
        user_response_at: "2026-03-01T10:05:00.000Z",
      }),
    );

    renderEventPage("/tasks/t1/events/2");

    await screen.findByRole("heading", { name: /^add a message$/i });
    expect(screen.getByText("You replied — waiting on agent")).not.toHaveAttribute(
      "data-awaiting-response",
    );
    const convo = screen.getByRole("log", {
      name: /conversation on this event/i,
    });
    expect(convo).toHaveTextContent("Approved");
    const sentAt = within(convo).getByRole("time");
    expect(sentAt).toHaveAttribute("dateTime", "2026-03-01T10:05:00.000Z");
    expect(
      screen.getByRole("textbox", { name: /^add a message$/i }),
    ).toHaveValue("");
  });

  it("shows a client-side message when seq is not a positive integer", () => {
    renderEventPage("/tasks/t1/events/nope");

    expect(
      screen.getByText(/invalid event sequence/i),
    ).toBeInTheDocument();
  });

  it("shows Overview tab and usage for phase_completed events", async () => {
    const user = userEvent.setup();
    server.use(
      taskEventGet("t1", 9, {
        task_id: "t1",
        seq: 9,
        at: "2026-04-19T15:46:35.000Z",
        type: "phase_completed",
        by: "agent",
        data: {
          phase: "execute",
          status: "succeeded",
          cycle_id: "d60be771-3b1c-49a1-8710-2a11a963455a",
          phase_seq: 2,
          summary: "Hello **world**",
          details: {
            duration_ms: 1000,
            request_id: "r1",
            usage: { inputTokens: 100, outputTokens: 10 },
          },
        },
      }),
    );

    renderEventPage("/tasks/t1/events/9");

    expect(await screen.findByRole("tab", { name: /^overview$/i })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByText("execute")).toBeInTheDocument();
    expect(screen.getByText("succeeded")).toBeInTheDocument();
    expect(screen.getByText("100")).toBeInTheDocument();
    expect(screen.getByText("world")).toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: /^raw json$/i }));
    expect(screen.getByText(/"phase"/)).toBeInTheDocument();
  });

  it("rejects partially numeric seq values instead of coercing", () => {
    renderEventPage("/tasks/t1/events/2oops");

    expect(
      screen.getByText(/invalid event sequence/i),
    ).toBeInTheDocument();
  });
});
