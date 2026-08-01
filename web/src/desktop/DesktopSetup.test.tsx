import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  setDesktopBridgeForTests,
  type DesktopBridge,
} from "./bridge";
import { DesktopDatabaseForm } from "./DesktopDatabaseForm";
import { DesktopSetupGate } from "./DesktopSetupGate";

afterEach(() => {
  cleanup();
  setDesktopBridgeForTests(null);
});

function fakeBridge(overrides: Partial<DesktopBridge> = {}): DesktopBridge {
  return {
    getDatabaseConfig: vi.fn(async () => ({
      url: "",
      source: "none",
      configured: false,
      needsSetup: true,
    })),
    saveDatabaseConfig: vi.fn(async () => {}),
    testDatabaseConnection: vi.fn(async () => {}),
    quitApp: vi.fn(async () => {}),
    ...overrides,
  };
}

describe("DesktopDatabaseForm", () => {
  it("tests and saves via the bridge", async () => {
    const user = userEvent.setup();
    const bridge = fakeBridge();
    render(<DesktopDatabaseForm bridge={bridge} />);

    await user.type(screen.getByTestId("desktop-db-url"), "postgres://x");
    await user.click(screen.getByTestId("desktop-db-test"));
    expect(bridge.testDatabaseConnection).toHaveBeenCalledWith("postgres://x");

    await user.click(screen.getByTestId("desktop-db-save"));
    expect(bridge.saveDatabaseConfig).toHaveBeenCalledWith("postgres://x");
    expect(screen.getByTestId("desktop-db-quit")).toBeInTheDocument();
  });

  it("shows test errors", async () => {
    const user = userEvent.setup();
    const bridge = fakeBridge({
      testDatabaseConnection: vi.fn(async () => {
        throw new Error("connection refused");
      }),
    });
    render(<DesktopDatabaseForm bridge={bridge} />);
    await user.type(screen.getByTestId("desktop-db-url"), "postgres://bad");
    await user.click(screen.getByTestId("desktop-db-test"));
    expect(screen.getByTestId("desktop-db-error")).toHaveTextContent(
      "connection refused",
    );
  });
});

describe("DesktopSetupGate", () => {
  it("renders children when not desktop", () => {
    render(
      <DesktopSetupGate>
        <div data-testid="app-child">app</div>
      </DesktopSetupGate>,
    );
    expect(screen.getByTestId("app-child")).toBeInTheDocument();
  });

  it("shows setup when desktop needs configuration", async () => {
    setDesktopBridgeForTests(fakeBridge());
    render(
      <DesktopSetupGate>
        <div data-testid="app-child">app</div>
      </DesktopSetupGate>,
    );
    expect(await screen.findByTestId("desktop-setup-page")).toBeInTheDocument();
    expect(screen.queryByTestId("app-child")).not.toBeInTheDocument();
  });

  it("shows app when desktop is configured", async () => {
    setDesktopBridgeForTests(
      fakeBridge({
        getDatabaseConfig: vi.fn(async () => ({
          url: "postgres://ok",
          source: "file",
          configured: true,
          needsSetup: false,
        })),
      }),
    );
    render(
      <DesktopSetupGate>
        <div data-testid="app-child">app</div>
      </DesktopSetupGate>,
    );
    expect(await screen.findByTestId("app-child")).toBeInTheDocument();
  });
});
