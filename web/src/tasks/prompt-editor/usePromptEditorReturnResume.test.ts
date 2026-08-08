/** @vitest-environment jsdom */
import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import {
  writePromptEditorReturn,
  consumePromptEditorReturn,
} from "./promptEditorSession";
import { usePromptEditorReturnResume } from "./usePromptEditorReturnResume";

describe("usePromptEditorReturnResume", () => {
  it("applies title and html when resuming compose", () => {
    sessionStorage.clear();
    writePromptEditorReturn({
      resumeCompose: true,
      returnPath: "/",
      html: "<p>body</p>",
      title: "From IDE",
    });

    const setNewPrompt = vi.fn();
    const setNewTitle = vi.fn();
    const resumeComposeFromPromptEditor = vi.fn();

    renderHook(() =>
      usePromptEditorReturnResume({
        setNewPrompt,
        setNewTitle,
        resumeComposeFromPromptEditor,
        promptEditorSuspended: true,
      }),
    );

    expect(setNewPrompt).toHaveBeenCalledWith("<p>body</p>");
    expect(setNewTitle).toHaveBeenCalledWith("From IDE");
    expect(resumeComposeFromPromptEditor).toHaveBeenCalled();
    expect(consumePromptEditorReturn()).toBeNull();
  });
});
