export function SettingsSelectChevronIcon({ open }: { open: boolean }) {
  return (
    <span
      className={
        open
          ? "settings-dropdown-chevron settings-dropdown-chevron--open"
          : "settings-dropdown-chevron"
      }
      aria-hidden="true"
    >
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
        <path
          d="M4 6l4 4 4-4"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </span>
  );
}

export function SettingsSelectCheckIcon() {
  return (
    <svg
      className="settings-dropdown-option-check"
      width="14"
      height="14"
      viewBox="0 0 14 14"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M2.5 7.25 5.5 10.25 11.5 3.75"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
