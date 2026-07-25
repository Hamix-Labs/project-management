import { Modal } from "@/shared/Modal";
import type { TokenUsageProjection } from "@/types";
import {
  formatTokenCount,
} from "../../task-display/formatTokenCount";

type Props = {
  tokenUsage: TokenUsageProjection;
  onClose: () => void;
};

function BreakdownRow({
  label,
  value,
}: {
  label: string;
  value: number;
}) {
  const formatted = formatTokenCount(value);
  return (
    <div className="task-token-usage-breakdown-row">
      <span className="task-token-usage-breakdown-label">{label}</span>
      <span
        className="task-token-usage-breakdown-value"
        aria-label={formatted.ariaLabel}
      >
        {formatted.label}
      </span>
    </div>
  );
}

export function TokenUsageBreakdownModal({ tokenUsage, onClose }: Props) {
  return (
    <Modal onClose={onClose} labelledBy="task-token-usage-title">
      <section className="panel modal-sheet task-token-usage-modal">
        <h2 id="task-token-usage-title" className="task-token-usage-title">
          Token usage
        </h2>
        <div className="task-token-usage-breakdown">
          <BreakdownRow
            label="Execute agent"
            value={tokenUsage.execute_consumed_tokens}
          />
          <BreakdownRow
            label="Verify phase"
            value={tokenUsage.verify_consumed_tokens}
          />
          <BreakdownRow label="Total" value={tokenUsage.consumed_tokens} />
        </div>
        <div className="task-token-usage-footer">
          <button type="button" className="secondary" onClick={onClose}>
            Close
          </button>
        </div>
      </section>
    </Modal>
  );
}
