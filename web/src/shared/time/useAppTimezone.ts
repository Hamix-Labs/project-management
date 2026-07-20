import { useAppSettingsQuery } from "@/hooks/useAppSettingsQuery";
import { detectBrowserTimezone } from "./appTimezonePure";

/**
 * useAppTimezone returns the IANA timezone the SPA should use to
 * render every operator-facing timestamp.
 *
 * Precedence (highest to lowest):
 *  1. `settings.display_timezone` — a non-empty explicit override
 *     chosen in the SettingsPage selector and validated server-side
 *     via `time.LoadLocation`. Always wins.
 *  2. The operator's browser timezone (`detectBrowserTimezone()`).
 *     Used whenever the server returns the empty-string
 *     "auto-detect" sentinel (the default seed) OR when the settings
 *     query is still loading, so the first paint already lands in
 *     local time rather than flashing UTC for a frame.
 *  3. DEFAULT_APP_TIMEZONE ("UTC") — only if the Intl API refuses to
 *     produce a zone at all. See `detectBrowserTimezone`.
 *
 * Stage 1 of the task scheduling plan introduced the field; later
 * stages (3–5) call this hook from every timestamp render so a single
 * PATCH /settings { display_timezone } re-renders the whole SPA in
 * the chosen zone via React Query invalidation.
 */
export function useAppTimezone(): string {
  const { data: settings } = useAppSettingsQuery();
  if (!settings) return detectBrowserTimezone();
  const tz = settings.display_timezone;
  if (typeof tz !== "string" || tz.length === 0) {
    return detectBrowserTimezone();
  }
  return tz;
}
