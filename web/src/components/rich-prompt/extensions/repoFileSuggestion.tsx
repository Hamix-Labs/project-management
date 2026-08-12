import { Extension } from "@tiptap/core";
import Suggestion, { exitSuggestion } from "@tiptap/suggestion";
import { PluginKey } from "@tiptap/pm/state";
import { RepoFileSuggestionList, type RepoSuggestionItem } from "./repoFileSuggestionList";
import { resolveSuggestionItems } from "./repoFileSuggestionItems";
import { createRepoFileSuggestionRender } from "./repoFileSuggestionRender";

export type { RepoSuggestionItem } from "./repoFileSuggestionList";

export const repoFileSuggestionPluginKey = new PluginKey("repoFileSuggestion");

export type RepoFilePickedPayload = {
  path: string;
  insertAt: number;
};

export type RepoFileSuggestionOptions = {
  onRepoUnavailable: () => void;
  onRepoAvailable: () => void;
  onSuggestFetchChange?: (busy: boolean) => void;
  onFilePicked?: (payload: RepoFilePickedPayload) => void;
  getWorktreeId?: () => string | undefined;
};

export const RepoFileSuggestion = Extension.create<RepoFileSuggestionOptions>({
  name: "repoFileSuggestion",

  addOptions() {
    return {
      onRepoUnavailable: () => {},
      onRepoAvailable: () => {},
      onSuggestFetchChange: undefined as
        | ((busy: boolean) => void)
        | undefined,
      onFilePicked: undefined as
        | ((payload: RepoFilePickedPayload) => void)
        | undefined,
    };
  },

  addProseMirrorPlugins() {
    const onUnavailable = this.options.onRepoUnavailable;
    const onAvailable = this.options.onRepoAvailable;
    const onSuggestFetchChange = this.options.onSuggestFetchChange;
    const onFilePicked = this.options.onFilePicked;
    const getWorktreeId = this.options.getWorktreeId;
    const setFetchBusy = (busy: boolean) => {
      onSuggestFetchChange?.(busy);
    };

    let mentionSearchSeq = 0;
    let lastRepoSuggestionItems: RepoSuggestionItem[] = [];

    return [
      Suggestion<RepoSuggestionItem, RepoSuggestionItem>({
        pluginKey: repoFileSuggestionPluginKey,
        editor: this.editor,
        char: "@",
        allowSpaces: false,
        allowedPrefixes: [" ", "\n"],
        command: ({ editor, range, props }) => {
          const insertAt = range.from;
          const path = props.path.replace(/\\/g, "/");
          editor.chain().focus().deleteRange(range).run();
          exitSuggestion(editor.view, repoFileSuggestionPluginKey);
          onFilePicked?.({ path, insertAt });
        },
        items: async ({ query }) => {
          mentionSearchSeq += 1;
          const seq = mentionSearchSeq;
          try {
            const scopedWorktreeId = getWorktreeId?.()?.trim();
            if (!scopedWorktreeId) {
              lastRepoSuggestionItems = [];
              return [];
            }
            const resolved = await resolveSuggestionItems(
              scopedWorktreeId,
              query,
            );
            if (seq !== mentionSearchSeq) return lastRepoSuggestionItems;
            if (resolved.unavailable) onUnavailable();
            if (resolved.available) onAvailable();
            lastRepoSuggestionItems = resolved.items;
            return lastRepoSuggestionItems;
          } catch {
            if (seq !== mentionSearchSeq) return lastRepoSuggestionItems;
            return lastRepoSuggestionItems;
          } finally {
            if (seq === mentionSearchSeq) {
              setFetchBusy(false);
            }
          }
        },
        render: () =>
          createRepoFileSuggestionRender({
            getWorktreeId,
            onUnavailable,
            onAvailable,
            setFetchBusy,
            setLastItems: (items) => {
              lastRepoSuggestionItems = items;
            },
          }),
      }),
    ];
  },
});

// Keep the list component import used for type-side coupling in tests/docs.
void RepoFileSuggestionList;
