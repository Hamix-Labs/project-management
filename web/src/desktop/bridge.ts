/**
 * Sole WebView IPC call site for Hamix desktop (Wails).
 * Pages must not call window.go directly — use this module (ADR-0095).
 */

export type DatabaseConfig = {
  url: string;
  source: string;
  configured: boolean;
  needsSetup: boolean;
};

export type DesktopBridge = {
  getDatabaseConfig: () => Promise<DatabaseConfig>;
  saveDatabaseConfig: (url: string) => Promise<void>;
  testDatabaseConnection: (url: string) => Promise<void>;
  quitApp: () => Promise<void>;
};

type WailsApp = {
  GetDatabaseConfig: () => Promise<DatabaseConfig>;
  SaveDatabaseConfig: (url: string) => Promise<void>;
  TestDatabaseConnection: (url: string) => Promise<void>;
  QuitApp: () => void | Promise<void>;
};

declare global {
  interface Window {
    go?: {
      main?: {
        App?: WailsApp;
      };
    };
  }
}

let testBridge: DesktopBridge | null = null;

/** Test-only: inject a fake bridge (Vitest). */
export function setDesktopBridgeForTests(bridge: DesktopBridge | null): void {
  testBridge = bridge;
}

function wailsApp(): WailsApp | null {
  return window.go?.main?.App ?? null;
}

function fromWails(app: WailsApp): DesktopBridge {
  return {
    getDatabaseConfig: () => app.GetDatabaseConfig(),
    saveDatabaseConfig: (url) => app.SaveDatabaseConfig(url),
    testDatabaseConnection: (url) => app.TestDatabaseConnection(url),
    quitApp: async () => {
      await app.QuitApp();
    },
  };
}

/** True when running inside the Wails desktop host (or a test bridge). */
export function isDesktopHost(): boolean {
  return testBridge != null || wailsApp() != null;
}

export function getDesktopBridge(): DesktopBridge | null {
  if (testBridge) return testBridge;
  const app = wailsApp();
  return app ? fromWails(app) : null;
}
