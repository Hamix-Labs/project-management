import { screen, within, act, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  draftCreate,
  draftCreateCapture,
  draftsListEmpty,
} from "@/test/handlers/drafts";
import { listCursorModelsOk } from "@/test/handlers/settings";
import { openNewTaskModal } from "@/test/integration/createModalHelpers";
import {
  appDefaultHandlers,
  renderTasksHome,
  setupAppTest,
} from "@/test/integration/appHarness";
import { server } from "@/test/server";

describe("draft autosave on create modal", () => {
  beforeEach(() => {
    setupAppTest();
    server.use(
      ...appDefaultHandlers(),
      listCursorModelsOk(),
      draftsListEmpty(),
    );
  });

  it("keeps draft autosave failures inside the modal", async () => {
    const user = userEvent.setup();
    server.use(draftCreate(404, { error: "Not Found" }));

    renderTasksHome();
    await screen.findByText("No tasks yet");

    const dialog = await openNewTaskModal(user);
    await user.type(within(dialog).getByLabelText(/^title$/i), "Autosave test");

    expect(
      await within(dialog).findByText(
        /Draft autosave failed\. You can still create the task\./i,
      ),
    ).toBeInTheDocument();
    expect(screen.queryByRole("alert")).toBeNull();
  });

  // I8 — clicking Save draft is explicit intent, so it writes even when the
  // dirty gate says nothing changed. This deliberately reverses the older
  // "manual save no-ops when clean" contract: that gate is what made an
  // unfingerprinted field (repository_id) impossible to save by hand.
  it("submits an explicit save even when the draft looks unchanged", async () => {
    const user = userEvent.setup();
    const draftSaves: string[] = [];
    server.use(
      draftCreateCapture(
        (body) => draftSaves.push(body),
        { status: 201, body: { id: "d1", name: "Untitled draft" } },
      ),
    );

    renderTasksHome();
    await screen.findByText("No tasks yet");

    const dialog = await openNewTaskModal(user);
    await user.click(within(dialog).getByRole("button", { name: /^save draft$/i }));

    await waitFor(() => expect(draftSaves).toHaveLength(1));
    expect(JSON.parse(draftSaves[0]).payload).toHaveProperty("repository_id");
    expect(
      within(dialog).queryByText(
        /Draft autosave failed\. You can still create the task\./i,
      ),
    ).toBeNull();
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("clears prior autosave error when create modal is reopened", async () => {
    const user = userEvent.setup();
    server.use(draftCreate(404, { error: "Not Found" }));

    renderTasksHome();
    await screen.findByText("No tasks yet");

    const firstDialog = await openNewTaskModal(user);
    await user.type(within(firstDialog).getByLabelText(/^title$/i), "trigger autosave");
    expect(
      await within(firstDialog).findByText(
        /Draft autosave failed\. You can still create the task\./i,
      ),
    ).toBeInTheDocument();

    await user.click(within(firstDialog).getByRole("button", { name: /^cancel$/i }));

    const secondDialog = await openNewTaskModal(user);
    expect(
      within(secondDialog).queryByText(
        /Draft autosave failed\. You can still create the task\./i,
      ),
    ).toBeNull();
  });

  it("does not autosave untouched fresh drafts", async () => {
    const user = userEvent.setup();
    const draftSaves: string[] = [];
    server.use(
      draftCreateCapture(
        (body) => draftSaves.push(body),
        { status: 201, body: { id: "d1", name: "Untitled draft" } },
      ),
    );

    renderTasksHome();
    await screen.findByText("No tasks yet");

    const dialog = await openNewTaskModal(user);
    expect(dialog).toBeInTheDocument();

    await waitFor(() => {
      const repo = within(dialog).getByRole("combobox", { name: /^repository$/i });
      expect(repo).not.toHaveTextContent(/^select repository$/i);
    });

    vi.useFakeTimers();
    try {
      await act(async () => {
        await vi.advanceTimersByTimeAsync(1200);
      });
      expect(draftSaves).toHaveLength(0);
    } finally {
      vi.useRealTimers();
    }
  });
});
