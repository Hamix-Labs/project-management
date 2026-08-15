import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import {
  draftGet,
  draftsList,
  draftsListPending,
} from "@/test/handlers/drafts";
import {
  appDefaultHandlers,
  renderTasksAt,
  renderTasksHome,
  setupAppTest,
} from "@/test/integration/appHarness";
import { server } from "@/test/server";

const savedDraft = {
  id: "draft-saved",
  name: "Saved draft",
  created_at: "2026-04-07T10:00:00Z",
  updated_at: "2026-04-07T10:05:00Z",
};

describe("draft resume from home", () => {
  beforeEach(() => {
    setupAppTest();
    server.use(...appDefaultHandlers());
  });

  it("opens the resume modal on the list when drafts exist and stays there until pick", async () => {
    const user = userEvent.setup();
    server.use(draftsList([savedDraft]));

    renderTasksHome();
    await screen.findByText("No tasks yet");
    await user.click(screen.getByRole("button", { name: /\+?\s*new task/i }));

    expect(
      await screen.findByRole("heading", {
        name: /resume a draft or start fresh/i,
      }),
    ).toBeInTheDocument();
    expect(screen.getByText("No tasks yet")).toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: /^new task$/i }),
    ).not.toBeInTheDocument();
  });

  it("navigates to compose with no resume modal when there are no drafts", async () => {
    const user = userEvent.setup();

    renderTasksHome();
    await screen.findByText("No tasks yet");
    await user.click(screen.getByRole("button", { name: /\+?\s*new task/i }));

    expect(
      await screen.findByRole("heading", { name: /^new task$/i }, { timeout: 10_000 }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: /resume a draft or start fresh/i }),
    ).not.toBeInTheDocument();
  });

  it("does not auto-open the resume modal on a direct /tasks/new visit", async () => {
    server.use(draftsList([savedDraft]));

    renderTasksAt(["/tasks/new"]);

    expect(
      await screen.findByRole("heading", { name: /^new task$/i }, { timeout: 10_000 }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: /resume a draft or start fresh/i }),
    ).not.toBeInTheDocument();
  });

  it("resumes a picked draft on compose via the draft query param", async () => {
    const user = userEvent.setup();
    server.use(
      draftsList([savedDraft]),
      draftGet("draft-saved", {
        name: "Saved draft",
        created_at: savedDraft.created_at,
        updated_at: savedDraft.updated_at,
        payload: {
          title: "Resumed from list",
          initial_prompt: "<p>Body</p>",
          priority: "medium",
          checklist_items: [],
        },
      }),
    );

    renderTasksHome();
    await screen.findByText("No tasks yet");
    await user.click(screen.getByRole("button", { name: /\+?\s*new task/i }));
    await user.click(
      await screen.findByRole("button", { name: /resume: saved draft/i }),
    );

    const heading = await screen.findByRole(
      "heading",
      { name: /^new task$/i },
      { timeout: 10_000 },
    );
    const page = heading.closest(".task-compose-page");
    expect(page).toBeInstanceOf(HTMLElement);
    await waitFor(() => {
      expect(
        within(page as HTMLElement).getByLabelText(/^title$/i),
      ).toHaveValue("Resumed from list");
    });
  });

  it("keeps the user on the list while drafts load, then opens the picker", async () => {
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
    expect(screen.getByText("No tasks yet")).toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: /resume a draft or start fresh/i }),
    ).not.toBeInTheDocument();
    expect(screen.queryByText(/loading drafts/i)).not.toBeInTheDocument();

    await deferred.resolve(HttpResponse.json({ drafts: [savedDraft] }));

    expect(
      await screen.findByRole("heading", {
        name: /resume a draft or start fresh/i,
      }),
    ).toBeInTheDocument();
    expect(screen.getByText("No tasks yet")).toBeInTheDocument();
  });
});
