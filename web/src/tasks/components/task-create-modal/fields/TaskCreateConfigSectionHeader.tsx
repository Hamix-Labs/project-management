import type { ReactNode } from "react";

type Props = {
  id: string;
  title: string;
  icon: ReactNode;
};

/** Shared icon + title row for Agent / Tags / Schedule sections. */
export function TaskCreateConfigSectionHeader({ id, title, icon }: Props) {
  return (
    <div className="task-create-config-section__head">
      <span className="task-create-config-section__icon" aria-hidden="true">
        {icon}
      </span>
      <h3 id={id} className="task-create-config-section__title">
        {title}
      </h3>
    </div>
  );
}
