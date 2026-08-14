import type { ReactNode } from "react";
import { Link } from "react-router-dom";

type Props = {
  title: string;
  subtitle?: string | null;
  backTo: string;
  backLabel?: string;
  topActions?: ReactNode;
  children: ReactNode;
  errors?: ReactNode;
  footer?: ReactNode;
  /** Reserved for Plan 4 assist thread; empty in Wave A. */
  assist?: ReactNode;
};

/** Page chrome for task/template compose (ADR-0100). */
export function TaskComposeLayout({
  title,
  subtitle,
  backTo,
  backLabel = "Back",
  topActions,
  children,
  errors,
  footer,
  assist,
}: Props) {
  const hasAssist = assist != null;

  return (
    <section
      className="task-compose-page"
      data-has-assist={hasAssist ? "true" : "false"}
    >
      <header className="task-compose-page__topbar">
        <Link className="task-compose-page__back" to={backTo}>
          ← {backLabel}
        </Link>
        <div className="task-compose-page__title-block">
          <h1 className="task-compose-page__title" id="task-compose-page-title">
            {title}
          </h1>
          {subtitle ? (
            <p className="task-compose-page__subtitle">{subtitle}</p>
          ) : null}
        </div>
        {topActions ? (
          <div className="task-compose-page__top-actions">{topActions}</div>
        ) : null}
      </header>

      <div className="task-compose-page__grid">
        <div className="task-compose-page__form-column">
          {children}
          {errors ? (
            <div className="task-compose-page__errors">{errors}</div>
          ) : null}
          {footer != null ? (
            <footer className="task-compose-page__footer">{footer}</footer>
          ) : null}
        </div>
        <aside
          className="task-compose-page__assist"
          aria-label="Draft assist"
          data-empty={hasAssist ? "false" : "true"}
        >
          {assist}
        </aside>
      </div>
    </section>
  );
}
