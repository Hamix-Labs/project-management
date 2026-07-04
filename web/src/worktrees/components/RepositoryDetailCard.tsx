import type { ReactNode } from "react";

type Props = {
  headingId?: string;
  children: ReactNode;
  footer?: ReactNode;
  className?: string;
};

export function RepositoryDetailCard({ headingId, children, footer, className = "" }: Props) {
  return (
    <section
      className={`repository-detail-card worktrees-page ${className}`.trim()}
      aria-labelledby={headingId}
    >
      {children}
      {footer}
    </section>
  );
}
