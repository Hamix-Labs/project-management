import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { HttpResponse, http } from "msw";
import {
  draftAssistCreateSession,
  draftAssistSessionFrame,
  draftAssistStartRunOk,
  makeDraftAssistSessionFixture,
} from "@/test/handlers/draftAssist";
import { ensureMswListening } from "@/test/mswLifecycle";
import { server } from "@/test/server";
import {
  DraftAssistProvider,
  DraftAssistThread,
  useDraftAssistContext,
} from "./index";

ensureMswListening("bypass");

type NamedListener = (ev: { data?: string }) => void;

class MockEventSource {
  static instances: MockEventSource[] = [];
  static latest(): MockEventSource | null {
    return MockEventSource.instances[MockEventSource.instances.length - 1] ?? null;
  }
  onopen: (() => void) | null = null;
  onerror: (() => void) | null = null;
  readonly listeners = new Map<string, Set<NamedListener>>();
  close = vi.fn();
  constructor(public url: string) {
    MockEventSource.instances.push(this);
    queueMicrotask(() => this.onopen?.());
  }
  addEventListener(kind: string, cb: NamedListener) {
    let set = this.listeners.get(kind);
    if (!set) {
      set = new Set();
      this.listeners.set(kind, set);
    }
    set.add(cb);
  }
  dispatch(kind: string, payload: unknown) {
    const data = typeof payload === "string" ? payload : JSON.stringify(payload);
    const set = this.listeners.get(kind);
    if (set) for (const cb of set) cb({ data });
  }
}

