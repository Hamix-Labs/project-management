import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, type RequestHandler } from "msw";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { AppSettings, AppSettingsPatch } from "@/api/settings";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import {
  appSettingsOk,
  appSettingsPatchError,
  appSettingsPatchOk,
  appSettingsPatchPending,
  listCursorModelsOk,
  probeCursorFail,
  probeCursorOk,
} from "@/test/handlers/settings";
import { server } from "@/test/server";
import { APP_SETTINGS_DEFAULTS } from "@/test/settingsDefaults";
import { SettingsPage } from "./SettingsPage";

/** Settings page fixture: filled CLI path + updated_at for form hydration. */
function pageSettings(overrides: Partial<AppSettings> = {}): Partial<AppSettings> {
  return {
    cursor_bin: "/usr/local/bin/cursor-agent",
    sse_replay_enabled: true,
    updated_at: "2026-04-18T12:00:00Z",
    ...overrides,
  };
}

function usePageHandlers(...extra: RequestHandler[]) {
  server.use(
    appSettingsOk(pageSettings()),
    listCursorModelsOk({
      binary_path: "/usr/local/bin/cursor-agent",
      models: [{ id: "auto", label: "Auto" }],
    }),
    ...extra,
  );
}

/** Edit the Cursor CLI path field in Agent settings. */
async function editCursorBin(value: string) {
  const cursorBin = await screen.findByLabelText(/^CLI path$/);
  await userEvent.clear(cursorBin);
  await userEvent.type(cursorBin, value);
  return cursorBin;
}

