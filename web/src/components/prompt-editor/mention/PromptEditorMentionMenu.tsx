import { useEffect, useRef } from "react";
import type { SuggestionMenuProps } from "@blocknote/react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { splitRepoPath } from "../repoFileRef";

export type PromptFileMentionItem = {
  title: string;
  onItemClick: () => void;
  query: string;
};

/** Row height in px; must match `.prompt-editor-mention-menu__item`. */
const itemHeight = 33;
/** How many rows the scroller shows before it starts scrolling. */
const visibleItems = 9;
/** Must match `.prompt-editor-mention-menu` width. */
const menuWidth = 340;

function FileIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      aria-hidden="true"
    >
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
      <path d="M14 2v6h6" />
    </svg>
  );
}

export function PromptEditorMentionMenu(
  props: SuggestionMenuProps<PromptFileMentionItem>,
) {
  const { items, selectedIndex, onItemClick, loadingState } = props;
  const query = items[0]?.query ?? "";
  const loading =
    loadingState === "loading" || loadingState === "loading-initial";

  const scrollRef = useRef<HTMLDivElement>(null);
  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => scrollRef.current,
    estimateSize: () => itemHeight,
    overscan: 8,
    // The menu is mounted and measured in the same frame it becomes visible;
    // without a starting rect the first paint would be an empty popup.
    initialRect: { width: menuWidth, height: itemHeight * visibleItems },
  });

  // Arrow keys move BlockNote's selection, not the scroll position, so the
  // selected row has to be pulled into view for keyboard navigation to work
  // past the first screenful.
  useEffect(() => {
    if (selectedIndex == null) return;
    virtualizer.scrollToIndex(selectedIndex, { align: "auto" });
  }, [selectedIndex, virtualizer]);

  const virtualItems = virtualizer.getVirtualItems();

  return (
    <div className="prompt-editor-mention-menu">
      <div className="prompt-editor-mention-menu__search">
        {loading ? (
          "Searching files…"
        ) : (
          <>
            Searching files for <b>{query || "…"}</b>
            {items.length > 0 ? (
              <span className="prompt-editor-mention-menu__count">
                {items.length.toLocaleString("en-US")}
              </span>
            ) : null}
          </>
        )}
      </div>
      {items.length === 0 && !loading ? (
        <div className="prompt-editor-mention-menu__empty">
          No files match that search.
        </div>
      ) : null}
      {items.length > 0 ? (
        <div
          ref={scrollRef}
          className="prompt-editor-mention-menu__list"
          style={{ maxHeight: itemHeight * visibleItems }}
          role="listbox"
          aria-label="Repository files"
        >
          <div
            className="prompt-editor-mention-menu__viewport"
            style={{ height: virtualizer.getTotalSize() }}
          >
            {virtualItems.map((virtualItem) => {
              const item = items[virtualItem.index];
              if (!item) return null;
              const { fileName, dirPath } = splitRepoPath(item.title);
              const isSelected = virtualItem.index === selectedIndex;
              return (
                <button
                  key={item.title}
                  type="button"
                  role="option"
                  aria-selected={isSelected}
                  aria-posinset={virtualItem.index + 1}
                  aria-setsize={items.length}
                  data-selected={isSelected ? "true" : "false"}
                  className="prompt-editor-mention-menu__item"
                  style={{ transform: `translateY(${virtualItem.start}px)` }}
                  onClick={() => onItemClick?.(item)}
                >
                  <FileIcon />
                  <span className="prompt-editor-mention-menu__fname">
                    {fileName}
                  </span>
                  {dirPath ? (
                    <span className="prompt-editor-mention-menu__fpath">
                      {dirPath}
                    </span>
                  ) : null}
                </button>
              );
            })}
          </div>
        </div>
      ) : null}
      <div className="prompt-editor-mention-menu__footer">
        <span>
          <span className="prompt-editor-kbd">↑↓</span> navigate
        </span>
        <span>
          <span className="prompt-editor-kbd">↵</span> embed file
        </span>
        <span>
          <span className="prompt-editor-kbd">esc</span> dismiss
        </span>
      </div>
    </div>
  );
}
