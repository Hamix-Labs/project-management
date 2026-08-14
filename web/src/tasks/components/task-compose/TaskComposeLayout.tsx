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
  /** Right-hand handoff rail (Destination / Priority / Agent / Tags). */
  rightRail?: ReactNode;
  /** Fixed bottom action bar (readiness + Cancel / Save / Create). */
  stickyFooter?: ReactNode;
  /** Draft-assist thread column. Shown beside the handoff rail when both are set. */
  assist?: ReactNode;
};

function BackArrowIcon() {
  return (
    <svg
      className="task-compose-page__back-icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
    >
      <path d="M19 12H5" />
      <path d="m12 19-7-7 7-7" />
    </svg>
  );
}

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
  rightRail,
  stickyFooter,
  assist,
}: Props) {
  const hasRail = rightRail != null;
  const hasAssist = assist != null;

  return (
    <section
      className={
        hasRail
          ? "task-compose-page task-compose-page--v2"
          : "task-compose-page"
      }
      data-has-assist={hasAssist ? "true" : "false"}
      data-has-rail={hasRail ? "true" : "false"}
    >
      <header className="task-compose-page__topbar">
        <div className="task-compose-page__heading">
          <Link
            className="task-compose-page__back"
            to={backTo}
            aria-label={backLabel}
          >
            {hasRail ? <BackArrowIcon /> : `← ${backLabel}`}
          </Link>
          <div className="task-compose-page__title-block">
            <h1
              className="task-compose-page__title"
              id="task-compose-page-title"
            >
              {title}
            </h1>
            {subtitle ? (
              <p className="task-compose-page__subtitle">{subtitle}</p>
            ) : null}
          </div>
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
          {footer != null && !stickyFooter ? (
            <footer className="task-compose-page__footer">{footer}</footer>
          ) : null}
        </div>
        {hasAssist ? (
          <aside
            className="task-compose-page__assist"
            aria-label="Draft assist"
            data-empty="false"
          >
            {assist}
          </aside>
        ) : !hasRail ? (
          <aside
            className="task-compose-page__assist"
            aria-label="Draft assist"
            data-empty="true"
          />
        ) : null}
        {hasRail ? (
          <aside className="task-compose-page__rail" aria-label="Handoff">
            {rightRail}
          </aside>
        ) : null}
      </div>

      {stickyFooter ? (
        <div className="task-compose-page__sticky-footer">
          <div className="task-compose-page__sticky-footer-inner">
            {stickyFooter}
          </div>
        </div>
      ) : null}
    </section>
  );
}