function renderPage(options?: { initialEntry?: string }) {
  const { initialEntry = "/settings" } = options ?? {};
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter future={ROUTER_FUTURE_FLAGS} initialEntries={[initialEntry]}>
        <SettingsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("SettingsPage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("scrolls to Cursor agent section after load when URL hash is #cursor-agent", async () => {
    const scrollIntoView = vi.fn();
    const prev = Object.getOwnPropertyDescriptor(Element.prototype, "scrollIntoView");
    Object.defineProperty(Element.prototype, "scrollIntoView", {
      configurable: true,
      writable: true,
      value: scrollIntoView,
    });
    usePageHandlers();

    try {
      renderPage({ initialEntry: "/settings#cursor-agent" });
      await screen.findByTestId("settings-cursor-model-select");
      await waitFor(() => expect(scrollIntoView).toHaveBeenCalled());
      expect(document.getElementById("cursor-agent")).not.toBeNull();
    } finally {
      if (prev) {
        Object.defineProperty(Element.prototype, "scrollIntoView", prev);
      } else {
        delete (Element.prototype as { scrollIntoView?: unknown }).scrollIntoView;
      }
    }
  });

  it("loads the settings row and pre-populates the form", async () => {
    usePageHandlers();

    renderPage();
    expect(await screen.findByLabelText(/^CLI path$/)).toHaveValue(
      "/usr/local/bin/cursor-agent",
    );
    expect(screen.getByLabelText(/Max execute duration/)).toHaveValue(0);
  });

  it("formats Last saved in the selected display timezone (explicit IANA)", async () => {
    server.use(
      appSettingsOk(
        pageSettings({
          display_timezone: "Europe/Berlin",
          // 10:00 UTC → 12:00 in Berlin on 2026-07-18 (CEST, UTC+2).
          updated_at: "2026-07-18T10:00:00Z",
        }),
      ),
      listCursorModelsOk({
        binary_path: "/usr/local/bin/cursor-agent",
        models: [{ id: "auto", label: "Auto" }],
      }),
    );

    renderPage();
    const chip = await screen.findByTestId("settings-last-updated");
    expect(chip.textContent).toMatch(/12:00/);
    // longOffset-style suffix for CEST (UTC+2), not a US abbreviation.
    expect(chip.textContent).toMatch(/GMT\+2|GMT\+02:00/i);
  });

  it("never PATCHes agent_paused from the form and does not render a badge for it", async () => {
    // `agent_paused` is owned by automation (agents/scripts hitting
    // PATCH /settings directly). The SettingsPage must:
    //   1. Not surface a "status" row for it — the top-bar
    //      live agent status chrome (when wired) is the single source of live agent
    //      status, and duplicating it on a configuration form
    //      confused operators (read-only row mixed into an editable
    //      form) and didn't generalize to multi-agent anyway.
    //   2. Never include agent_paused in the diff sent on Save, even
    //      after the GET response changes — otherwise saving an
    //      unrelated field would race-clobber a concurrent script
    //      that just paused the agent.
    const patches: AppSettingsPatch[] = [];
    server.use(
      appSettingsOk(pageSettings({ agent_paused: true })),
      listCursorModelsOk({
        binary_path: "/usr/local/bin/cursor-agent",
        models: [{ id: "auto", label: "Auto" }],
      }),
      appSettingsPatchOk(
        pageSettings({
          agent_paused: true,
          cursor_bin: "/usr/local/bin/cursor-agent-2",
          updated_at: "2026-04-18T12:34:00Z",
        }),
        (body) => {
          expect(body).not.toHaveProperty("agent_paused");
          patches.push(body);
        },
      ),
    );

    renderPage();

    // Wait for the form to hydrate, then assert the (retired) status
    // row is gone and no stray "Paused"/"Running" pill bled through.
    await screen.findByLabelText(/^CLI path$/);
    expect(
      screen.queryByTestId("settings-agent-paused-status"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("settings-agent-paused"),
    ).not.toBeInTheDocument();
    expect(screen.queryByText(/Agent pause status/i)).not.toBeInTheDocument();

    // Editing an unrelated field and saving must still succeed
    // without the patch including agent_paused.
    const cursorBin = screen.getByLabelText(/^CLI path$/);
    await userEvent.clear(cursorBin);
    await userEvent.type(cursorBin, "/usr/local/bin/cursor-agent-2");

    const saveButton = screen.getByRole("button", { name: /Save changes/i });
    await userEvent.click(saveButton);

    await waitFor(() => {
      expect(patches.length).toBe(1);
    });
  });

  it("PATCHes only the changed fields and updates form state on success", async () => {
    const patches: AppSettingsPatch[] = [];
    usePageHandlers(
      appSettingsPatchOk(
        pageSettings({
          cursor_bin: "/opt/local/bin/cursor-agent-2",
          updated_at: "2026-04-18T12:30:00Z",
        }),
        (body) => {
          expect(Object.keys(body)).toEqual(["cursor_bin"]);
          expect(body.cursor_bin).toBe("/opt/local/bin/cursor-agent-2");
          patches.push(body);
        },
      ),
    );

    renderPage();
    await editCursorBin("/opt/local/bin/cursor-agent-2");

    const saveBtn = screen.getByRole("button", { name: /Save changes/ });
    expect(saveBtn).not.toBeDisabled();
    await userEvent.click(saveBtn);

    await waitFor(() => expect(screen.getByTestId("settings-status")).toHaveTextContent(
      /saved/i,
    ));
    expect(patches.length).toBe(1);
  });

  it(
    "auto-dismisses the success banner after a few seconds",
    async () => {
      usePageHandlers(
        appSettingsPatchOk(
          pageSettings({
            cursor_bin: "/opt/local/bin/cursor-agent-2",
            updated_at: "2026-04-18T12:30:00Z",
          }),
        ),
      );

      renderPage();
      await editCursorBin("/opt/local/bin/cursor-agent-2");
      await userEvent.click(screen.getByRole("button", { name: /Save changes/ }));

      await waitFor(() =>
        expect(screen.getByTestId("settings-status")).toHaveTextContent(/saved/i),
      );

      await waitFor(
        () => expect(screen.queryByTestId("settings-status")).not.toBeInTheDocument(),
        { timeout: 6_000 },
      );
    },
    12_000,
  );

  it("disables Save when no fields have changed", async () => {
    usePageHandlers();
    renderPage();
    const saveBtn = await screen.findByRole("button", { name: /Save changes/ });
    expect(saveBtn).toBeDisabled();
  });

  it("calls /settings/probe-cursor and shows the version on success", async () => {
    usePageHandlers(probeCursorOk({ version: "2026.04" }));

    renderPage();
    const probeBtn = await screen.findByRole("button", { name: /Test binary/ });
    await userEvent.click(probeBtn);
    await waitFor(() =>
      expect(screen.getByTestId("settings-status")).toHaveTextContent(/2026\.04/),
    );
  });

  it("surfaces the PATH-resolved binary path when the field is blank, in both the status and the help text", async () => {
    // Without this, an operator who leaves the cursor-bin field blank
    // and clicks Test sees only "Cursor binary OK" and has no idea
    // which binary on PATH was actually exec'd. The /settings/probe-cursor
    // response now carries `binary_path`; the SPA must surface it.
    server.use(
      appSettingsOk(pageSettings({ cursor_bin: "" })),
      listCursorModelsOk({
        binary_path: "/usr/local/bin/cursor-agent",
        models: [{ id: "auto", label: "Auto" }],
      }),
      probeCursorOk({
        binary_path: "/opt/local/bin/cursor-agent",
        version: "2026.05",
      }),
    );

    renderPage();
    const probeBtn = await screen.findByRole("button", {
      name: /Test binary/,
    });
    await userEvent.click(probeBtn);
    await waitFor(() =>
      expect(screen.getByTestId("settings-status")).toHaveTextContent(
        /at \/opt\/local\/bin\/cursor-agent.*2026\.05/,
      ),
    );
    expect(
      screen.getByTestId("settings-resolved-cursor-bin"),
    ).toHaveTextContent("/opt/local/bin/cursor-agent");
  });

  it("surfaces probe failures via the error channel (role='alert', not role='status')", async () => {
    // Session #36 — probe `{ ok: false, error }` is semantically a
    // failure and must announce assertively to screen-readers, AND
    // must NOT appear in the success-styled `settings-status`
    // region (which is now reserved for actual successes).
    usePageHandlers(probeCursorFail("spawn ENOENT"));

    renderPage();
    const probeBtn = await screen.findByRole("button", { name: /Test binary/ });
    await userEvent.click(probeBtn);
    await waitFor(() =>
      expect(screen.getByTestId("settings-status-error")).toHaveTextContent(
        /spawn ENOENT/,
      ),
    );
    expect(screen.queryByTestId("settings-status")).not.toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent(/spawn ENOENT/);
  });

  it("surfaces patch errors via the error channel and does not show the success status", async () => {
    // Session #36 regression for the failed-PATCH case: previously
    // a 500 from PATCH /settings rendered through the same
    // `role="status"` channel as a successful save, which was a
    // direct a11y regression (assertive failure announced as polite).
    usePageHandlers(appSettingsPatchError(500, "internal: disk full"));

    renderPage();
    await editCursorBin("/opt/local/bin/cursor-agent-2");

    const saveBtn = screen.getByRole("button", { name: /Save changes/ });
    await userEvent.click(saveBtn);

    await waitFor(() =>
      expect(screen.getByTestId("settings-status-error")).toBeInTheDocument(),
    );
    // The error message text varies by API client error shape, but
    // the assertive alert region must always render so screen
    // readers announce the failure.
    expect(screen.getByRole("alert")).toBeInTheDocument();
    expect(screen.queryByTestId("settings-status")).not.toBeInTheDocument();
  });

  it("preserves in-flight typing on other fields when a PATCH resolves", async () => {
    // Session #37 regression: if the user edits field A, hits Save,
    // and then keeps typing in field B while the PATCH is still in
    // flight, the post-resolution `setForm(toFormState(next))` used
    // to clobber field B back to its server value (silently losing
    // the user's typing). The fix snapshots `formAtSubmit` and only
    // applies server truth per-field where the form hasn't been
    // re-edited since submit.
    const [pendingHandler, deferred] = appSettingsPatchPending();
    usePageHandlers(pendingHandler);

    renderPage();
    const cursorInput = await editCursorBin("/opt/local/bin/cursor-agent-2");

    const saveBtn = screen.getByRole("button", { name: /Save changes/ });
    await userEvent.click(saveBtn);

    // PATCH is in flight; the user keeps typing in max execute duration
    // (a field NOT in the submitted patch body).
    const maxInput = screen.getByLabelText(/Max execute duration/i);
    await userEvent.clear(maxInput);
    await userEvent.type(maxInput, "120");

    deferred.resolve(
      HttpResponse.json({
        ...APP_SETTINGS_DEFAULTS,
        ...pageSettings({
          cursor_bin: "/opt/local/bin/cursor-agent-2",
          updated_at: "2026-04-19T12:30:00Z",
        }),
      }),
    );

    await waitFor(() =>
      expect(screen.getByTestId("settings-status")).toHaveTextContent(/saved/i),
    );
    expect(maxInput).toHaveValue(120);
    expect(cursorInput).toHaveValue("/opt/local/bin/cursor-agent-2");
    expect(screen.getByRole("button", { name: /Save changes/ })).not.toBeDisabled();
  });

  it("preserves user re-edits to the same field while a PATCH is in flight", async () => {
    // Session #37 regression: the user types /A, hits Save, then
    // changes their mind to /B while the PATCH (carrying /A) is
    // still in flight. The PATCH resolves with /A. The user's
    // current intent is /B; clobbering back to /A would be silent
    // data loss + violate the user's mental model.
    const [pendingHandler, deferred] = appSettingsPatchPending();
    usePageHandlers(pendingHandler);

    renderPage();
    const cursorInput = await editCursorBin("/var/repos/A");

    await userEvent.click(screen.getByRole("button", { name: /Save changes/ }));

    await userEvent.clear(cursorInput);
    await userEvent.type(cursorInput, "/var/repos/B");

    deferred.resolve(
      HttpResponse.json({
        ...APP_SETTINGS_DEFAULTS,
        ...pageSettings({
          cursor_bin: "/var/repos/A",
          updated_at: "2026-04-19T12:30:00Z",
        }),
      }),
    );

    await waitFor(() =>
      expect(screen.getByTestId("settings-status")).toHaveTextContent(/saved/i),
    );
    expect(cursorInput).toHaveValue("/var/repos/B");
  });

  it("rejects negative max_run_duration_seconds", async () => {
    usePageHandlers();
    renderPage();
    const maxInput = await screen.findByLabelText(/Max execute duration/);
    await userEvent.clear(maxInput);
    await userEvent.type(maxInput, "-5");
    expect(screen.getByRole("alert")).toHaveTextContent(
      /non-negative integer/i,
    );
    expect(screen.getByRole("button", { name: /Save changes/ })).toBeDisabled();
  });
});
