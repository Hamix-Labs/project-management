import { describe, expect, it } from "vitest";
import {
  parseTaskTemplateDetail,
  parseTaskTemplateSummaryList,
} from "./parseTaskApiTemplates";

const baseSummary = {
  id: "tpl-1",
  name: "Bug fix template",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-02T00:00:00Z",
};

describe("parseTaskTemplateSummaryList", () => {
  it("parses enriched summary fields", () => {
    expect(
      parseTaskTemplateSummaryList({
        templates: [
          {
            ...baseSummary,
            primary_tag: "Refactor",
            project_id: "proj-1",
            repository_id: "repo-1",
            instantiate_count: 3,
          },
        ],
      }),
    ).toEqual([
      {
        ...baseSummary,
        primary_tag: "Refactor",
        project_id: "proj-1",
        repository_id: "repo-1",
        instantiate_count: 3,
      },
    ]);
  });

  it("defaults instantiate_count to 0 when omitted", () => {
    expect(
      parseTaskTemplateSummaryList({
        templates: [baseSummary],
      }),
    ).toEqual([
      {
        ...baseSummary,
        instantiate_count: 0,
      },
    ]);
  });

  it("omits primary_tag when absent or empty", () => {
    expect(
      parseTaskTemplateSummaryList({
        templates: [
          { ...baseSummary, instantiate_count: 0 },
          {
            ...baseSummary,
            id: "tpl-2",
            primary_tag: "",
            instantiate_count: 1,
          },
          {
            ...baseSummary,
            id: "tpl-3",
            primary_tag: "   ",
            instantiate_count: 2,
          },
        ],
      }),
    ).toEqual([
      { ...baseSummary, instantiate_count: 0 },
      { ...baseSummary, id: "tpl-2", instantiate_count: 1 },
      { ...baseSummary, id: "tpl-3", instantiate_count: 2 },
    ]);
  });

  it("rejects negative instantiate_count", () => {
    expect(() =>
      parseTaskTemplateSummaryList({
        templates: [{ ...baseSummary, instantiate_count: -1 }],
      }),
    ).toThrow(/non-negative integer/);
  });

  it("rejects non-numeric instantiate_count", () => {
    expect(() =>
      parseTaskTemplateSummaryList({
        templates: [{ ...baseSummary, instantiate_count: "bad" }],
      }),
    ).toThrow(/must be a number/);
  });
});

describe("parseTaskTemplateDetail", () => {
  it("defaults instantiate_count on detail responses", () => {
    expect(
      parseTaskTemplateDetail({
        ...baseSummary,
        payload: {
          title: "Fix bug",
          initial_prompt: "",
          status: "ready",
          priority: "medium",
          checklist_items: [],
          depends_on: [],
        },
      }),
    ).toMatchObject({
      ...baseSummary,
      instantiate_count: 0,
    });
  });
});
