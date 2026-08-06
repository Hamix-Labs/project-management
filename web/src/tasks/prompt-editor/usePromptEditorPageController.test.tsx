import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { PromptDocumentAdapter } from "./types";

const load = vi.fn();
const save = vi.fn(async () => undefined);

const adapter: PromptDocumentAdapter = {
  load: (...args) => load(...args),
  save,
};

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
    const { result } = renderHook(() => usePromptEditorPageController());

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
});
