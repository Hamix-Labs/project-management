import { UiTestModeBanner } from "@/dev/UiTestModeBanner";
import { useTasksAppMeta } from "@/tasks/app/TasksAppProvider";
import { ErrorBanner } from "../shared/ErrorBanner";
import { RouteAnnouncer } from "./RouteAnnouncer";
import { RoutedMainOutlet } from "./RoutedMainOutlet";

/**
 * Full-viewport layout for surfaces that own their own chrome (e.g. Prompt Editor).
 * No product nav — keep TasksAppProvider above so compose suspend/resume still works.
 */
export function ImmersiveShell() {
  const { error } = useTasksAppMeta();

  return (
    <div className="app app--immersive">
      <a href="#main-content" className="skip-link">
        Skip to main content
      </a>
      <UiTestModeBanner />
      {error ? <ErrorBanner message={error} /> : null}
      <main id="main-content" tabIndex={-1}>
        <RoutedMainOutlet />
      </main>
      <RouteAnnouncer />
    </div>
  );
}
