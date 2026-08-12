export function SearchGlyph() {
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

type IconKind = "file" | "doc" | "config";

export function iconKindFor(path: string): IconKind {
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

export function FileGlyph({ kind }: { kind: IconKind }) {
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

export function EnterGlyph() {
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
