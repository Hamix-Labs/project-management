import type { ReactElement } from "react";
import type { TimelineKind } from "./timelineTypes";

type GlyphProps = { className?: string };

export function TimelineCheckGlyph({ className }: GlyphProps) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M3.5 8.5 6.5 11.5 12.5 4.5"
        stroke="currentColor"
        strokeWidth="2.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function TimelineXGlyph({ className }: GlyphProps) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M4.5 4.5 11.5 11.5M11.5 4.5 4.5 11.5"
        stroke="currentColor"
        strokeWidth="2.2"
        strokeLinecap="round"
      />
    </svg>
  );
}

export function TimelinePlayGlyph({ className }: GlyphProps) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M5.5 3.75v8.5L12.25 8 5.5 3.75Z"
        fill="currentColor"
      />
    </svg>
  );
}

export function TimelineBotGlyph({ className }: GlyphProps) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <rect
        x="3"
        y="5"
        width="10"
        height="8"
        rx="2"
        stroke="currentColor"
        strokeWidth="1.5"
      />
      <path
        d="M8 2.5v2.5"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
      <circle cx="6" cy="9" r="1" fill="currentColor" />
      <circle cx="10" cy="9" r="1" fill="currentColor" />
    </svg>
  );
}

export function TimelineArrowGlyph({ className }: GlyphProps) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M3.5 8h8M8.5 4.5 12 8l-3.5 3.5"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function TimelinePlusGlyph({ className }: GlyphProps) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M8 3.5v9M3.5 8h9"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  );
}

export function TimelineCommentGlyph({ className }: GlyphProps) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M3.5 3.5h9A1.5 1.5 0 0 1 14 5v4.5A1.5 1.5 0 0 1 12.5 11H8l-3.5 2.5V11H3.5A1.5 1.5 0 0 1 2 9.5V5A1.5 1.5 0 0 1 3.5 3.5Z"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function TimelineCalendarGlyph({ className }: GlyphProps) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <rect
        x="2.5"
        y="3.5"
        width="11"
        height="10"
        rx="1.5"
        stroke="currentColor"
        strokeWidth="1.4"
      />
      <path
        d="M2.5 6.5h11M5.5 2.5v2M10.5 2.5v2"
        stroke="currentColor"
        strokeWidth="1.4"
        strokeLinecap="round"
      />
    </svg>
  );
}

export function TimelineThumbUpGlyph({ className }: GlyphProps) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M5.5 7 7 3a1.5 1.5 0 0 1 1.5 1.5V6.5H12a1 1 0 0 1 .98 1.2l-1 4.5A1 1 0 0 1 11 13H5.5V7Z"
        stroke="currentColor"
        strokeWidth="1.4"
        strokeLinejoin="round"
      />
      <path
        d="M5.5 7H3.5a1 1 0 0 0-1 1v4a1 1 0 0 0 1 1h2V7Z"
        stroke="currentColor"
        strokeWidth="1.4"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function TimelineChevronGlyph({ className }: GlyphProps) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M4.5 6.5 8 10l3.5-3.5"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function TimelineCheckSmallGlyph({ className }: GlyphProps) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M3.5 8.5 6.5 11.5 12.5 4.5"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

const KIND_GLYPH: Record<TimelineKind, (props: GlyphProps) => ReactElement> = {
  "verification-passed": TimelineCheckGlyph,
  "verification-failed": TimelineXGlyph,
  "agent-started": TimelinePlayGlyph,
  "agent-finished": TimelineBotGlyph,
  "status-changed": TimelineArrowGlyph,
  "task-created": TimelinePlusGlyph,
  "review-approved": TimelineThumbUpGlyph,
  comment: TimelineCommentGlyph,
};

export function TimelineKindGlyph({
  kind,
  className,
}: {
  kind: TimelineKind;
  className?: string;
}) {
  const Glyph = KIND_GLYPH[kind];
  return <Glyph className={className} />;
}
