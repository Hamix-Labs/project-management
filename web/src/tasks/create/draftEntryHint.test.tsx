import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { listCursorModelsOk } from "@/test/handlers/settings";
import {
  appDefaultHandlers,
  renderTasksHome,
  setupAppTest,
} from "@/test/integration/appHarness";
import { draftsListPending } from "@/test/handlers/drafts";
import { server } from "@/test/server";

describe("draft entry hints", () => {
  beforeEach(() => {
    setupAppTest();
    server.use(...appDefaultHandlers());
  });

  it("stays on the list while drafts load and goes to compose when none exist", async () => {
    const user = userEvent.setup();
    const [pendingHandler, deferred] = draftsListPending();
    server.use(pendingHandler);

    renderTasksHome();
    await screen.findByText("No tasks yet");
    await user.click(screen.getByRole("button", { name: /\+?\s*new task/i }));
    expect(screen.getByRole("button", { name: /\+?\s*new task/i })).toHaveAttribute(
      "aria-busy",
      "true",
    );
    expect(screen.queryByText(/loading drafts/i)).not.toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: /resume a draft or start fresh/i }),
    ).not.toBeInTheDocument();

    await deferred.resolve(HttpResponse.json({ drafts: [] }));
    expect(
      await screen.findByRole("heading", { name: /^new task$/i }, { timeout: 10_000 }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: /resume a draft or start fresh/i }),
    ).not.toBeInTheDocument();
  });

  it("shows entry hint when drafts fail and opens fresh create form", async () => {
    const user = userEvent.setup();
    server.use(
      listCursorModelsOk(),
      http.get("/task-drafts", () =>
        HttpResponse.json({ error: "drafts unavailable" }, { status: 500 }),
      ),
    );

    renderTasksHome();
    await screen.findByText("No tasks yet");
    await user.click(screen.getByRole("button", { name: /\+?\s*new task/i }));

    expect(
      await screen.findByRole("heading", { name: /^new task$/i }, { timeout: 10_000 }),
    ).toBeInTheDocument();
    const draftsHintAlert = await screen.findByRole("alert");
    expect(draftsHintAlert).toHaveTextContent(
      /saved drafts are unavailable right now/i,
    );
    expect(
      screen.getByRole("button", { name: /retry loading drafts/i }),
    ).toBeInTheDocument();
  });

  it("retries draft loading from entry hint and opens draft picker when available", async () => {
    const user = userEvent.setup();
    let draftAttempts = 0;
    server.use(
      http.get("/task-drafts", () => {
        draftAttempts += 1;
        if (draftAttempts === 1) {
          return HttpResponse.json({ error: "drafts unavailable" }, { status: 500 });
        }
        return HttpResponse.json({
          drafts: [
            {
              id: "d1",
              name: "Recovered draft",
              created_at: "2026-04-07T10:00:00Z",
              updated_at: "2026-04-07T10:05:00Z",
            },
          ],
        });
      }),
    );

    renderTasksHome();
    await screen.findByText("No tasks yet");
    await user.click(screen.getByRole("button", { name: /\+?\s*new task/i }));
    await screen.findByRole("heading", { name: /^new task$/i }, { timeout: 10_000 });
    await user.click(screen.getByRole("button", { name: /retry loading drafts/i }));

    expect(
      await screen.findByRole("heading", { name: /resume a draft or start fresh/i }),
    ).toBeInTheDocument();
  });
});
