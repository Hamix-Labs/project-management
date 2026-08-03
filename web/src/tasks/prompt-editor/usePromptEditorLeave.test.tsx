import { renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import { usePromptEditorLeave } from "./usePromptEditorLeave";

function wrapper({ children }: { children: ReactNode }) {
  return (
    <MemoryRouter future={ROUTER_FUTURE_FLAGS} initialEntries={["/prompt/draft/x"]}>
      {children}
    </MemoryRouter>
  );
}

describe("usePromptEditorLeave", () => {
  it("mounts under MemoryRouter without requiring a data router", () => {
    const htmlRef = { current: "<p>hi</p>" };
    const dirtyRef = { current: false };
    const { result } = renderHook(
      () =>
        usePromptEditorLeave({
          adapter: {
            load: vi.fn(),
            save: vi.fn(async () => undefined),
          },
          launch: {
            returnPath: "/",
            resumeCompose: true,
          },
          htmlRef,
          dirtyRef,
          setSessionError: vi.fn(),
          setLastSavedAt: vi.fn(),
          setSaving: vi.fn(),
        }),
      { wrapper },
    );
    expect(result.current.leavePending).toBe(false);
    expect(typeof result.current.leaveEditor).toBe("function");
    expect(typeof result.current.leaveWithoutSave).toBe("function");
  });
});
