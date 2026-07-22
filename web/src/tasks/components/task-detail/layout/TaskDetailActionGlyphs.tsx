/** Inline SVGs for task-detail action chrome — no lucide dependency. */

type GlyphProps = {
  className?: string;
};

const base = {
  width: 16,
  height: 16,
  viewBox: "0 0 24 24",
  fill: "none" as const,
  xmlns: "http://www.w3.org/2000/svg",
  "aria-hidden": true as const,
};

export function TaskDetailCopyGlyph({ className }: GlyphProps) {
  return (
    <svg className={className} {...base}>
      <rect
        x="9"
        y="9"
        width="13"
        height="13"
        rx="2"
        stroke="currentColor"
        strokeWidth={2}
      />
      <path
        d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function TaskDetailCheckGlyph({ className }: GlyphProps) {
  return (
    <svg className={className} {...base}>
      <path
        d="M20 6 9 17l-5-5"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function TaskDetailExternalLinkGlyph({ className }: GlyphProps) {
  return (
    <svg className={className} {...base}>
      <path
        d="M15 3h6v6"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M10 14 21 3"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function TaskDetailPlusGlyph({ className }: GlyphProps) {
  return (
    <svg className={className} {...base}>
      <path
        d="M12 5v14M5 12h14"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
      />
    </svg>
  );
}

export function TaskDetailClockGlyph({ className }: GlyphProps) {
  return (
    <svg className={className} {...base}>
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth={2} />
      <path
        d="M12 7v5l3 2"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function TaskDetailGitBranchGlyph({ className }: GlyphProps) {
  return (
    <svg className={className} {...base}>
      <circle cx="6" cy="6" r="2.5" stroke="currentColor" strokeWidth={2} />
      <circle cx="6" cy="18" r="2.5" stroke="currentColor" strokeWidth={2} />
      <circle cx="18" cy="12" r="2.5" stroke="currentColor" strokeWidth={2} />
      <path
        d="M6 8.5v7M8.5 6h5.5a3 3 0 0 1 3 3v.5"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
      />
    </svg>
  );
}

export function TaskDetailGitCommitGlyph({ className }: GlyphProps) {
  return (
    <svg className={className} {...base}>
      <circle cx="12" cy="12" r="3" stroke="currentColor" strokeWidth={2} />
      <path
        d="M3 12h6M15 12h6"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
      />
    </svg>
  );
}
