import type { HTMLAttributes } from "react";
import type { Status } from "@/types";
import { STATUS_META } from "@/lib/taskStatusDisplay";
import { StatusChip } from "./StatusChip";

type Props = HTMLAttributes<HTMLSpanElement> & {
  status: Status;
  /** Overrides the default status label (e.g. Creating PR during open-pr runs). */
  label?: string;
};

export function StatusBadge({ status, label, className, ...rest }: Props) {
  const meta = STATUS_META[status];
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
