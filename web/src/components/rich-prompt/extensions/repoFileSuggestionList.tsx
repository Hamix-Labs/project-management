import { useEffect, useRef } from "react";

export type RepoSuggestionItem = { path: string };

type ListProps = {
  items: RepoSuggestionItem[];
  command: (item: RepoSuggestionItem) => void;
  /** Live TipTap query after `@` — shown in the decorative search header. */
  query?: string;
  /** Keyboard-selected row index (owned by the suggestion plugin). */
  selectedIndex?: number;
};

export function splitPath(path: string): [string, string] {
  const normalized = path.replace(/\\/g, "/");
  const idx = normalized.lastIndexOf("/");
  if (idx === -1) return ["", normalized];
  return [normalized.slice(0, idx + 1), normalized.slice(idx + 1)];
}

type IconKind = "file" | "doc" | "config";

function iconKindFor(path: string): IconKind {
  const name = path.replace(/\\/g, "/").split("/").pop() ?? path;
  if (/\.(md|mdc|txt|rst)$/i.test(name)) return "doc";
  if (
    name.startsWith(".") ||
    /\.(json|ya?ml|toml|gitignore|editorconfig|lock)$/i.test(name)
  ) {
    return "config";
  }
  return "file";
}

function SearchGlyph() {
  return (
    <svg
      className="mention-dropdown__glyph"
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <circle cx="11" cy="11" r="7" />
      <path d="m20 20-3.5-3.5" />
    </svg>
  );
}

function FileGlyph({ kind }: { kind: IconKind }) {
  if (kind === "doc") {
    return (
      <svg
        className="mention-dropdown__glyph"
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
      >
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
        <path d="M14 2v6h6" />
        <path d="M8 13h8" />
        <path d="M8 17h6" />
      </svg>
    );
  }
  if (kind === "config") {
    return (
      <svg
        className="mention-dropdown__glyph"
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
      >
        <path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
      </svg>
    );
  }
  return (
    <svg
      className="mention-dropdown__glyph"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
      <path d="M14 2v6h6" />
      <path d="M10 12h.01" />
      <path d="M10 16h.01" />
      <path d="M14 12h.01" />
      <path d="M14 16h.01" />
    </svg>
  );
}

function EnterGlyph() {
  return (
    <svg
      className="mention-dropdown__glyph"
      width="12"
      height="12"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M9 10h10v8" />
      <path d="m15 14-4 4 4 4" />
      <path d="M5 4v10a4 4 0 0 0 4 4h1" />
    </svg>
  );
}

export function RepoFileSuggestionList({
  items,
  command,
  query = "",
  selectedIndex = 0,
}: ListProps) {
  const listRef = useRef<HTMLUListElement>(null);
  const trimmedQuery = query.trim();

  useEffect(() => {
    const el = listRef.current?.querySelector<HTMLElement>(
      `[data-index="${selectedIndex}"]`,
    );
    // jsdom does not implement scrollIntoView.
    if (el && typeof el.scrollIntoView === "function") {
      el.scrollIntoView({ block: "nearest" });
    }
  }, [selectedIndex, items]);

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

      <ul
        ref={listRef}
        role="listbox"
        aria-label="Matching repository files"
        className="mention-dropdown__list"
      >
        {items.length === 0 ? (
          <li className="mention-dropdown__empty" role="presentation">
            {trimmedQuery
              ? `No files match “${trimmedQuery}”`
              : "No matching files"}
          </li>
        ) : (
          items.map((item, i) => {
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
          })
        )}
      </ul>

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
