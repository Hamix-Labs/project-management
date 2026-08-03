import { BlockNoteEditor } from "@blocknote/core";
import {
  looksLikeStoredHtml,
  plainTextToInitialHtml,
  previewTextFromPrompt,
  promptHasVisibleContent,
} from "@/lib/promptFormat";
import {
  promptEditorSchema,
  type PromptEditorSchema,
} from "./blockNoteSchema";

type Editor = BlockNoteEditor<
  PromptEditorSchema["blockSchema"],
  PromptEditorSchema["inlineContentSchema"],
  PromptEditorSchema["styleSchema"]
>;

export type HtmlToInitialBlocksResult = {
  blocks: ReturnType<Editor["tryParseHTMLToBlocks"]>;
  usedFallback: boolean;
};

function htmlForEditor(value: string): string {
  if (!value.trim()) return "<p></p>";
  return looksLikeStoredHtml(value) ? value : plainTextToInitialHtml(value);
}

function blockTreeHasText(blocks: HtmlToInitialBlocksResult["blocks"]): boolean {
  const walk = (nodes: HtmlToInitialBlocksResult["blocks"]): boolean => {
    for (const block of nodes) {
      const content: unknown = block.content;
      if (typeof content === "string" && content.trim()) return true;
      if (Array.isArray(content)) {
        for (const inline of content as unknown[]) {
          if (typeof inline === "string" && inline.trim()) return true;
          if (
            inline &&
            typeof inline === "object" &&
            "text" in inline &&
            typeof (inline as { text?: unknown }).text === "string" &&
            ((inline as { text: string }).text).trim()
          ) {
            return true;
          }
        }
      }
      if (block.children?.length && walk(block.children as typeof blocks)) {
        return true;
      }
    }
    return false;
  };
  return walk(blocks);
}

function createParseEditor(): Editor {
  return BlockNoteEditor.create({ schema: promptEditorSchema }) as Editor;
}

/**
 * Parse stored prompt HTML into BlockNote initialContent.
 * Never returns an empty document for non-empty input — falls back to plain
 * paragraphs and sets usedFallback so the UI can warn.
 */
export function htmlToInitialBlocks(html: string): HtmlToInitialBlocksResult {
  const editor = createParseEditor();
  const prepared = htmlForEditor(html);
  const expectContent = promptHasVisibleContent(html);

  const parse = (source: string) => editor.tryParseHTMLToBlocks(source);

  try {
    const blocks = parse(prepared);
    if (expectContent && !blockTreeHasText(blocks)) {
      const fallbackHtml = plainTextToInitialHtml(previewTextFromPrompt(html));
      return { blocks: parse(fallbackHtml), usedFallback: true };
    }
    if (blocks.length === 0) {
      return { blocks: parse("<p></p>"), usedFallback: expectContent };
    }
    return { blocks, usedFallback: false };
  } catch {
    if (!expectContent) {
      return { blocks: parse("<p></p>"), usedFallback: false };
    }
    const fallbackHtml = plainTextToInitialHtml(previewTextFromPrompt(html));
    try {
      return { blocks: parse(fallbackHtml), usedFallback: true };
    } catch {
      const text = previewTextFromPrompt(html);
      return {
        blocks: [
          {
            type: "paragraph",
            content: [{ type: "text", text, styles: {} }],
          },
        ] as unknown as HtmlToInitialBlocksResult["blocks"],
        usedFallback: true,
      };
    }
  }
}

/** Normalize a value prop into HTML suitable for BlockNote import. */
export function normalizePromptHtmlForEditor(value: string): string {
  return htmlForEditor(value);
}
