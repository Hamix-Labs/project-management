import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import { RepoFileMention } from "./extensions/repoFileMention";
import {
  RepoFileSuggestion,
  type RepoFileSuggestionOptions,
} from "./extensions/repoFileSuggestion";
import { PressSpaceForAI } from "./extensions/pressSpaceForAI";
import { SlashMenu, type SlashMenuCatalog } from "./extensions/slashMenu";

export const EMPTY_BLOCK_PLACEHOLDER = "Press Space for AI or / for commands";

export type BuildRichPromptExtensionsOpts = {
  onAiTrigger?: (msg: string) => void;
  slashMenu?: SlashMenuCatalog;
};

export function buildRichPromptExtensions(
  placeholderOverride: string | undefined,
  repoOpts: RepoFileSuggestionOptions,
  opts: BuildRichPromptExtensionsOpts = {},
) {
  const emptyBlockPlaceholder = placeholderOverride?.trim()
    ? placeholderOverride
    : EMPTY_BLOCK_PLACEHOLDER;
  return [
    StarterKit.configure({
      heading: { levels: [2, 3, 4] },
    }),
    Placeholder.configure({
      showOnlyCurrent: true,
      placeholder: ({ node, hasAnchor }) => {
        if (!hasAnchor) return "";
        if (node.type.name !== "paragraph") return "";
        if (node.content.size !== 0) return "";
        return emptyBlockPlaceholder;
      },
    }),
    RepoFileMention,
    RepoFileSuggestion.configure(repoOpts),
    PressSpaceForAI.configure({
      onAiTrigger: opts.onAiTrigger,
    }),
    SlashMenu.configure({
      onAiTrigger: opts.onAiTrigger,
      catalog: opts.slashMenu ?? "all",
    }),
  ];
}
