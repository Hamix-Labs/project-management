import type { LiveTimelineIconRole } from "@/tasks/cycleDisplay/liveTimelineRows";

type IconProps = {
  className?: string;
};

function WorkingIcon({ className }: IconProps) {
  return (
    <svg
      className={["cycle-live-icon", "cycle-live-icon--spin", className]
        .filter(Boolean)
        .join(" ")}
      width={14}
      height={14}
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden
    >
      <circle
        cx={12}
        cy={12}
        r={9}
        stroke="currentColor"
        strokeWidth={2}
        strokeOpacity={0.25}
      />
      <path
        d="M21 12a9 9 0 0 0-9-9"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
      />
    </svg>
  );
}

function DoneIcon({ className }: IconProps) {
  return (
    <span
      className={["cycle-live-icon", "cycle-live-icon--done", className]
        .filter(Boolean)
        .join(" ")}
      aria-hidden
    >
      <svg width={12} height={12} viewBox="0 0 24 24" fill="none">
        <path
          d="M5 12.5 10 17.5 19 7"
          stroke="currentColor"
          strokeWidth={2.5}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </span>
  );
}

function CallIcon({ className }: IconProps) {
  return (
    <span
      className={["cycle-live-icon", "cycle-live-icon--call", className]
        .filter(Boolean)
        .join(" ")}
      aria-hidden
    >
      <svg width={12} height={12} viewBox="0 0 24 24" fill="none">
        <path
          d="M7 17 17 7M10 7h7v7"
          stroke="currentColor"
          strokeWidth={2.5}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </span>
  );
}

function FailedIcon({ className }: IconProps) {
  return (
    <span
      className={["cycle-live-icon", "cycle-live-icon--failed", className]
        .filter(Boolean)
        .join(" ")}
      aria-hidden
    >
      <svg width={12} height={12} viewBox="0 0 24 24" fill="none">
        <path
          d="M7 7l10 10M17 7 7 17"
          stroke="currentColor"
          strokeWidth={2.5}
          strokeLinecap="round"
        />
      </svg>
    </span>
  );
}

function NeutralIcon({ className }: IconProps) {
  return (
    <span
      className={["cycle-live-icon", "cycle-live-icon--neutral", className]
        .filter(Boolean)
        .join(" ")}
      aria-hidden
    >
      <svg width={12} height={12} viewBox="0 0 24 24" fill="none">
        <circle cx={12} cy={12} r={3.5} fill="currentColor" />
      </svg>
    </span>
  );
}

/** Decorative icon for a live timeline row; always aria-hidden. */
export function CycleLiveProgressIcon({
  role,
  className,
}: {
  role: LiveTimelineIconRole;
  className?: string;
}) {
  if (role === "working") return <WorkingIcon className={className} />;
  if (role === "done") return <DoneIcon className={className} />;
  if (role === "call") return <CallIcon className={className} />;
  if (role === "failed") return <FailedIcon className={className} />;
  return <NeutralIcon className={className} />;
}
