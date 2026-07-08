type IconProps = {
  className?: string;
};

/** Three stacked boxes — matches the project picker icon in the create-task redesign. */
export function ProjectsStackIcon({ className }: IconProps) {
  return (
    <svg
      className={className}
      width={16}
      height={16}
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden
    >
      <path
        d="M12 3 3 7.5 12 12l9-4.5L12 3Z"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinejoin="round"
      />
      <path
        d="m3 12 9 4.5 9-4.5"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinejoin="round"
      />
      <path
        d="m3 16.5 9 4.5 9-4.5"
        stroke="currentColor"
        strokeWidth={1.75}
        strokeLinejoin="round"
      />
    </svg>
  );
}
