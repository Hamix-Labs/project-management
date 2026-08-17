import type { Editor } from "@tiptap/core";
import type { ReactNode } from "react";

type IconProps = { className?: string };

function BoldIcon({ className }: IconProps) {
  return (
    <svg className={className} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M6 4h8a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z" />
      <path d="M6 12h9a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z" />
    </svg>
  );
}

function ItalicIcon({ className }: IconProps) {
  return (
    <svg className={className} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <line x1="19" y1="4" x2="10" y2="4" />
      <line x1="14" y1="20" x2="5" y2="20" />
      <line x1="15" y1="4" x2="9" y2="20" />
    </svg>
  );
}

function Heading2Icon({ className }: IconProps) {
  return (
    <svg className={className} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M4 12h8" />
      <path d="M4 18V6" />
      <path d="M12 18V6" />
      <path d="M21 18h-4c0-4 4-3 4-6 0-1.5-2-2.5-4-1" />
    </svg>
  );
}

function Heading3Icon({ className }: IconProps) {
  return (
    <svg className={className} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M4 12h8" />
      <path d="M4 18V6" />
      <path d="M12 18V6" />
      <path d="M17.5 10.5c1.7-1 3.5 0 3.5 1.5a2 2 0 0 1-2 2" />
      <path d="M17 17.5c2 1.5 4 .3 4-1.5a2 2 0 0 0-2-2" />
    </svg>
  );
}

function ListIcon({ className }: IconProps) {
  return (
    <svg className={className} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <line x1="8" y1="6" x2="21" y2="6" />
      <line x1="8" y1="12" x2="21" y2="12" />
      <line x1="8" y1="18" x2="21" y2="18" />
      <line x1="3" y1="6" x2="3.01" y2="6" />
      <line x1="3" y1="12" x2="3.01" y2="12" />
      <line x1="3" y1="18" x2="3.01" y2="18" />
    </svg>
  );
}

function ListOrderedIcon({ className }: IconProps) {
  return (
    <svg className={className} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <line x1="10" y1="6" x2="21" y2="6" />
      <line x1="10" y1="12" x2="21" y2="12" />
      <line x1="10" y1="18" x2="21" y2="18" />
      <path d="M4 6h1v4" />
      <path d="M4 10h2" />
      <path d="M6 18H4c0-1 2-2 2-3s-1-1.5-2-1" />
    </svg>
  );
}

function CodeIcon({ className }: IconProps) {
  return (
    <svg className={className} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <polyline points="16 18 22 12 16 6" />
      <polyline points="8 6 2 12 8 18" />
    </svg>
  );
}

export type RichPromptMenuBarProps = {
  editor: Editor | null;
  disabled?: boolean;
  /**
   * `"icons"` renders SVG glyphs; `"text"` keeps legacy labels;
   * `"none"` hides formatting buttons (slash menu is the insert path).
   */
  variant?: "text" | "icons" | "none";
  /** Optional trailing slot (e.g. word count). */
  right?: ReactNode;
};

export function RichPromptMenuBar({
  editor,
  disabled,
  variant = "text",
  right,
}: RichPromptMenuBarProps) {
  if (!editor) return null;
  if (variant === "none") {
    if (!right) return null;
    return (
      <div className="rich-prompt-toolbar" data-variant="none">
        <span className="rich-prompt-toolbar__right">{right}</span>
      </div>
    );
  }
  const d = Boolean(disabled);
  const icons = variant === "icons";

  return (
    <div
      className="rich-prompt-toolbar"
      role="toolbar"
      aria-label="Text formatting"
      data-variant={variant}
    >
      <button
        type="button"
        className="secondary toolbar-btn"
        disabled={d || !editor.can().chain().focus().toggleBold().run()}
        onClick={() => editor.chain().focus().toggleBold().run()}
        aria-pressed={editor.isActive("bold")}
        aria-label="Bold"
        title="Bold"
      >
        {icons ? <BoldIcon /> : "Bold"}
      </button>
      <button
        type="button"
        className="secondary toolbar-btn"
        disabled={d || !editor.can().chain().focus().toggleItalic().run()}
        onClick={() => editor.chain().focus().toggleItalic().run()}
        aria-pressed={editor.isActive("italic")}
        aria-label="Italic"
        title="Italic"
      >
        {icons ? <ItalicIcon /> : "Italic"}
      </button>
      {icons ? <span className="rich-prompt-toolbar__sep" aria-hidden="true" /> : null}
      <button
        type="button"
        className="secondary toolbar-btn"
        disabled={d}
        onClick={() =>
          editor.chain().focus().toggleHeading({ level: 2 }).run()
        }
        aria-pressed={editor.isActive("heading", { level: 2 })}
        aria-label="Heading 2"
        title="Heading 2"
      >
        {icons ? <Heading2Icon /> : "H2"}
      </button>
      <button
        type="button"
        className="secondary toolbar-btn"
        disabled={d}
        onClick={() =>
          editor.chain().focus().toggleHeading({ level: 3 }).run()
        }
        aria-pressed={editor.isActive("heading", { level: 3 })}
        aria-label="Heading 3"
        title="Heading 3"
      >
        {icons ? <Heading3Icon /> : "H3"}
      </button>
      {icons ? <span className="rich-prompt-toolbar__sep" aria-hidden="true" /> : null}
      <button
        type="button"
        className="secondary toolbar-btn"
        disabled={d || !editor.can().chain().focus().toggleBulletList().run()}
        onClick={() => editor.chain().focus().toggleBulletList().run()}
        aria-pressed={editor.isActive("bulletList")}
        aria-label="Bulleted list"
        title="Bulleted list"
      >
        {icons ? <ListIcon /> : "List"}
      </button>
      <button
        type="button"
        className="secondary toolbar-btn"
        disabled={d || !editor.can().chain().focus().toggleOrderedList().run()}
        onClick={() => editor.chain().focus().toggleOrderedList().run()}
        aria-pressed={editor.isActive("orderedList")}
        aria-label="Numbered list"
        title="Numbered list"
      >
        {icons ? <ListOrderedIcon /> : "Numbered"}
      </button>
      {icons ? <span className="rich-prompt-toolbar__sep" aria-hidden="true" /> : null}
      <button
        type="button"
        className="secondary toolbar-btn"
        disabled={d || !editor.can().chain().focus().toggleCode().run()}
        onClick={() => editor.chain().focus().toggleCode().run()}
        aria-pressed={editor.isActive("code")}
        aria-label="Code"
        title="Code"
      >
        {icons ? <CodeIcon /> : "Code"}
      </button>
      {!icons ? (
        <button
          type="button"
          className="secondary toolbar-btn"
          disabled={d}
          onClick={() => editor.chain().focus().setParagraph().run()}
        >
          Paragraph
        </button>
      ) : null}
      {right ? (
        <span className="rich-prompt-toolbar__right">{right}</span>
      ) : null}
    </div>
  );
}
