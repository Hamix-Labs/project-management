import type { HTMLAttributes } from "react";
import type { Status } from "@/types";
import { STATUS_META } from "@/lib/taskStatusDisplay";
import { StatusChip } from "./StatusChip";

type Props = HTMLAttributes<HTMLSpanElement> & {
  status: Status;
};

export function StatusBadge({ status, className, ...rest }: Props) {
  const meta = STATUS_META[status];
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
