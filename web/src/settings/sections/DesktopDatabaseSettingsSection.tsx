import { useEffect, useState } from "react";
import { SectionCard } from "./settingsSectionLayout";
import { SECTION_IDS } from "./sectionIds";
import {
  getDesktopBridge,
  isDesktopHost,
  type DesktopBridge,
} from "@/desktop/bridge";
import { DesktopDatabaseForm } from "@/desktop/DesktopDatabaseForm";

/**
 * Desktop-only: change the local Postgres URL (restart required).
 * Hidden in the browser against taskapi.
 */
export function DesktopDatabaseSettingsSection() {
  const [bridge, setBridge] = useState<DesktopBridge | null>(null);
  const [initialUrl, setInitialUrl] = useState("");

  useEffect(() => {
    if (!isDesktopHost()) return;
    const b = getDesktopBridge();
    if (!b) return;
    setBridge(b);
    void b.getDatabaseConfig().then((cfg) => setInitialUrl(cfg.url));
  }, []);

  if (!bridge) return null;

  return (
    <SectionCard id={SECTION_IDS.database} title="Database connection">
      <p className="settings-field-help">
        Local desktop config only. Changing the URL requires restarting Hamix.
      </p>
      <DesktopDatabaseForm
        bridge={bridge}
        initialUrl={initialUrl}
        editing
      />
    </SectionCard>
  );
}