beforeEach(() => {
  MockEventSource.instances = [];
  vi.stubGlobal("EventSource", MockEventSource);
  // Catch-all DELETE so unmount teardown doesn't log MSW warnings.
  server.use(
    http.delete(
      /\/draft-assist\/sessions\/[^/]+$/,
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
});

/** Small harness to drive the provider from tests. */
function ThreadHarness({
  initialPrompt = "",
  onPromptChange,
}: {
  initialPrompt?: string;
  onPromptChange?: (v: string) => void;
}) {
  return (
    <DraftAssistProvider
      getSnapshot={() => ({ title: "", prompt: initialPrompt })}
      worktreeId="wt-1"
      getPromptSnapshot={() => initialPrompt}
      onApplyPromptPatch={onPromptChange}
    >
      <TriggerControls />
      <DraftAssistThread />
    </DraftAssistProvider>
  );
}

function TriggerControls() {
  const ctx = useDraftAssistContext();
  return (
    <>
      <button type="button" onClick={() => ctx.open("Tighten this brief")}>
        open
      </button>
      <button type="button" onClick={() => ctx.send("Also add tests")}>
        send-follow-up
      </button>
    </>
  );
}

async function waitForEventSource(): Promise<MockEventSource> {
  await waitFor(() => {
    expect(MockEventSource.instances.length).toBeGreaterThan(0);
  });
  const es = MockEventSource.latest();
  if (!es) throw new Error("no EventSource opened");
  return es;
}

describe("DraftAssistThread — happy path", () => {
  it("open() shows user bubble, opens session, streams tokens, terminal Prompt updated", async () => {
    const user = userEvent.setup();
    const fixture = makeDraftAssistSessionFixture({ id: "sess-happy" });
    const applyPatch = vi.fn();
    server.use(
      draftAssistCreateSession(fixture),
      draftAssistStartRunOk(fixture.id, "run-happy"),
    );

    render(
      <ThreadHarness
        initialPrompt="<p>seed</p>"
        onPromptChange={applyPatch}
      />,
    );

    await user.click(screen.getByRole("button", { name: /^open$/ }));

    expect(await screen.findByText("Tighten this brief")).toBeInTheDocument();

    const es = await waitForEventSource();

    // Server hands out the session frame first.
    await act(async () => {
      es.dispatch("session", draftAssistSessionFrame(fixture.id, 1));
    });

    // Status → thinking.
    await act(async () => {
      es.dispatch("status", {
        id: 2,
        kind: "status",
        run_id: "run-happy",
        at: "2026-08-14T00:00:01Z",
        data: { status: "thinking" },
      });
    });
    expect(await screen.findByText("Thinking…")).toBeInTheDocument();

    // Streaming tokens accumulate into one assistant bubble.
    await act(async () => {
      es.dispatch("token", {
        id: 3,
        kind: "token",
        run_id: "run-happy",
        at: "2026-08-14T00:00:02Z",
        data: { delta: "Sure — " },
      });
      es.dispatch("token", {
        id: 4,
        kind: "token",
        run_id: "run-happy",
        at: "2026-08-14T00:00:02Z",
        data: { delta: "tightened." },
      });
    });

    const assistantMsg = await screen.findByTestId("draft-assist-assistant-message");
    expect(assistantMsg).toHaveTextContent("Sure — tightened.");

    // A patch frame updates the prompt in place.
    await act(async () => {
      es.dispatch("patch", {
        id: 5,
        kind: "patch",
        run_id: "run-happy",
        at: "2026-08-14T00:00:03Z",
        data: { op: "set", value: "<p>tight brief</p>", summary: "rewrote" },
      });
    });
    expect(applyPatch).toHaveBeenCalledWith("<p>tight brief</p>");

    // Terminal done → assistant bubble closes and status flips to Prompt updated.
    await act(async () => {
      es.dispatch("done", {
        id: 6,
        kind: "done",
        run_id: "run-happy",
        at: "2026-08-14T00:00:04Z",
        data: { status: "done" },
      });
    });
    expect(await screen.findByText("Prompt updated")).toBeInTheDocument();
  });
});

describe("DraftAssistThread — cancel", () => {
  it("Stop calls cancelRun and reflects cancelling → Assistant stopped", async () => {
    const user = userEvent.setup();
    const fixture = makeDraftAssistSessionFixture({ id: "sess-cancel" });
    let cancels = 0;
    server.use(
      draftAssistCreateSession(fixture),
      draftAssistStartRunOk(fixture.id, "run-cancel"),
      http.post(`/draft-assist/sessions/${fixture.id}/runs/run-cancel/cancel`, () => {
        cancels += 1;
        return HttpResponse.json(
          { run_id: "run-cancel", status: "cancelling" },
          { status: 202 },
        );
      }),
    );

    render(<ThreadHarness />);
    await user.click(screen.getByRole("button", { name: /^open$/ }));
    const es = await waitForEventSource();

    // Kick the run through status=streaming so Stop is visible.
    await act(async () => {
      es.dispatch("status", {
        id: 1,
        kind: "status",
        run_id: "run-cancel",
        at: "2026-08-14T00:00:01Z",
        data: { status: "streaming" },
      });
    });

    // Wait for run_accepted (POST /runs) so run id is registered.
    await waitFor(() => {
      expect(screen.getByTestId("draft-assist-stop")).toBeInTheDocument();
    });

    await user.click(screen.getByTestId("draft-assist-stop"));

    expect(await screen.findByText("Stopping…")).toBeInTheDocument();
    await waitFor(() => {
      expect(cancels).toBe(1);
    });

    // Terminal cancelled frame.
    await act(async () => {
      es.dispatch("done", {
        id: 2,
        kind: "done",
        run_id: "run-cancel",
        at: "2026-08-14T00:00:02Z",
        data: { status: "cancelled" },
      });
    });
    expect(await screen.findByText("Assistant stopped")).toBeInTheDocument();
  });
});

describe("DraftAssistThread — patch apply integration", () => {
  it("find_replace patch mutates prompt via onApplyPromptPatch", async () => {
    const user = userEvent.setup();
    const fixture = makeDraftAssistSessionFixture({ id: "sess-patch" });
    let current = "<p>hello world</p>";
    const onPatch = vi.fn((next: string) => {
      current = next;
    });
    server.use(
      draftAssistCreateSession(fixture),
      draftAssistStartRunOk(fixture.id, "run-p"),
    );

    render(
      <DraftAssistProvider
        getSnapshot={() => ({ prompt: current })}
        worktreeId="wt-1"
        getPromptSnapshot={() => current}
        onApplyPromptPatch={onPatch}
      >
        <TriggerControls />
        <DraftAssistThread />
      </DraftAssistProvider>,
    );
    await user.click(screen.getByRole("button", { name: /^open$/ }));
    const es = await waitForEventSource();

    await act(async () => {
      es.dispatch("patch", {
        id: 1,
        kind: "patch",
        run_id: "run-p",
        at: "2026-08-14T00:00:00Z",
        data: { op: "find_replace", find: "hello", value: "hi", summary: "casual" },
      });
    });

    expect(onPatch).toHaveBeenCalledWith("<p>hi world</p>");
    expect(await screen.findByTestId("draft-assist-patch-row")).toBeInTheDocument();
  });
});

describe("DraftAssistThread — a11y", () => {
  it("status region is aria-live polite; error region is assertive", async () => {
    const user = userEvent.setup();
    const fixture = makeDraftAssistSessionFixture({ id: "sess-a11y" });
    server.use(
      draftAssistCreateSession(fixture),
      draftAssistStartRunOk(fixture.id, "run-a11y"),
    );

    render(<ThreadHarness />);
    await user.click(screen.getByRole("button", { name: /^open$/ }));
    const es = await waitForEventSource();

    const statusRegion = screen.getByRole("status");
    expect(statusRegion).toHaveAttribute("aria-live", "polite");

    await act(async () => {
      es.dispatch("error", {
        id: 1,
        kind: "error",
        run_id: "run-a11y",
        at: "2026-08-14T00:00:01Z",
        data: { code: "boom", message: "connection lost" },
      });
    });

    // Both the inline error row (`role="alert"`) and the banner
    // (`role="alert"` + `aria-live="assertive"`) should be present.
    const alerts = await screen.findAllByRole("alert");
    expect(alerts.length).toBeGreaterThanOrEqual(1);
    const assertive = alerts.find(
      (el) => el.getAttribute("aria-live") === "assertive",
    );
    expect(assertive).toBeDefined();
  });

  it("Stop button is keyboard-reachable via Tab", async () => {
    const user = userEvent.setup();
    const fixture = makeDraftAssistSessionFixture({ id: "sess-kbd" });
    server.use(
      draftAssistCreateSession(fixture),
      draftAssistStartRunOk(fixture.id, "run-kbd"),
    );

    render(<ThreadHarness />);
    await user.click(screen.getByRole("button", { name: /^open$/ }));
    const es = await waitForEventSource();

    await act(async () => {
      es.dispatch("status", {
        id: 1,
        kind: "status",
        run_id: "run-kbd",
        at: "2026-08-14T00:00:01Z",
        data: { status: "streaming" },
      });
    });

    const stop = await screen.findByTestId("draft-assist-stop");
    stop.focus();
    expect(stop).toHaveFocus();
  });
});
