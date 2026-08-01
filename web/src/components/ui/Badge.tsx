import type { HTMLAttributes, ReactNode } from "react";
import type { Status } from "@/types";
import { cn } from "@/utils/cn";

export type BadgeProps = HTMLAttributes<HTMLSpanElement> & {
  status?: Status;
  /** Override when not mapping from task status */
  tone?: "neutral" | "success" | "warning" | "brand";
  children: ReactNode;
};

const STATUS_CLASS: Record<Status, string> = {
  ready: "ui-badge--status-ready",
  running: "ui-badge--status-running",
  blocked: "ui-badge--status-on_hold",
  review: "ui-badge--status-running",
  pr_ready: "ui-badge--status-pr_ready",
  done: "ui-badge--status-done",
  failed: "ui-badge--status-failed",
  on_hold: "ui-badge--status-on_hold",
  closed: "ui-badge--status-on_hold",
};

export function Badge({ status, tone, className, children, ...rest }: BadgeProps) {
  const statusClass = status ? STATUS_CLASS[status] : undefined;
  const toneClass =
    !status && tone
      ? {
          neutral: "ui-badge--status-on_hold",
          success: "ui-badge--status-done",
          warning: "ui-badge--status-needs_user",
          brand: "ui-badge--status-running",
        }[tone]
      : undefined;

  return (
    <span className={cn("ui-badge", statusClass, toneClass, className)} {...rest}>
      {children}
    </span>
  );
}
