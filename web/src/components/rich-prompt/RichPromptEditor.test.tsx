import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
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
});
