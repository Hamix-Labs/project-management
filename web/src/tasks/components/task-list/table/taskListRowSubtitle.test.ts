import { describe, expect, it } from "vitest";
import { statusListLabel } from "../../../task-display/statusListLabel";
import { taskListRowSubtitle } from "./taskListRowSubtitle";

describe("taskListRowSubtitle", () => {
  it("shows prompt preview when present", () => {
    expect(
      taskListRowSubtitle({
        promptPreview: "  Do the thing  ",
      }),
    ).toBe("Do the thing");
  });

  it("shows prompt preview even when project column is populated", () => {
    expect(
      taskListRowSubtitle({
        promptPreview: "Some prompt",
      }),
    ).toBe("Some prompt");
  });

  it("returns undefined when there is nothing to say", () => {
    expect(
      taskListRowSubtitle({
        promptPreview: "   ",
      }),
    ).toBeUndefined();
  });
});

describe("statusListLabel", () => {
  it("maps running to in-progress copy", () => {
    expect(statusListLabel("running")).toBe("Running");
  });
});
