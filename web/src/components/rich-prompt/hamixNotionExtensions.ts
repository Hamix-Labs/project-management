import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import { TaskList, TaskItem } from "@tiptap/extension-list";
import { Color, TextStyle } from "@tiptap/extension-text-style";
import { Selection } from "@tiptap/extensions";
import { Typography } from "@tiptap/extension-typography";
import { Highlight } from "@tiptap/extension-highlight";
import { Superscript } from "@tiptap/extension-superscript";
import { Subscript } from "@tiptap/extension-subscript";
import { TextAlign } from "@tiptap/extension-text-align";
import { UniqueID } from "@tiptap/extension-unique-id";
import { Emoji, gitHubEmojis } from "@tiptap/extension-emoji";
import { HorizontalRule } from "@/components/tiptap-node/horizontal-rule-node/horizontal-rule-node-extension";
import { UiState } from "@/components/tiptap-extension/ui-state-extension";
import { TableKit } from "@/components/tiptap-node/table-node/extensions/table-node-extension";
import { TableHandleExtension } from "@/components/tiptap-node/table-node/extensions/table-handle";
import { ListNormalizationExtension } from "@/components/tiptap-extension/list-normalization-extension";
import { Indent } from "@/components/tiptap-extension/indent-extension";
import { TripleClickBlockSelection } from "@/components/tiptap-extension/triple-click-block-selection-extension";

const UNIQUE_ID_TYPES = [
  "table",
  "paragraph",
  "bulletList",
  "orderedList",
  "taskList",
  "heading",
  "blockquote",
  "codeBlock",
] as const;

export function buildHamixNotionExtensions(emptyBlockPlaceholder: string) {
  return [
    StarterKit.configure({
      heading: { levels: [1, 2, 3, 4] },
      horizontalRule: false,
      dropcursor: { width: 2 },
      link: { openOnClick: false },
    }),
    HorizontalRule,
    TextAlign.configure({ types: ["heading", "paragraph"] }),
    Placeholder.configure({
      showOnlyCurrent: true,
      emptyNodeClass: "is-empty with-slash",
      placeholder: ({ node, hasAnchor }) => {
        if (!hasAnchor) return "";
        if (node.type.name !== "paragraph") return "";
        if (node.content.size !== 0) return "";
        return emptyBlockPlaceholder;
      },
    }),
    Emoji.configure({
      emojis: gitHubEmojis.filter((emoji) => !emoji.name.includes("regional")),
      forceFallbackImages: true,
    }),
    TableKit.configure({
      table: { resizable: true, cellMinWidth: 120 },
    }),
    TextStyle,
    Superscript,
    Subscript,
    Indent,
    Color,
    TaskList,
    TaskItem.configure({ nested: true }),
    Highlight.configure({ multicolor: true }),
    Selection,
    TableHandleExtension,
    ListNormalizationExtension,
    TripleClickBlockSelection,
    UniqueID.configure({ types: [...UNIQUE_ID_TYPES] }),
    Typography,
    UiState,
  ];
}
