import type { ReactNode } from "react";
import type { FieldRequirement } from "@/shared/FieldLabel";
import { FieldRequirementBadge } from "@/shared/FieldLabel";

type SectionVariant =
  | "essentials"
  | "prompt"
  | "criteria"
  | "context"
  | "execution";

type Props = {
  title: string;
  lede?: string;
  children: ReactNode;
  variant: SectionVariant;
  action?: ReactNode;
  requirement?: FieldRequirement;
};

export function TaskCreateModalSection({
  title,
  lede,
  children,
  variant,
  action,
  requirement,
}: Props) {
  const headId = `task-create-modal-section-${variant}`;

  return (
    <section
      className={`task-create-modal-section task-create-modal-section--${variant}`}
      aria-labelledby={headId}
    >
      <header className="task-create-modal-section__head">
        <div className="task-create-modal-section__head-text">
          <div className="task-create-modal-section__title-row">
            <h3 id={headId} className="task-create-modal-section__title">
              {title}
            </h3>
            {requirement ? (
              <FieldRequirementBadge requirement={requirement} />
            ) : null}
          </div>
          {lede ? (
            <p className="task-create-modal-section__lede">{lede}</p>
          ) : null}
        </div>
        {action ? (
          <div className="task-create-modal-section__action">{action}</div>
        ) : null}
      </header>
      <div className="task-create-modal-section__body">{children}</div>
    </section>
  );
}
