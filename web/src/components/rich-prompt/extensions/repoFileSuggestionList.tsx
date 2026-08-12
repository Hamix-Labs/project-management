import { useEffect, useMemo, useRef, useState } from "react";
import {
  EnterGlyph,
  FileGlyph,
  iconKindFor,
  SearchGlyph,
} from "./repoFileSuggestionGlyphs";

export type RepoSuggestionItem = { path: string };

const ROW_HEIGHT_PX = 36;
const LIST_MAX_HEIGHT_PX = 280;
const OVERSCAN = 6;

type ListProps = {
  items: RepoSuggestionItem[];
  command: (item: RepoSuggestionItem) => void;
  query?: string;
  selectedIndex?: number;
  indexing?: boolean;
};

export function splitPath(path: string): [string, string] {
  const normalized = path.replace(/\\/g, "/");
  const idx = normalized.lastIndexOf("/");
  if (idx === -1) return ["", normalized];
  return [normalized.slice(0, idx + 1), normalized.slice(idx + 1)];
}

export function RepoFileSuggestionList({
  items,
  command,
  query = "",
  selectedIndex = -1,
  indexing = false,
}: ListProps) {
  const listRef = useRef<HTMLDivElement>(null);
  const [scrollTop, setScrollTop] = useState(0);
  const trimmedQuery = query.trim();

  const viewportHeight = Math.min(
    LIST_MAX_HEIGHT_PX,
    Math.max(ROW_HEIGHT_PX, items.length * ROW_HEIGHT_PX || ROW_HEIGHT_PX),
  );

  const { start, end } = useMemo(() => {
    const startIdx = Math.max(
      0,
      Math.floor(scrollTop / ROW_HEIGHT_PX) - OVERSCAN,
    );
    const visible = Math.ceil(viewportHeight / ROW_HEIGHT_PX) + OVERSCAN * 2;
    return {
      start: startIdx,
      end: Math.min(items.length, startIdx + visible),
    };
  }, [scrollTop, viewportHeight, items.length]);

  useEffect(() => {
    const el = listRef.current;
    if (!el || items.length === 0 || selectedIndex < 0) return;
    const rowTop = selectedIndex * ROW_HEIGHT_PX;
    const rowBottom = rowTop + ROW_HEIGHT_PX;
    if (rowTop < el.scrollTop) {
      el.scrollTop = rowTop;
    } else if (rowBottom > el.scrollTop + el.clientHeight) {
      el.scrollTop = rowBottom - el.clientHeight;
    }
  }, [selectedIndex, items.length]);

  const slice = items.slice(start, end);

  let emptyMessage: string;
  if (indexing && items.length === 0) {
    emptyMessage = "Indexing repository files…";
  } else if (trimmedQuery) {
    emptyMessage = `No files match “${trimmedQuery}”`;
  } else {
    emptyMessage = "No matching files";
  }

  return (
    <div className="mention-dropdown mention-dropdown--repo-files tiptap-suggestion-list">
      <div className="mention-dropdown__search" aria-hidden="true">
        <SearchGlyph />
        <span
          className={
            trimmedQuery
              ? "mention-dropdown__search-query"
              : "mention-dropdown__search-placeholder"
          }
        >
          {trimmedQuery || "Search repository files..."}
        </span>
        <kbd className="mention-dropdown__at-kbd">@</kbd>
      </div>

      <div
        ref={listRef}
        role="listbox"
        aria-label="Matching repository files"
        className="mention-dropdown__list mention-dropdown__list--virtual"
        style={{ height: viewportHeight, overflowY: "auto" }}
        onScroll={(e) => setScrollTop(e.currentTarget.scrollTop)}
      >
        {items.length === 0 ? (
          <div className="mention-dropdown__empty" role="presentation">
            {emptyMessage}
          </div>
        ) : (
          <ul
            className="mention-dropdown__virtual-window"
            style={{
              height: items.length * ROW_HEIGHT_PX,
              position: "relative",
              margin: 0,
              padding: 0,
              listStyle: "none",
            }}
          >
            {slice.map((item, offset) => {
              const i = start + offset;
              const [dir, name] = splitPath(item.path);
              const active = i === selectedIndex;
              return (
                <li
                  key={item.path}
                  className={[
                    "mention-option",
                    "mention-option--repo-file",
                    active ? "mention-option--active" : "",
                  ]
                    .filter(Boolean)
                    .join(" ")}
                  role="option"
                  aria-selected={active}
                  data-index={i}
                  style={{
                    position: "absolute",
                    top: i * ROW_HEIGHT_PX,
                    left: 0,
                    right: 0,
                    height: ROW_HEIGHT_PX,
                  }}
                >
                  <button type="button" onClick={() => command(item)}>
                    <FileGlyph kind={iconKindFor(item.path)} />
                    <span className="mention-dropdown__path">
                      {dir ? (
                        <span className="mention-dropdown__dir">{dir}</span>
                      ) : null}
                      <span className="mention-dropdown__name">{name}</span>
                    </span>
                  </button>
                </li>
              );
            })}
          </ul>
        )}
      </div>

      <div className="mention-dropdown__footer" aria-hidden="true">
        {indexing ? (
          <span>Indexing files…</span>
        ) : (
          <span>
            <span className="mention-dropdown__footer-keys">↑↓</span> to navigate
          </span>
        )}
        <span className="mention-dropdown__footer-select">
          <EnterGlyph /> to select
        </span>
      </div>
    </div>
  );
}
