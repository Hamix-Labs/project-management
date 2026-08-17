import type { SuggestionItem } from "@/components/tiptap-ui-utils/suggestion-menu";
import type {
  SlashMenuConfig,
  SlashMenuItemType,
} from "@/components/tiptap-ui/slash-dropdown-menu/use-slash-dropdown-menu";
import { AiSparklesIcon } from "@/components/tiptap-icons/ai-sparkles-icon";
import { AtSignIcon } from "@/components/tiptap-icons/at-sign-icon";

/** Template slash rows we keep. Image, user mention, TOC, and Tiptap Cloud AI stay off. */
export const HAMIX_SLASH_ENABLED_ITEMS: SlashMenuItemType[] = [
  "text",
  "heading_1",
  "heading_2",
  "heading_3",
  "bullet_list",
  "ordered_list",
  "task_list",
  "quote",
  "code_block",
  "emoji",
  "table",
  "divider",
];

export function hamixSlashCustomItems(
  onAiTrigger?: (msg: string) => void,
): SuggestionItem[] {
  return [
    {
      title: "Ask AI",
      subtext: "Open the inline composer",
      keywords: ["ai", "ask", "assistant", "space"],
      badge: AiSparklesIcon,
      group: "Hamix",
      onSelect: () => {
        onAiTrigger?.("");
      },
    },
    {
      title: "Mention a file",
      subtext: "Insert @ to search the worktree",
      keywords: ["mention", "file", "at", "@"],
      badge: AtSignIcon,
      group: "Hamix",
      onSelect: ({ editor }) => {
        editor.chain().focus().insertContent("@").run();
      },
    },
  ];
}

export function buildHamixSlashMenuConfig(
  onAiTrigger?: (msg: string) => void,
): SlashMenuConfig {
  return {
    enabledItems: HAMIX_SLASH_ENABLED_ITEMS,
    customItems: hamixSlashCustomItems(onAiTrigger),
  };
}
