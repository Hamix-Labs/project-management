// @vitest-environment jsdom
import { Editor } from "@tiptap/core";
import Placeholder from "@tiptap/extension-placeholder";
import StarterKit from "@tiptap/starter-kit";
import { waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { clearRepoFileIndex } from "@/lib/repoFileIndex";
import { RepoFileMention } from "./repoFileMention";
import { RepoFileSuggestion } from "./repoFileSuggestion";
import { GIT_TEST_WORKTREE_ID } from "@/test/handlers/git";

const suggestionOptions = {
  getWorktreeId: () => GIT_TEST_WORKTREE_ID,
};

function filesPageResponse(paths: string[], status = 200) {
  return new Response(
    JSON.stringify({ paths, has_more: false, source: "git" }),
    {
      status,
      headers: { "Content-Type": "application/json" },
    },
  );
}

describe("RepoFileSuggestion", () => {
  beforeEach(() => {
    clearRepoFileIndex();
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(filesPageResponse(["a/b.go"])),
    );
  });

  afterEach(() => {
    clearRepoFileIndex();
    vi.unstubAllGlobals();
  });

  it("invokes onRepoUnavailable when /repo/files returns 503", async () => {
    vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 503 }));
    const onRepoUnavailable = vi.fn();
    const editor = new Editor({
      extensions: [
        StarterKit,
        Placeholder.configure({ placeholder: "" }),
        RepoFileMention,
        RepoFileSuggestion.configure({ onRepoUnavailable, ...suggestionOptions }),
      ],
      content: "<p></p>",
    });
    editor.chain().insertContent("@").run();
    await waitFor(() => expect(onRepoUnavailable).toHaveBeenCalled());
    editor.destroy();
  });

  it("invokes onRepoUnavailable when index warm fetch throws", async () => {
    vi.mocked(fetch).mockRejectedValue(new Error("network"));
    const onRepoUnavailable = vi.fn();
    const editor = new Editor({
      extensions: [
        StarterKit,
        Placeholder.configure({ placeholder: "" }),
        RepoFileMention,
        RepoFileSuggestion.configure({ onRepoUnavailable, ...suggestionOptions }),
      ],
      content: "<p></p>",
    });
    editor.chain().insertContent("@").run();
    await waitFor(() => expect(onRepoUnavailable).toHaveBeenCalled());
    editor.destroy();
  });

  it("invokes onRepoAvailable when file index warm succeeds", async () => {
    const onRepoAvailable = vi.fn();
    const editor = new Editor({
      extensions: [
        StarterKit,
        Placeholder.configure({ placeholder: "" }),
        RepoFileMention,
        RepoFileSuggestion.configure({ onRepoAvailable, ...suggestionOptions }),
      ],
      content: "<p></p>",
    });
    editor.chain().insertContent("@").run();
    await waitFor(() => expect(onRepoAvailable).toHaveBeenCalled());
    editor.destroy();
  });
});
