import { describe, expect, it } from "vitest";
import { resolveEditorTitle } from "./resolveEditorTitle";

describe("resolveEditorTitle", () => {
  it("compose uses form title or Untitled task", () => {
    expect(resolveEditorTitle("compose", {})).toBe("Untitled task");
    expect(resolveEditorTitle("compose", { formTitle: "  " })).toBe(
      "Untitled task",
    );
    expect(resolveEditorTitle("compose", { formTitle: " Ship it " })).toBe(
      "Ship it",
    );
  });

  it("edit-task and polish use task name or Untitled task", () => {
    expect(resolveEditorTitle("edit-task", {})).toBe("Untitled task");
    expect(resolveEditorTitle("edit-task", { taskName: "Fix CI" })).toBe(
      "Fix CI",
    );
    expect(resolveEditorTitle("polish", { taskName: "  " })).toBe(
      "Untitled task",
    );
    expect(resolveEditorTitle("polish", { taskName: "Polish me" })).toBe(
      "Polish me",
    );
  });

  it("template uses template name or Untitled template", () => {
    expect(resolveEditorTitle("template", {})).toBe("Untitled template");
    expect(
      resolveEditorTitle("template", { templateName: "  Agent brief  " }),
    ).toBe("Agent brief");
  });
});
