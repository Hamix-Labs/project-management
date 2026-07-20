import { useQuery, type UseQueryResult } from "@tanstack/react-query";
import {
  type AppSettings,
  fetchAppSettings,
} from "@/api/settings";
import { settingsQueryKeys } from "@/lib/settingsQueryKeys";

/**
 * Read-only AppSettings query for inner-ring consumers (shared time,
 * task create, etc.). Mutations stay in `settings/useAppSettings`.
 */
export function useAppSettingsQuery(): UseQueryResult<AppSettings, Error> {
  return useQuery<AppSettings>({
    queryKey: settingsQueryKeys.app(),
    queryFn: ({ signal }) => fetchAppSettings({ signal }),
  });
}
