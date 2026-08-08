import { describe, expect, it } from "vitest";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import type { DraftSavePayload } from "@/types";
import {
  draftPayloadFingerprint,
  normalizeDraftPromptForDirty,
} from "./draftAutosaveSignature";

function basePayload(): DraftSavePayload {
  return {
    id: "draft-1",
    name: "Untitled",
    payload: {
      title: "Hello",
      initial_prompt: "<p>body</p>",
      priority: "medium",
      runner: TASK_TEST_DEFAULTS.runner,
      cursor_model: TASK_TEST_DEFAULTS.cursor_model,
      project_id: "",
      repository_id: "",
      worktree_id: "",
      checklist_items: [],
    },
  };
}

function withPayload(patch: Partial<DraftSavePayload["payload"]>): DraftSavePayload {
  const base = basePayload();
  return { ...base, payload: { ...base.payload, ...patch } };
}

describe("normalizeDraftPromptForDirty", () => {
  it.each([
    "",
    "<p></p>",
    "<P></P>",
    "<p><br></p>",
    "<p><br/></p>",
    "<p>&nbsp;</p>",
    "<p>&#160;&#160;</p>",
    "  \n\u200B\uFEFF  ",
  ])("treats editor-empty markup %j as empty", (markup) => {
    expect(normalizeDraftPromptForDirty(markup)).toBe("");
  });

  it("preserves prompts that have visible content", () => {
    expect(normalizeDraftPromptForDirty("<p>Hello world</p>")).toBe(
      "<p>Hello world</p>",
    );
  });
});

describe("draftPayloadFingerprint", () => {
  it("returns identical strings for payloads that only differ in editor whitespace", () => {
    const a = draftPayloadFingerprint(withPayload({ initial_prompt: "<p></p>" }));
    const b = draftPayloadFingerprint(withPayload({ initial_prompt: "<p><br></p>" }));
    expect(a).toBe(b);
  });

  it("changes when title flips", () => {
    const a = draftPayloadFingerprint(basePayload());
    const b = draftPayloadFingerprint(withPayload({ title: "Renamed" }));
    expect(a).not.toBe(b);
  });

  it("changes when the draft name flips", () => {
    const a = draftPayloadFingerprint(basePayload());
    const b = draftPayloadFingerprint({ ...basePayload(), name: "Renamed draft" });
    expect(a).not.toBe(b);
  });

  it("changes when checklist items reorder", () => {
    const a = draftPayloadFingerprint(
      withPayload({ checklist_items: [{ text: "one" }, { text: "two" }] }),
    );
    const b = draftPayloadFingerprint(
      withPayload({ checklist_items: [{ text: "two" }, { text: "one" }] }),
    );
    expect(a).not.toBe(b);
  });

  it("changes when verify commands are added to a criterion", () => {
    const a = draftPayloadFingerprint(
      withPayload({ checklist_items: [{ text: "Ship with tests" }] }),
    );
    const b = draftPayloadFingerprint(
      withPayload({
        checklist_items: [
          {
            text: "Ship with tests",
            verify_commands: [{ command: "go test ./...", expected_outcome: "pass" }],
          },
        ],
      }),
    );
    expect(a).not.toBe(b);
  });

  it("reproduces the payload it was given, with only the prompt normalized", () => {
    const payload = withPayload({ initial_prompt: "<p>Body</p>" });
    expect(JSON.parse(draftPayloadFingerprint(payload))).toEqual(payload);
  });

  it.each([
    ["project", { project_id: "project-1" }, { project_id: "project-2" }],
    ["repository", { repository_id: "repo-1" }, { repository_id: "repo-2" }],
    ["worktree", { worktree_id: "wt-1" }, { worktree_id: "wt-2" }],
  ])("changes when the bound %s changes", (_name, before, after) => {
    expect(draftPayloadFingerprint(withPayload(before))).not.toBe(
      draftPayloadFingerprint(withPayload(after)),
    );
  });
});
