/** @vitest-environment jsdom */
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  createPromptDocumentAdapter,
  readEphemeralPrompt,
  writeEphemeralPrompt,
} from "./promptDocumentAdapter";

vi.mock("@/api", () => ({
  getTask: vi.fn(),
  getTaskDraft: vi.fn(),
  getTaskTemplate: vi.fn(),
  patchTask: vi.fn(),
  patchTaskTemplate: vi.fn(),
  saveTaskDraft: vi.fn(),
}));

import {
  getTaskDraft,
  getTaskTemplate,
  patchTask,
  patchTaskTemplate,
  saveTaskDraft,
} from "@/api";

const mockedGetDraft = vi.mocked(getTaskDraft);
const mockedSaveDraft = vi.mocked(saveTaskDraft);
const mockedPatchTask = vi.mocked(patchTask);
const mockedGetTemplate = vi.mocked(getTaskTemplate);
const mockedPatchTemplate = vi.mocked(patchTaskTemplate);

describe("createPromptDocumentAdapter.saveName", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
  });

  it("patches task title", async () => {
    mockedPatchTask.mockResolvedValue({} as never);
    const adapter = createPromptDocumentAdapter({ kind: "task", id: "t1" });
    await adapter.saveName("  New title  ");
    expect(mockedPatchTask).toHaveBeenCalledWith("t1", { title: "New title" });
  });

  it("saves draft name and payload.title together", async () => {
    mockedGetDraft.mockResolvedValue({
      id: "d1",
      name: "Old",
      created_at: "",
      updated_at: "",
      payload: {
        title: "Old",
        initial_prompt: "<p>body</p>",
        priority: "medium",
        checklist_items: [],
      },
    });
    mockedSaveDraft.mockResolvedValue({ id: "d1", name: "Renamed" });
    const adapter = createPromptDocumentAdapter({ kind: "draft", id: "d1" });
    await adapter.saveName("  Renamed  ");
    expect(mockedSaveDraft).toHaveBeenCalledWith({
      id: "d1",
      name: "Renamed",
      payload: {
        title: "Renamed",
        initial_prompt: "<p>body</p>",
        priority: "medium",
        checklist_items: [],
      },
    });
  });

  it("patches existing template name", async () => {
    mockedGetTemplate.mockResolvedValue({
      id: "tpl1",
      name: "Old tpl",
      created_at: "",
      updated_at: "",
      payload: {
        title: "x",
        initial_prompt: "<p>t</p>",
        priority: "low",
        checklist_items: [],
      },
    } as never);
    mockedPatchTemplate.mockResolvedValue({} as never);
    const adapter = createPromptDocumentAdapter({
      kind: "template",
      id: "tpl1",
    });
    await adapter.saveName("  Fresh  ");
    expect(mockedPatchTemplate).toHaveBeenCalledWith("tpl1", {
      name: "Fresh",
      payload: {
        title: "x",
        initial_prompt: "<p>t</p>",
        priority: "low",
        checklist_items: [],
      },
    });
  });

  it("writes ephemeral name for new template", async () => {
    writeEphemeralPrompt("template-new", {
      html: "<p>seed</p>",
      name: "New template",
    });
    const adapter = createPromptDocumentAdapter(
      { kind: "template", id: "new" },
      { worktreeId: "wt-1" },
    );
    await adapter.saveName("  Draft tpl  ");
    expect(readEphemeralPrompt("template-new")).toEqual({
      html: "<p>seed</p>",
      worktreeId: "wt-1",
      name: "Draft tpl",
    });
  });

  it("updates ephemeral session name only", async () => {
    writeEphemeralPrompt("e1", {
      html: "<p>polish</p>",
      name: "Old task",
      worktreeId: "wt",
    });
    const adapter = createPromptDocumentAdapter({
      kind: "ephemeral",
      id: "e1",
    });
    await adapter.saveName("  Display  ");
    expect(readEphemeralPrompt("e1")).toEqual({
      html: "<p>polish</p>",
      worktreeId: "wt",
      name: "Display",
    });
    expect(mockedPatchTask).not.toHaveBeenCalled();
  });
});
