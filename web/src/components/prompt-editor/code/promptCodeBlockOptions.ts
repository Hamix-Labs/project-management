import type { CodeBlockOptions } from "@blocknote/core";
import { codeBlockOptions } from "@blocknote/code-block";
import { createPromptCodeHighlighter } from "./shiki.bundle";

export type PromptCodeLanguage = {
  id: string;
  name: string;
};

export const promptCodeBlockOptions = {
  ...codeBlockOptions,
  supportedLanguages: {
    ...codeBlockOptions.supportedLanguages,
    go: {
      name: "Go",
      aliases: ["go", "golang"],
    },
  },
  createHighlighter: () => createPromptCodeHighlighter(),
} satisfies CodeBlockOptions;

export function promptCodeLanguages(): PromptCodeLanguage[] {
  return Object.entries(promptCodeBlockOptions.supportedLanguages).map(
    ([id, { name }]) => ({ id, name }),
  );
}
