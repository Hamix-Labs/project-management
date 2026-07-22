import { useEffect, useId, useRef, useState } from "react";
import { buildEditorOpenFolderUri } from "@/tasks/task-git/editorOpenUri";
import {
  editorsForMenu,
  getLastEditorId,
  setLastEditorId,
} from "@/tasks/task-git/editors/lastEditorPreference";
import type { EditorId } from "@/tasks/task-git/editors/registry";
import { TaskDetailExternalLinkGlyph } from "./TaskDetailActionGlyphs";

type Props = {
  openPath: string;
};

export function OpenInEditorMenu({ openPath }: Props) {
  const menuId = useId();
  const rootRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const [open, setOpen] = useState(false);
  const [preferredId, setPreferredId] = useState<EditorId>(() =>
    getLastEditorId(),
  );

  useEffect(() => {
    if (!open) return;

    function onDocMouseDown(e: MouseEvent) {
      const target = e.target;
      if (!(target instanceof Node)) return;
      if (rootRef.current?.contains(target)) return;
      setOpen(false);
    }

    function onKey(e: KeyboardEvent) {
      if (e.key !== "Escape") return;
      e.preventDefault();
      e.stopPropagation();
      setOpen(false);
      triggerRef.current?.focus();
    }

    document.addEventListener("mousedown", onDocMouseDown);
    window.addEventListener("keydown", onKey, true);
    return () => {
      document.removeEventListener("mousedown", onDocMouseDown);
      window.removeEventListener("keydown", onKey, true);
    };
  }, [open]);

  const editors = editorsForMenu(preferredId);

  return (
    <div
      ref={rootRef}
      className="task-detail-open-in"
      data-testid="task-detail-open-in"
    >
      <button
        ref={triggerRef}
        type="button"
        className="btn-utility task-detail-open-in-trigger"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={open ? menuId : undefined}
        aria-label="Open worktree in editor"
        data-testid="task-detail-open-in-trigger"
        onClick={() => {
          if (!open) {
            setPreferredId(getLastEditorId());
          }
          setOpen((wasOpen) => !wasOpen);
        }}
      >
        <TaskDetailExternalLinkGlyph className="task-detail-action-glyph" />
        Open in
        <span className="task-detail-open-in-chevron" aria-hidden="true">
          ▾
        </span>
      </button>
      {open ? (
        <ul
          id={menuId}
          role="menu"
          className="task-detail-open-in-menu"
          data-testid="task-detail-open-in-menu"
        >
          {editors.map((editor) => {
            const href = buildEditorOpenFolderUri(openPath, editor.scheme);
            const isPreferred = editor.id === preferredId;
            return (
              <li key={editor.id} role="none">
                <a
                  role="menuitem"
                  className="task-detail-open-in-item"
                  href={href}
                  data-editor-id={editor.id}
                  data-preferred={isPreferred ? "true" : "false"}
                  aria-label={`Open worktree in ${editor.label}`}
                  onClick={() => {
                    setLastEditorId(editor.id);
                    setPreferredId(editor.id);
                    setOpen(false);
                  }}
                >
                  <TaskDetailExternalLinkGlyph className="task-detail-action-glyph" />
                  {editor.label}
                </a>
              </li>
            );
          })}
        </ul>
      ) : null}
    </div>
  );
}
