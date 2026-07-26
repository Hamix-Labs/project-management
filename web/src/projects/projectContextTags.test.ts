import { describe, expect, it } from "vitest";
import type { ProjectContextItem } from "@/types";
import {
  collectProjectContextTags,
  groupProjectContextByTag,
} from "./projectContextTags";

function item(
  overrides: Partial<ProjectContextItem> & Pick<ProjectContextItem, "id" | "tag">,
): ProjectContextItem {
  return {
    project_id: "p1",
    title: overrides.title ?? overrides.id,
    description: "",
    body: "body",
    created_by: "user",
    pinned: false,
    created_at: "2026-04-27T00:00:00Z",
    updated_at: "2026-04-27T00:00:00Z",
    ...overrides,
  };
}

describe("projectContextTags", () => {
  it("collects unique tags case-insensitively", () => {
    expect(
      collectProjectContextTags([
        item({ id: "1", tag: "Payment rules" }),
        item({ id: "2", tag: "payment rules" }),
        item({ id: "3", tag: "Codebase tour" }),
      ]),
    ).toEqual(["Codebase tour", "Payment rules"]);
  });

  it("groups items by tag", () => {
    const groups = groupProjectContextByTag([
      item({ id: "1", tag: "Payment rules", title: "A" }),
      item({ id: "2", tag: "Codebase tour", title: "B" }),
      item({ id: "3", tag: "payment rules", title: "C" }),
    ]);
    expect(groups.map((g) => g.label)).toEqual([
      "Codebase tour",
      "Payment rules",
    ]);
    expect(groups[1].items.map((i) => i.id)).toEqual(["1", "3"]);
  });
});
