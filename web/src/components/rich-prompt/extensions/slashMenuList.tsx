import { useEffect, useRef } from "react";
import { EnterGlyph } from "./repoFileSuggestionGlyphs";

export type SlashMenuItem = {
  id: string;
  label: string;
  hint?: string;
  keywords: string[];
};

type ListProps = {
  items: SlashMenuItem[];
  command: (item: SlashMenuItem) => void;
  query?: string;
  selectedIndex?: number;
};

export function filterSlashItems(
  items: SlashMenuItem[],
  rawQuery: string,
): SlashMenuItem[] {
  const q = rawQuery.trim().toLowerCase();
  if (q === "") return items;
  return items.filter((item) => {
    if (item.id.toLowerCase().includes(q)) return true;
    if (item.label.toLowerCase().includes(q)) return true;
    return item.keywords.some((k) => k.toLowerCase().includes(q));
  });
}

/** Slash menu popover list — mirrors the repo file suggestion list layout. */
export function SlashMenuList({
  items,
  command,
  query = "",
  selectedIndex = -1,
}: ListProps) {
  const listRef = useRef<HTMLUListElement>(null);
  const trimmedQuery = query.trim();

  useEffect(() => {
    const el = listRef.current;
    if (!el || selectedIndex < 0) return;
    const row = el.querySelector<HTMLElement>(
      `[data-index="${selectedIndex}"]`,
    );
    if (row && typeof row.scrollIntoView === "function") {
      row.scrollIntoView({ block: "nearest" });
    }
  }, [selectedIndex]);

  return (
    <div className="mention-dropdown mention-dropdown--slash tiptap-suggestion-list">
      <div className="mention-dropdown__search" aria-hidden="true">
        <span
          className={
            trimmedQuery
              ? "mention-dropdown__search-query"
              : "mention-dropdown__search-placeholder"
          }
        >
          {trimmedQuery ? `/${trimmedQuery}` : "Insert a block or trigger AI"}
        </span>
        <kbd className="mention-dropdown__at-kbd">/</kbd>
      </div>

      <div
        role="presentation"
        className="mention-dropdown__list"
        style={{ maxHeight: 280, overflowY: "auto" }}
      >
        {items.length === 0 ? (
          <div className="mention-dropdown__empty" role="presentation">
            {trimmedQuery
              ? `No commands match “${trimmedQuery}”`
              : "No commands available"}
          </div>
        ) : (
          <ul
            ref={listRef}
            role="listbox"
            aria-label="Slash commands"
            className="slash-menu__options"
            style={{ margin: 0, padding: 0, listStyle: "none" }}
          >
            {items.map((item, i) => {
              const active = i === selectedIndex;
              return (
                <li
                  key={item.id}
                  className={[
                    "mention-option",
                    "slash-menu__option",
                    active ? "mention-option--active" : "",
                  ]
                    .filter(Boolean)
                    .join(" ")}
                  role="option"
                  aria-selected={active}
                  data-index={i}
                >
                  <button type="button" onClick={() => command(item)}>
                    <span className="slash-menu__label">{item.label}</span>
                    {item.hint ? (
                      <span className="slash-menu__hint">{item.hint}</span>
                    ) : null}
                  </button>
                </li>
              );
            })}
          </ul>
        )}
      </div>

      <div className="mention-dropdown__footer" aria-hidden="true">
        <span>
          <span className="mention-dropdown__footer-keys">↑↓</span> to navigate
        </span>
        <span className="mention-dropdown__footer-select">
          <EnterGlyph /> to select
        </span>
      </div>
    </div>
  );
}
