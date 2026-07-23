import { useState } from "react";
import { useTaskTokenUsage } from "../../hooks/useTaskTokenUsage";
import { formatTokenCount } from "../../task-display/formatTokenCount";
import { TokenUsageBreakdownModal } from "./TokenUsageBreakdownModal";

type Props = {
  taskId: string;
};

export function TokenUsageChip({ taskId }: Props) {
  const [open, setOpen] = useState(false);
  const usageQuery = useTaskTokenUsage(taskId);
  const tokenUsage = usageQuery.data?.token_usage;
  const known = tokenUsage?.known === true;
  const formatted =
    known && tokenUsage ? formatTokenCount(tokenUsage.consumed_tokens) : null;

  return (
    <>
      <button
        type="button"
        className="task-token-usage-chip"
        data-known={known ? "true" : "false"}
        data-testid="task-token-usage-chip"
        aria-label={
          known && formatted
            ? `Token usage: ${formatted.ariaLabel}. Open breakdown.`
            : "Token usage unknown. Open breakdown."
        }
        onClick={() => setOpen(true)}
      >
        {known && formatted ? (
          <>
            <span className="task-token-usage-chip-label">Tokens</span>
            <span aria-hidden="true">{formatted.label}</span>
          </>
        ) : (
          <span className="muted">Tokens —</span>
        )}
      </button>
      {open && usageQuery.data ? (
        <TokenUsageBreakdownModal
          tokenUsage={usageQuery.data.token_usage}
          onClose={() => setOpen(false)}
        />
      ) : null}
    </>
  );
}
