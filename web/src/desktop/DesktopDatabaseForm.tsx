import { useState } from "react";
import type { DesktopBridge } from "./bridge";
import "./desktopSetup.css";

type Props = {
  bridge: DesktopBridge;
  /** Pre-fill when editing from Settings. */
  initialUrl?: string;
  /** When true, copy emphasizes restart after save. */
  editing?: boolean;
};

function formatBridgeError(err: unknown): string {
  if (err instanceof Error) return err.message;
  if (typeof err === "string") return err;
  return "Something went wrong talking to the desktop host.";
}

export function DesktopDatabaseForm({
  bridge,
  initialUrl = "",
  editing = false,
}: Props) {
  const [url, setUrl] = useState(initialUrl);
  const [busy, setBusy] = useState<"test" | "save" | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  async function onTest() {
    setBusy("test");
    setError(null);
    setMessage(null);
    try {
      await bridge.testDatabaseConnection(url.trim());
      setMessage("Connection succeeded.");
    } catch (err) {
      setError(formatBridgeError(err));
    } finally {
      setBusy(null);
    }
  }

  async function onSave() {
    setBusy("save");
    setError(null);
    setMessage(null);
    try {
      await bridge.saveDatabaseConfig(url.trim());
      setSaved(true);
      setMessage(
        editing
          ? "Saved. Restart Hamix for the new database URL to take effect."
          : "Saved. Restart Hamix to continue.",
      );
    } catch (err) {
      setError(formatBridgeError(err));
    } finally {
      setBusy(null);
    }
  }

  async function onQuit() {
    await bridge.quitApp();
  }

  return (
    <div className="desktop-db-form">
      <label className="desktop-db-field">
        <span className="desktop-db-label">Postgres connection URL</span>
        <input
          className="desktop-db-input"
          type="text"
          name="database_url"
          autoComplete="off"
          spellCheck={false}
          placeholder="postgres://user:pass@localhost:5432/hamix"
          value={url}
          onChange={(e) => {
            setUrl(e.target.value);
            setSaved(false);
          }}
          data-testid="desktop-db-url"
        />
      </label>
      <p className="desktop-db-help">
        Hamix does not install Postgres. Use a local instance, Docker, or a
        hosted database, then paste the connection string here.
      </p>
      {error ? (
        <p className="desktop-db-error" role="alert" data-testid="desktop-db-error">
          {error}
        </p>
      ) : null}
      {message ? (
        <p className="desktop-db-message" role="status" data-testid="desktop-db-message">
          {message}
        </p>
      ) : null}
      <div className="desktop-db-actions">
        <button
          type="button"
          className="desktop-db-btn desktop-db-btn-secondary"
          disabled={busy !== null || !url.trim()}
          onClick={() => void onTest()}
          data-testid="desktop-db-test"
        >
          {busy === "test" ? "Testing…" : "Test connection"}
        </button>
        <button
          type="button"
          className="desktop-db-btn desktop-db-btn-primary"
          disabled={busy !== null || !url.trim()}
          onClick={() => void onSave()}
          data-testid="desktop-db-save"
        >
          {busy === "save" ? "Saving…" : "Save"}
        </button>
        {saved ? (
          <button
            type="button"
            className="desktop-db-btn desktop-db-btn-primary"
            onClick={() => void onQuit()}
            data-testid="desktop-db-quit"
          >
            Quit Hamix
          </button>
        ) : null}
      </div>
    </div>
  );
}
