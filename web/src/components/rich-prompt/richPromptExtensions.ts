import { RepoFileMention } from "./extensions/repoFileMention";
import {
  RepoFileSuggestion,
  type RepoFileSuggestionOptions,
} from "./extensions/repoFileSuggestion";
import { PressSpaceForAI } from "./extensions/pressSpaceForAI";
import { buildHamixNotionExtensions } from "./hamixNotionExtensions";

export const EMPTY_BLOCK_PLACEHOLDER = "Press Space for AI or / for commands";

export type BuildRichPromptExtensionsOpts = {
  onAiTrigger?: (msg: string) => void;
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
    ...buildHamixNotionExtensions(emptyBlockPlaceholder),
    RepoFileMention,
    RepoFileSuggestion.configure(repoOpts),
    PressSpaceForAI.configure({
      onAiTrigger: opts.onAiTrigger,
    }),
  ];
}
