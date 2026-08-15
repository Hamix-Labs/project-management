type Props = {
  count: number;
};

/** Compact pill for criteria that include execute-agent verify checks. */
export function ChecklistVerifyBadge({ count }: Props) {
  if (count <= 0) {
    return null;
  }

  const noun = count === 1 ? "command" : "commands";
  const label = `${count} ${noun}`;

  return (
    <span
      className="task-checklist-verify-badge"
      aria-label={`${count} verify ${noun} for the execute agent`}
    >
      <svg
        className="task-checklist-verify-badge__icon"
        width={12}
        height={12}
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        aria-hidden
      >
        <path
          d="M7 8.5 11 12l-4 3.5"
          stroke="currentColor"
          strokeWidth={1.75}
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M12.5 16.5h5"
          stroke="currentColor"
          strokeWidth={1.75}
          strokeLinecap="round"
        />
      </svg>
      <span className="task-checklist-verify-badge__label">{label}</span>
    </span>
  );
}
