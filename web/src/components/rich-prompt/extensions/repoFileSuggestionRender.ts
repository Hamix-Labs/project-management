import { ReactRenderer } from "@tiptap/react";
import type {
  SuggestionKeyDownProps,
  SuggestionProps,
} from "@tiptap/suggestion";
import type { Instance as TippyInstance } from "tippy.js";
import tippy from "tippy.js";
import {
  getRepoFileIndexSnapshot,
  subscribeRepoFileIndex,
  warmRepoFileIndex,
} from "@/lib/repoFileIndex";
import { RepoFileSuggestionList, type RepoSuggestionItem } from "./repoFileSuggestionList";
import { itemsFromIndex } from "./repoFileSuggestionItems";
import {
  MENTION_POPOVER_Z_INDEX,
  referenceRectForSuggestion,
} from "./repoFileSuggestionReferenceRect";

type RenderDeps = {
  getWorktreeId?: () => string | undefined;
  onUnavailable: () => void;
  onAvailable: () => void;
  setFetchBusy: (busy: boolean) => void;
  setLastItems: (items: RepoSuggestionItem[]) => void;
};

/** TipTap Suggestion render() lifecycle for the repo file popover. */
export function createRepoFileSuggestionRender(deps: RenderDeps) {
  const { getWorktreeId, onUnavailable, onAvailable, setFetchBusy, setLastItems } =
    deps;

  let component: ReactRenderer | null = null;
  let popup: TippyInstance | null = null;
  let latestProps: SuggestionProps<
    RepoSuggestionItem,
    RepoSuggestionItem
  > | null = null;
  let selectedIndex = 0;
  let lastQuery = "";
  let unsubIndex: (() => void) | null = null;
  let indexing = false;

  const listProps = (
    props: SuggestionProps<RepoSuggestionItem, RepoSuggestionItem>,
    index: number,
    indexBusy: boolean,
  ) => ({
    items: props.items,
    query: props.query,
    selectedIndex: index,
    indexing: indexBusy,
    command: (item: RepoSuggestionItem) => {
      props.command(item);
    },
  });

  const refreshFromIndex = () => {
    const wt = getWorktreeId?.()?.trim();
    if (!wt || !latestProps) return;
    const snap = getRepoFileIndexSnapshot(wt);
    if (snap.status === "error") {
      onUnavailable();
      latestProps = { ...latestProps, items: [] };
      setLastItems([]);
      selectedIndex = 0;
      indexing = false;
      component?.updateProps(listProps(latestProps, selectedIndex, false));
      return;
    }
    const next = itemsFromIndex(wt, latestProps.query);
    indexing = next.indexing;
    latestProps = { ...latestProps, items: next.items };
    setLastItems(next.items);
    const max = Math.max(0, next.items.length - 1);
    selectedIndex = Math.min(selectedIndex, max);
    if (next.items.length === 0) selectedIndex = 0;
    component?.updateProps(listProps(latestProps, selectedIndex, indexing));
    if (next.items.length > 0 || snap.status === "ready") {
      onAvailable();
    }
  };

  return {
    onBeforeStart() {
      setFetchBusy(true);
    },
    onBeforeUpdate() {
      setFetchBusy(true);
    },
    onStart(props: SuggestionProps<RepoSuggestionItem, RepoSuggestionItem>) {
      latestProps = props;
      selectedIndex = 0;
      lastQuery = props.query;
      const wt = getWorktreeId?.()?.trim();
      if (wt) {
        warmRepoFileIndex(wt);
        indexing = itemsFromIndex(wt, props.query).indexing;
        unsubIndex = subscribeRepoFileIndex(wt, refreshFromIndex);
      }
      component = new ReactRenderer(RepoFileSuggestionList, {
        props: listProps(props, selectedIndex, indexing),
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
        maxWidth: "min(100vw - 2rem, 28rem)",
        offset: [0, 6],
      });
      popup = Array.isArray(t) ? t[0]! : t;
    },

    onUpdate(props: SuggestionProps<RepoSuggestionItem, RepoSuggestionItem>) {
      latestProps = props;
      if (props.query !== lastQuery) {
        selectedIndex = 0;
        lastQuery = props.query;
      } else {
        const max = Math.max(0, props.items.length - 1);
        selectedIndex = Math.min(selectedIndex, max);
      }
      if (props.items.length === 0) selectedIndex = 0;
      const wt = getWorktreeId?.()?.trim();
      indexing = wt ? itemsFromIndex(wt, props.query).indexing : false;
      component?.updateProps(listProps(props, selectedIndex, indexing));
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
        selectedIndex = Math.min(
          selectedIndex + 1,
          latestProps.items.length - 1,
        );
        component?.updateProps(
          listProps(latestProps, selectedIndex, indexing),
        );
        return true;
      }
      if (props.event.key === "ArrowUp") {
        props.event.preventDefault();
        selectedIndex = Math.max(selectedIndex - 1, 0);
        component?.updateProps(
          listProps(latestProps, selectedIndex, indexing),
        );
        return true;
      }
      if (props.event.key === "Enter") {
        const item = latestProps.items[selectedIndex];
        if (!item) return false;
        props.event.preventDefault();
        latestProps.command(item);
        return true;
      }
      return false;
    },

    onExit() {
      unsubIndex?.();
      unsubIndex = null;
      setLastItems([]);
      selectedIndex = 0;
      setFetchBusy(false);
      popup?.destroy();
      component?.destroy();
    },
  };
}
