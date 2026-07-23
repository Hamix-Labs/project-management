export function SparkleGlyph({ size = 16 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.4"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M8 2.25L9.35 6.4 13.5 7.75 9.35 9.1 8 13.25 6.65 9.1 2.5 7.75 6.65 6.4z" />
      <path d="M12.5 2.25v2" />
      <path d="M11.5 3.25h2" />
    </svg>
  );
}

export function CloseGlyph() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      aria-hidden="true"
    >
      <path d="M4 4l8 8M12 4l-8 8" />
    </svg>
  );
}
