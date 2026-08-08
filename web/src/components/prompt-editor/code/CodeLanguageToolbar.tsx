import { useEffect, useId, useMemo, useRef, useState } from "react";
import type { PromptCodeLanguage } from "./promptCodeBlockOptions";

export type CodeLanguageToolbarProps = {
  languages: PromptCodeLanguage[];
  value: string;
  disabled?: boolean;
  onChange: (languageId: string) => void;
  onCopy: () => void | Promise<void>;
};

export function CodeLanguageToolbar({
  languages,
  value,
  disabled = false,
  onChange,
  onCopy,
}: CodeLanguageToolbarProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [copied, setCopied] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);
  const listId = useId();

  const currentName =
    languages.find((l) => l.id === value)?.name ?? (value || "Plain Text");

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    const sorted = [...languages].sort((a, b) => a.name.localeCompare(b.name));
    if (!q) return sorted;
    return sorted.filter(
      (l) =>
        l.name.toLowerCase().includes(q) || l.id.toLowerCase().includes(q),
    );
  }, [languages, query]);

  useEffect(() => {
    if (!open) return;
    searchRef.current?.focus();
    const onDocPointer = (e: MouseEvent) => {
      if (!rootRef.current?.contains(e.target as Node)) {
        setOpen(false);
        setQuery("");
      }
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setOpen(false);
        setQuery("");
      }
    };
    document.addEventListener("mousedown", onDocPointer);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDocPointer);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  return (
    <div className="prompt-code-toolbar" ref={rootRef} contentEditable={false}>
      <button
        type="button"
        className="prompt-code-toolbar__lang"
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={open ? listId : undefined}
        onClick={() => {
          if (disabled) return;
          setOpen((v) => !v);
          setQuery("");
        }}
      >
        <span>{currentName}</span>
        <svg
          width="12"
          height="12"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          aria-hidden="true"
        >
          <path d="M6 9l6 6 6-6" />
        </svg>
      </button>
      <span className="prompt-code-toolbar__divider" aria-hidden="true" />
      <button
        type="button"
        className="prompt-code-toolbar__copy"
        disabled={disabled}
        title={copied ? "Copied" : "Copy"}
        aria-label={copied ? "Copied" : "Copy code"}
        onClick={() => {
          void Promise.resolve(onCopy()).then(() => {
            setCopied(true);
            window.setTimeout(() => setCopied(false), 1200);
          });
        }}
      >
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.8"
          aria-hidden="true"
        >
          <rect x="9" y="9" width="13" height="13" rx="2" />
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
        </svg>
      </button>

      {open ? (
        <div className="prompt-code-lang-menu" role="presentation">
          <input
            ref={searchRef}
            type="search"
            className="prompt-code-lang-menu__search"
            placeholder="Search for a language…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            aria-label="Search for a language"
            autoComplete="off"
          />
          <ul
            id={listId}
            className="prompt-code-lang-menu__list"
            role="listbox"
            aria-label="Languages"
          >
            {filtered.length === 0 ? (
              <li className="prompt-code-lang-menu__empty">No languages</li>
            ) : (
              filtered.map((lang) => {
                const selected = lang.id === value;
                return (
                  <li key={lang.id} role="option" aria-selected={selected}>
                    <button
                      type="button"
                      className="prompt-code-lang-menu__item"
                      data-selected={selected ? "true" : undefined}
                      onClick={() => {
                        onChange(lang.id);
                        setOpen(false);
                        setQuery("");
                      }}
                    >
                      <span>{lang.name}</span>
                      {selected ? (
                        <svg
                          width="14"
                          height="14"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          strokeWidth="2.2"
                          aria-hidden="true"
                        >
                          <path d="M20 6L9 17l-5-5" />
                        </svg>
                      ) : null}
                    </button>
                  </li>
                );
              })
            )}
          </ul>
        </div>
      ) : null}
    </div>
  );
}
