import { useDocumentTitle } from "@/shared/useDocumentTitle";
import {
  DisplaySettingsSection,
  PhasesSettingsSection,
  RunnerSettingsSection,
  SettingsActions,
  SettingsHeader,
  SettingsLoadingState,
  SettingsStatusMessage,
} from "./sections";
import { UiTestModeSettingsSection } from "./UiTestModeSettingsSection";
import { SettingsNav, type SettingsNavItem } from "./SettingsNav";
import { SECTION_IDS } from "./sections/sectionIds";
import {
  useSettingsPageModel,
  type SettingsPageLoadedViewProps,
} from "./useSettingsPageModel";
import "./settings.css";

const SETTINGS_NAV_ITEMS: SettingsNavItem[] = [
  { id: SECTION_IDS.cursorAgent, label: "Runner" },
  { id: SECTION_IDS.phases, label: "Phases" },
  { id: SECTION_IDS.display, label: "Display" },
  { id: SECTION_IDS.developer, label: "Developer" },
];

function SettingsPageLoadedView({
  form,
  status,
  resolvedDefaultBin,
  isDirty,
  numericValidation,
  tzSelectOptions,
  browserTz,
  showCustomTz,
  lastUpdated,
  lastUpdatedFormatted,
  cursorModelsQuery,
  modelIdsFromList,
  verifyModelsQuery,
  verifyModelIdsFromList,
  patchPending,
  probePending,
  onField,
  onSubmit,
  onProbe,
  onDiscard,
}: SettingsPageLoadedViewProps) {
  const { maxInvalid, streamIdleInvalid, pickupInvalid } = numericValidation;

  return (
    <section className="settings-page">
      <SettingsHeader
        lastUpdated={lastUpdated}
        lastUpdatedFormatted={lastUpdatedFormatted}
      />

      <div className="settings-layout">
        <aside className="settings-layout-aside">
          <SettingsNav items={SETTINGS_NAV_ITEMS} />
        </aside>

        <form className="settings-form" onSubmit={onSubmit}>
          <RunnerSettingsSection
            form={form}
            resolvedDefaultBin={resolvedDefaultBin}
            probePending={probePending}
            onField={onField}
            onProbe={onProbe}
          />

          <PhasesSettingsSection
            form={form}
            pickupInvalid={pickupInvalid}
            maxInvalid={maxInvalid}
            streamIdleInvalid={streamIdleInvalid}
            cursorModelsQuery={cursorModelsQuery}
            modelIdsFromList={modelIdsFromList}
            verifyModelsQuery={verifyModelsQuery}
            verifyModelIdsFromList={verifyModelIdsFromList}
            onField={onField}
          />

          <DisplaySettingsSection
            form={form}
            browserTz={browserTz}
            options={tzSelectOptions}
            showCustomTz={showCustomTz}
            onField={onField}
          />

          <UiTestModeSettingsSection />

          <SettingsStatusMessage status={status} />

          <SettingsActions
            isDirty={isDirty}
            maxInvalid={maxInvalid}
            streamIdleInvalid={streamIdleInvalid}
            pickupInvalid={pickupInvalid}
            patchPending={patchPending}
            onDiscard={onDiscard}
          />
        </form>
      </div>
    </section>
  );
}

export function SettingsPage() {
  useDocumentTitle("Settings");
  const { error, refetch, loadedViewProps } = useSettingsPageModel();

  if (!loadedViewProps) {
    return (
      <SettingsLoadingState
        error={error}
        onRetry={() => {
          void refetch();
        }}
      />
    );
  }

  return <SettingsPageLoadedView {...loadedViewProps} />;
}

// Exported for tests that need the loaded view props shape.
export type { SettingsPageLoadedViewProps };
