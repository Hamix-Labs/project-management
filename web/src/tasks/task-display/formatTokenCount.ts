export type FormattedTokenCount = {
  label: string;
  ariaLabel: string;
};

function trimTrailingZeroDecimal(value: string): string {
  return value.replace(/\.0$/, "");
}

function formatScaledCount(value: number, divisor: number, suffix: string): string {
  const scaled = value / divisor;
  if (Math.abs(scaled) >= 100) {
    return `${Math.round(scaled)}${suffix}`;
  }
  return `${trimTrailingZeroDecimal(scaled.toFixed(1))}${suffix}`;
}

/** Compact token count for UI labels; full number lives in `ariaLabel`. */
export function formatTokenCount(value: number): FormattedTokenCount {
  const ariaLabel = `${value.toLocaleString()} tokens`;
  const abs = Math.abs(value);

  if (abs < 1000) {
    return { label: value.toLocaleString(), ariaLabel };
  }
  if (abs < 1_000_000) {
    return {
      label: formatScaledCount(value, 1000, "K"),
      ariaLabel,
    };
  }
  if (abs < 1_000_000_000) {
    return {
      label: formatScaledCount(value, 1_000_000, "M"),
      ariaLabel,
    };
  }
  return {
    label: formatScaledCount(value, 1_000_000_000, "B"),
    ariaLabel,
  };
}

/** Percentage of task tokens consumed by one attempt. */
export function formatShareOfTaskPct(value: number): string {
  const abs = Math.abs(value);
  if (abs >= 100) {
    return `${Math.round(value)}%`;
  }
  return `${trimTrailingZeroDecimal(value.toFixed(1))}%`;
}
