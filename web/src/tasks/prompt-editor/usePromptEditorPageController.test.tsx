import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { PromptDocumentAdapter } from "./types";

const load = vi.fn();
const save = vi.fn(async () => undefined);
const saveName = vi.fn(async () => undefined);

const adapter: PromptDocumentAdapter = {
  load: (...args) => load(...args),
  save,
  saveName,
};

const queryClient = new QueryClient();

function wrapper({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

vi.mock("./usePromptEditorRouteAdapter", () => ({
  usePromptEditorRouteAdapter: () => ({
    sourceKind: "draft",
    sourceId: "d1",
    kindOk: true,
    launch: { title: "Brief", returnPath: "/" },
    adapter,
  }),
}));

vi.mock("./usePromptEditorAutosave", () => ({
  usePromptEditorAutosave: () => ({ retrySave: vi.fn() }),
}));

vi.mock("./usePromptEditorLeave", () => ({
  usePromptEditorLeave: () => ({
    leaveEditor: vi.fn(),
    leaveWithoutSave: vi.fn(),
    leavePending: false,
  }),
}));

import { usePromptEditorPageController } from "./usePromptEditorPageController";

describe("usePromptEditorPageController dirty load commit", () => {
  beforeEach(() => {
    load.mockReset();
    save.mockReset();
    load.mockResolvedValue({ html: "<p>server one snapshot</p>" });
  });

  it("does not overwrite html when a late load commits while dirty", async () => {
    const { result } = renderHook(() => usePromptEditorPageController(), {
      wrapper,
    });

    await waitFor(() => expect(result.current.ready).toBe(true));
    expect(result.current.html).toBe("<p>server one snapshot</p>");

    act(() => {
      result.current.onChange("<p>local typed words here now</p>");
    });
    expect(result.current.html).toBe("<p>local typed words here now</p>");

    load.mockResolvedValue({ html: "<p>server two overwrite attempt</p>" });
    act(() => {
      result.current.retryLoad();
    });

    await waitFor(() => expect(result.current.ready).toBe(true));
    expect(result.current.html).toBe("<p>local typed words here now</p>");
  });

  it("updates wordCountLabel when onChange receives new html", async () => {
    const { result } = renderHook(() => usePromptEditorPageController(), {
      wrapper,
    });

    await waitFor(() => expect(result.current.ready).toBe(true));
    expect(result.current.wordCountLabel).toBe("~3 words");

    act(() => {
      result.current.onChange("<p>one two three four five</p>");
    });

    expect(result.current.wordCountLabel).toBe("~5 words");
  });
});
