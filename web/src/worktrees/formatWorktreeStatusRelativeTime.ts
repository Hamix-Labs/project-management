/**
 * Prose relative-time strings for worktree checkout status rows (reference UI).
 * Differs from shared formatRelativeTime compact buckets (`2 h ago`).
 */
export function formatWorktreeStatusRelativeTime(
  iso: string | null | undefined,
  now: Date = new Date(),
): string {
  if (!iso) return "";
  const then = new Date(iso);
  const t = then.getTime();
  if (!Number.isFinite(t)) return "";

  const deltaMs = now.getTime() - t;
  if (deltaMs < 45_000) return "just now";

  const minutes = Math.floor(deltaMs / 60_000);
  if (minutes < 60) {
    return minutes === 1 ? "1 minute ago" : `${minutes} minutes ago`;
  }

  const hours = Math.floor(deltaMs / 3_600_000);
  if (hours < 24) {
    return hours === 1 ? "1 hour ago" : `${hours} hours ago`;
  }

  const days = Math.floor(deltaMs / 86_400_000);
  if (days === 1) return "yesterday";
  if (days < 7) {
    return `${days} days ago`;
  }

  const weeks = Math.floor(days / 7);
  if (days < 30) {
    return weeks === 1 ? "1 week ago" : `${weeks} weeks ago`;
  }

  const months = Math.floor(days / 30);
  if (days < 365) {
    return months === 1 ? "1 month ago" : `${months} months ago`;
  }

  const years = Math.floor(days / 365);
  return years === 1 ? "1 year ago" : `${years} years ago`;
}
