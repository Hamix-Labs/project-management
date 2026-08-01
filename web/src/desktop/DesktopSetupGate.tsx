import { useEffect, useState } from "react";
import {
  getDesktopBridge,
  isDesktopHost,
  type DesktopBridge,
} from "./bridge";
import { DesktopSetupPage } from "./DesktopSetupPage";

type Phase = "loading" | "setup" | "app";

type Props = {
  children: React.ReactNode;
};

/**
 * When running in the Wails host without a DSN, show setup instead of the
 * main SPA (which would call the API). Browser / configured desktop → children.
 */
export function DesktopSetupGate({ children }: Props) {
  const [phase, setPhase] = useState<Phase>(() =>
    isDesktopHost() ? "loading" : "app",
  );
  const [bridge, setBridge] = useState<DesktopBridge | null>(null);

  useEffect(() => {
    if (!isDesktopHost()) {
      setPhase("app");
      return;
    }
    const b = getDesktopBridge();
    if (!b) {
      setPhase("app");
      return;
    }
    setBridge(b);
    let cancelled = false;
    void (async () => {
      try {
        const cfg = await b.getDatabaseConfig();
        if (cancelled) return;
        setPhase(cfg.needsSetup ? "setup" : "app");
      } catch {
        if (!cancelled) setPhase("setup");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  if (phase === "loading") {
    return (
      <div className="desktop-setup desktop-setup-loading" data-testid="desktop-setup-loading">
        Loading…
      </div>
    );
  }
  if (phase === "setup" && bridge) {
    return <DesktopSetupPage bridge={bridge} />;
  }
  return <>{children}</>;
}
