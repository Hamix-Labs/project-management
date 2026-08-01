import type { HTMLAttributes } from "react";
import type { CycleStatus } from "@/types/cycle";
import { CYCLE_STATUS_META } from "@/lib/cycleStatusDisplay";
import { StatusChip } from "./StatusChip";

type Props = HTMLAttributes<HTMLSpanElement> & {
  status: CycleStatus;
  /** Overrides the default cycle status label (e.g. Creating PR). */
  label?: string;
};

export function CycleStatusBadge({ status, label, className, ...rest }: Props) {
  const meta = CYCLE_STATUS_META[status];
  return (
    <StatusChip
      tone={meta.tone}
      label={label ?? meta.label}
      pulse={meta.pulse}
      className={className}
      {...rest}
    />
  );
}
