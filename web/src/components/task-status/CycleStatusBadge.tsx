import type { HTMLAttributes } from "react";
import type { CycleStatus } from "@/types/cycle";
import { CYCLE_STATUS_META } from "@/lib/cycleStatusDisplay";
import { StatusChip } from "./StatusChip";

type Props = HTMLAttributes<HTMLSpanElement> & {
  status: CycleStatus;
};

export function CycleStatusBadge({ status, className, ...rest }: Props) {
  const meta = CYCLE_STATUS_META[status];
  return (
    <StatusChip
      tone={meta.tone}
      label={meta.label}
      pulse={meta.pulse}
      className={className}
      {...rest}
    />
  );
}
