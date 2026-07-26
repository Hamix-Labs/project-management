type IconProps = {
  className?: string;
};

const svgProps = {
  width: 16,
  height: 16,
  viewBox: "0 0 24 24",
  fill: "none" as const,
  "aria-hidden": true as const,
};

/** Bot mark for the Agent config section header. */
export function AgentBotIcon({ className }: IconProps) {
  return (
    <svg className={className} {...svgProps}>
      <rect
        x="5"
        y="9"
        width="14"
        height="10"
        rx="2"
        stroke="currentColor"
        strokeWidth={1.75}
      />
      <path
        d="M12 5v4M9 14h.01M15 14h.01"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinecap="round"
      />
      <circle cx="9" cy="14" r="0.75" fill="currentColor" />
      <circle cx="15" cy="14" r="0.75" fill="currentColor" />
    </svg>
  );
}

/** Microchip mark for the Runner select trigger. */
export function AgentCpuIcon({ className }: IconProps) {
  return (
    <svg className={className} {...svgProps}>
      <rect
        x="7"
        y="7"
        width="10"
        height="10"
        rx="1.5"
        stroke="currentColor"
        strokeWidth={1.75}
      />
      <path
        d="M9 1v3M15 1v3M9 20v3M15 20v3M1 9h3M1 15h3M20 9h3M20 15h3"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinecap="round"
      />
      <rect x="10" y="10" width="4" height="4" rx="0.5" fill="currentColor" />
    </svg>
  );
}

/** Shield-check mark for the Verify chat select trigger. */
export function AgentShieldCheckIcon({ className }: IconProps) {
  return (
    <svg className={className} {...svgProps}>
      <path
        d="M12 3 5 6v5c0 4.5 3 7.5 7 9 4-1.5 7-4.5 7-9V6l-7-3Z"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinejoin="round"
      />
      <path
        d="m9.5 12 1.75 1.75L15 10"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

/** Tag mark for the Tags config section header. */
export function AgentTagIcon({ className }: IconProps) {
  return (
    <svg className={className} {...svgProps}>
      <path
        d="M20 12.5 12.5 20a2 2 0 0 1-2.8 0L4 14.3V4h10.3l5.7 5.7a2 2 0 0 1 0 2.8Z"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinejoin="round"
      />
      <circle cx="8.5" cy="8.5" r="1.25" fill="currentColor" />
    </svg>
  );
}

/** Calendar mark for the Schedule config section header. */
export function AgentCalendarIcon({ className }: IconProps) {
  return (
    <svg className={className} {...svgProps}>
      <rect
        x="3"
        y="5"
        width="18"
        height="16"
        rx="2"
        stroke="currentColor"
        strokeWidth={1.75}
      />
      <path
        d="M3 10h18M8 3v4M16 3v4"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinecap="round"
      />
    </svg>
  );
}

/** Status / list mark for edit-status and dependencies section headers. */
export function AgentListIcon({ className }: IconProps) {
  return (
    <svg className={className} {...svgProps}>
      <path
        d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinecap="round"
      />
    </svg>
  );
}
