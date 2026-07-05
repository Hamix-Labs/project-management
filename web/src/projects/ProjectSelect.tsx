import { useMemo } from "react";
import { CustomSelect, type CustomSelectOption } from "@/components/custom-select";
import type { Project } from "@/types";
import { ProjectsStackIcon } from "./ProjectsStackIcon";

type Props = {
  id: string;
  value: string;
  projects: Project[];
  loading?: boolean;
  disabled?: boolean;
  onChange: (projectId: string) => void;
};

export function ProjectSelect({
  id,
  value,
  projects,
  loading = false,
  disabled = false,
  onChange,
}: Props) {
  const options = useMemo<CustomSelectOption[]>(() => {
    const active = projects.filter((p) => p.status === "active");
    return active.map((project) => ({
      value: project.id,
      label: project.is_default ? "Default" : project.name,
    }));
  }, [projects]);

  return (
    <CustomSelect
      id={id}
      label="Project"
      value={value}
      options={options}
      onChange={onChange}
      disabled={disabled || loading}
      requirement="required"
      leadingIcon={<ProjectsStackIcon className="project-select__icon" />}
    />
  );
}
