import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import { RepoFileMention } from "./extensions/repoFileMention";
import {
  RepoFileSuggestion,
  type RepoFileSuggestionOptions,
} from "./extensions/repoFileSuggestion";

export function buildRichPromptExtensions(
  placeholder: string | undefined,
  repoOpts: RepoFileSuggestionOptions,
) {
  return [
    StarterKit.configure({
      heading: { levels: [2, 3, 4] },
    }),
    Placeholder.configure({
      placeholder: placeholder ?? "",
    }),
    RepoFileMention,
    RepoFileSuggestion.configure(repoOpts),
  ];
}
