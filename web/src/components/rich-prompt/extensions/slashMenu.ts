import { Extension } from "@tiptap/core";
import Suggestion, { exitSuggestion } from "@tiptap/suggestion";
import { PluginKey } from "@tiptap/pm/state";
import { ReactRenderer } from "@tiptap/react";
import type {
  SuggestionKeyDownProps,
  SuggestionProps,
} from "@tiptap/suggestion";
import type { Instance as TippyInstance } from "tippy.js";
import tippy from "tippy.js";
import {
  filterSlashItems,
  SlashMenuList,
  type SlashMenuItem,
} from "./slashMenuList";
import {
  MENTION_POPOVER_Z_INDEX,
  referenceRectForSuggestion,
} from "./repoFileSuggestionReferenceRect";

export const slashMenuPluginKey = new PluginKey("slashMenu");

export type SlashMenuOptions = {
  onAiTrigger?: (msg: string) => void;
};

export type SlashCommandId =
  | "heading-2"
  | "heading-3"
  | "bullet-list"
  | "ordered-list"
  | "blockquote"
  | "mention"
  | "ai";

export const SLASH_ITEMS: SlashMenuItem[] = [
  {
    id: "heading-2",
    label: "Heading 2",
    hint: "Section title",
    keywords: ["h2", "heading", "title"],
  },
  {
    id: "heading-3",
    label: "Heading 3",
    hint: "Subsection",
    keywords: ["h3", "heading", "subtitle"],
  },
  {
    id: "bullet-list",
    label: "Bulleted list",
    hint: "• Unordered list",
    keywords: ["list", "ul", "bullet", "unordered"],
  },
  {
    id: "ordered-list",
    label: "Numbered list",
    hint: "1. Ordered list",
    keywords: ["list", "ol", "numbered", "ordered"],
  },
  {
    id: "blockquote",
    label: "Quote",
    hint: "Block quote",
    keywords: ["quote", "blockquote", "callout"],
  },
  {
    id: "mention",
    label: "Mention a file",
    hint: "Insert @",
    keywords: ["mention", "file", "at", "@"],
  },
  {
    id: "ai",
    label: "Ask AI",
    hint: "Open the inline composer",
    keywords: ["ai", "assistant", "space", "help"],
  },
];

export function runSlashCommand(
  editor: import("@tiptap/core").Editor,
  range: import("@tiptap/core").Range,
  itemId: SlashCommandId,
  onAiTrigger?: (msg: string) => void,
): void {
  const chain = editor.chain().focus().deleteRange(range);
  switch (itemId) {
    case "heading-2":
      chain.setNode("heading", { level: 2 }).run();
      return;
    case "heading-3":
      chain.setNode("heading", { level: 3 }).run();
      return;
    case "bullet-list":
      chain.toggleBulletList().run();
      return;
    case "ordered-list":
      chain.toggleOrderedList().run();
      return;
    case "blockquote":
      chain.setBlockquote().run();
      return;
    case "mention":
      chain.insertContent("@").run();
      return;
    case "ai":
      chain.run();
      onAiTrigger?.("");
      return;
  }
}

export const SlashMenu = Extension.create<SlashMenuOptions>({
  name: "slashMenu",

  addOptions() {
    return {
      onAiTrigger: undefined,
    };
  },

  addProseMirrorPlugins() {
    const onAiTrigger = this.options.onAiTrigger;

    return [
      Suggestion<SlashMenuItem, SlashMenuItem>({
        pluginKey: slashMenuPluginKey,
        editor: this.editor,
        char: "/",
        allowSpaces: false,
        startOfLine: true,
        allow: ({ state, range }) => {
          // Only open when `/` sits at the very start of an otherwise-empty
          // block. `startOfLine` already guarantees the trigger position; we
          // additionally require the block to contain nothing after the match.
          const $from = state.doc.resolve(range.from);
          if ($from.parentOffset !== 0) return false;
          const parent = $from.parent;
          if (!parent.isTextblock) return false;
          const afterMatch = range.to - $from.start();
          return parent.content.size === afterMatch;
        },
        command: ({ editor, range, props }) => {
          runSlashCommand(
            editor,
            range,
            props.id as SlashCommandId,
            onAiTrigger,
          );
          exitSuggestion(editor.view, slashMenuPluginKey);
        },
        items: ({ query }) => filterSlashItems(SLASH_ITEMS, query),
        render: () => createSlashMenuRender(),
      }),
    ];
  },
});

function createSlashMenuRender() {
  let component: ReactRenderer | null = null;
  let popup: TippyInstance | null = null;
  let latestProps: SuggestionProps<SlashMenuItem, SlashMenuItem> | null = null;
  let selectedIndex = 0;

  const listProps = (
    props: SuggestionProps<SlashMenuItem, SlashMenuItem>,
    index: number,
  ) => ({
    items: props.items,
    query: props.query,
    selectedIndex: props.items.length === 0 ? -1 : index,
    command: (item: SlashMenuItem) => {
      props.command(item);
    },
  });

  return {
    onStart(props: SuggestionProps<SlashMenuItem, SlashMenuItem>) {
      latestProps = props;
      selectedIndex = 0;
      component = new ReactRenderer(SlashMenuList, {
        props: listProps(props, selectedIndex),
        editor: props.editor,
      });

      const t = tippy(document.body, {
        getReferenceClientRect: () =>
          latestProps != null
            ? referenceRectForSuggestion(latestProps)
            : new DOMRect(0, 0, 0, 0),
        appendTo: () => document.body,
        content: component.element,
        showOnCreate: true,
        interactive: true,
        trigger: "manual",
        placement: "bottom-start",
        zIndex: MENTION_POPOVER_Z_INDEX,
        theme: "repo-files-popover",
        arrow: false,
        maxWidth: "min(100vw - 2rem, 22rem)",
        offset: [0, 6],
      });
      popup = Array.isArray(t) ? t[0]! : t;
    },

    onUpdate(props: SuggestionProps<SlashMenuItem, SlashMenuItem>) {
      latestProps = props;
      if (props.items.length === 0) {
        selectedIndex = -1;
      } else {
        selectedIndex = Math.min(Math.max(selectedIndex, 0), props.items.length - 1);
      }
      component?.updateProps(listProps(props, selectedIndex));
      popup?.setProps({
        getReferenceClientRect: () =>
          latestProps != null
            ? referenceRectForSuggestion(latestProps)
            : new DOMRect(0, 0, 0, 0),
      });
    },

    onKeyDown(props: SuggestionKeyDownProps) {
      if (props.event.key === "Escape") {
        popup?.hide();
        return true;
      }
      if (!latestProps || latestProps.items.length === 0) {
        return false;
      }
      if (props.event.key === "ArrowDown") {
        props.event.preventDefault();
        selectedIndex = (selectedIndex + 1) % latestProps.items.length;
        component?.updateProps(listProps(latestProps, selectedIndex));
        return true;
      }
      if (props.event.key === "ArrowUp") {
        props.event.preventDefault();
        selectedIndex =
          selectedIndex <= 0 ? latestProps.items.length - 1 : selectedIndex - 1;
        component?.updateProps(listProps(latestProps, selectedIndex));
        return true;
      }
      if (props.event.key === "Enter") {
        if (selectedIndex < 0) return false;
        const item = latestProps.items[selectedIndex];
        if (!item) return false;
        props.event.preventDefault();
        latestProps.command(item);
        return true;
      }
      return false;
    },

    onExit() {
      popup?.destroy();
      component?.destroy();
      popup = null;
      component = null;
    },
  };
}
