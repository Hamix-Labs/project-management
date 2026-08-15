import { useCallback, useEffect, useState } from "react";
import { readyProbe } from "@/api/draftAssist";
import type { DraftAssistReady } from "@/types/draftAssist";

const TASKAPI_DOWN_COPY =
  "Assistant unavailable — taskapi is not running.";

function copyForReady(ready: DraftAssistReady | null): string | null {
  if (!ready || ready.ready) return null;
  switch (ready.reason) {
    case "missing_key":
      return "Assistant unavailable — set CURSOR_API_KEY and restart taskapi.";
    case "sidecar_down":
      return "Assistant unavailable — draft-assist sidecar is down.";
    case "no_runner":
      return "Assistant unavailable — hamix-draft-agent launcher is missing.";
    default:
      return "Assistant unavailable — try again in a moment.";
  }
}

type Props = {
  /** Called when the operator clicks Retry after a probe. */
  onRetry?: () => void;
};

/**
 * Surfaces GET /draft-assist/ready failures on the compose page before the
 * operator sends a message. Hidden when ready or while the probe is in flight.
 */
export function DraftAssistNotReadyBanner({ onRetry }: Props) {
  const [ready, setReady] = useState<DraftAssistReady | null>(null);
  const [unreachable, setUnreachable] = useState(false);
  const [probing, setProbing] = useState(true);

  const probe = useCallback(async () => {
    setProbing(true);
    try {
      setReady(await readyProbe());
      setUnreachable(false);
    } catch {
      setReady(null);
      setUnreachable(true);
    } finally {
      setProbing(false);
    }
  }, []);

  useEffect(() => {
    void probe();
  }, [probe]);

  const copy = unreachable ? TASKAPI_DOWN_COPY : copyForReady(ready);
  if (probing || copy == null) return null;

  return (
    <div
      className="task-compose-page__assist-banner"
      role="status"
      aria-live="polite"
    >
      <p className="task-compose-page__assist-banner-copy">{copy}</p>
      <button
        type="button"
        className="btn btn-secondary"
        onClick={() => {
          void probe();
          onRetry?.();
        }}
      >
        Retry
      </button>
    </div>
  );
}
