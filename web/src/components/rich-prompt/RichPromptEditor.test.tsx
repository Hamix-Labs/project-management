import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StrictMode, type ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ModalStackProvider } from "@/shared/ModalStackContext";
import { RichPromptEditor } from "./RichPromptEditor";

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  });
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <ModalStackProvider>{children}</ModalStackProvider>
      </QueryClientProvider>
    );
  }
  return { Wrapper };
}

describe("RichPromptEditor", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ status: "ok", checks: {} }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
      ),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("renders the formatting toolbar", () => {
    const { Wrapper } = makeWrapper();
    render(
      <RichPromptEditor id="rich-1" value="<p></p>" onChange={vi.fn()} />,
      { wrapper: Wrapper },
    );
    expect(
      screen.getByRole("toolbar", { name: /text formatting/i }),
    ).toBeInTheDocument();
  });

  it("mounts under StrictMode without throwing (TipTap immediatelyRender)", async () => {
    const { Wrapper } = makeWrapper();
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    render(
      <StrictMode>
        <RichPromptEditor id="rich-strict" value="<p></p>" onChange={vi.fn()} />
      </StrictMode>,
      { wrapper: Wrapper },
    );
    expect(
      await screen.findByRole("toolbar", { name: /text formatting/i }),
    ).toBeInTheDocument();
    expect(
      errorSpy.mock.calls.some((c) =>
        String(c[0] ?? "").includes("reading 'commands'"),
      ),
    ).toBe(false);
    errorSpy.mockRestore();
  });

  it("renders the empty-block placeholder copy", async () => {
    const { Wrapper } = makeWrapper();
    render(
      <RichPromptEditor id="rich-placeholder" value="<p></p>" onChange={vi.fn()} />,
      { wrapper: Wrapper },
    );
    await waitFor(() => {
      const paragraph = document.querySelector(
        "#rich-placeholder p",
      ) as HTMLElement | null;
      expect(paragraph?.getAttribute("data-placeholder")).toBe(
        "Press Space for AI or / for commands",
      );
    });
  });

  it("opens the inline composer and calls onAiTrigger on submit", async () => {
    const { Wrapper } = makeWrapper();
    const onAiTrigger = vi.fn();
    const user = userEvent.setup();
    render(
      <RichPromptEditor
        id="rich-space"
        value="<p></p>"
        onChange={vi.fn()}
        onAiTrigger={onAiTrigger}
      />,
      { wrapper: Wrapper },
    );

    const editorEl = document.querySelector<HTMLElement>("#rich-space");
    expect(editorEl).not.toBeNull();
    editorEl!.focus();
    await user.keyboard(" ");

    const composer = await screen.findByRole("dialog", {
      name: /ai assistant composer/i,
    });
    const input = screen.getByLabelText(/message the assistant/i);
    await user.type(input, "polish this brief");
    await user.keyboard("{Enter}");

    expect(onAiTrigger).toHaveBeenLastCalledWith("polish this brief");
    await waitFor(() => {
      expect(composer).not.toBeInTheDocument();
    });
  });

  it("closes the composer on Escape without invoking onAiTrigger a second time", async () => {
    const { Wrapper } = makeWrapper();
    const onAiTrigger = vi.fn();
    const user = userEvent.setup();
    render(
      <RichPromptEditor
        id="rich-escape"
        value="<p></p>"
        onChange={vi.fn()}
        onAiTrigger={onAiTrigger}
      />,
      { wrapper: Wrapper },
    );

    const editorEl = document.querySelector<HTMLElement>("#rich-escape");
    editorEl!.focus();
    await user.keyboard(" ");

    const composer = await screen.findByRole("dialog", {
      name: /ai assistant composer/i,
    });
    await user.keyboard("{Escape}");
    await waitFor(() => {
      expect(composer).not.toBeInTheDocument();
    });
    expect(onAiTrigger).toHaveBeenCalledTimes(1);
    expect(onAiTrigger).toHaveBeenLastCalledWith("");
  });
});
