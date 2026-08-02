import type { SuggestionMenuProps } from "@blocknote/react";
import { splitRepoPath } from "../repoFileRef";

export type PromptFileMentionItem = {
  title: string;
  onItemClick: () => void;
  query: string;
};

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

  return (
    <div className="prompt-editor-mention-menu" role="listbox">
      <div className="prompt-editor-mention-menu__search">
        {loadingState === "loading" || loadingState === "loading-initial"
          ? "Searching files…"
          : (
            <>
              Searching files for <b>{query || "…"}</b>
            </>
          )}
      </div>
      {items.map((item, i) => {
        const { fileName, dirPath } = splitRepoPath(item.title);
        return (
          <button
            key={item.title}
            type="button"
            role="option"
            aria-selected={i === selectedIndex}
            data-selected={i === selectedIndex ? "true" : "false"}
            className="prompt-editor-mention-menu__item"
            onClick={() => onItemClick?.(item)}
          >
            <FileIcon />
            <span className="prompt-editor-mention-menu__fname">{fileName}</span>
            {dirPath ? (
              <span className="prompt-editor-mention-menu__fpath">{dirPath}</span>
            ) : null}
          </button>
        );
      })}
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
