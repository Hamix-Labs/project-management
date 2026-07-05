import { useMemo } from "react";
import { CustomSelect, type CustomSelectOption } from "@/components/custom-select";
import {
  TEMPLATE_CATEGORY_LABELS,
  isTemplateCategoryLabel,
} from "@/tasks/templates/templateCategories";

type TaskCreateModalTemplateCategoryFieldProps = {
  idsPrefix: string;
  tagsCsv: string;
  disabled: boolean;
  onTagsCsvChange: (value: string) => void;
};

function categoryFromTagsCsv(tagsCsv: string): string {
  const first = tagsCsv
    .split(/[,;\n]+/)
    .map((tag) => tag.trim())
    .filter(Boolean)[0];
  if (first && isTemplateCategoryLabel(first)) return first;
  return "";
}

export function TaskCreateModalTemplateCategoryField({
  idsPrefix,
  tagsCsv,
  disabled,
  onTagsCsvChange,
}: TaskCreateModalTemplateCategoryFieldProps) {
  const selected = categoryFromTagsCsv(tagsCsv);
  const fieldId = `${idsPrefix}-category`;

  const options = useMemo(
    (): CustomSelectOption[] => [
      { value: "", label: "Select a category…" },
      ...TEMPLATE_CATEGORY_LABELS.map((label) => ({ value: label, label })),
    ],
    [],
  );

  return (
    <CustomSelect
      id={fieldId}
      label="Category"
      value={selected}
      options={options}
      listboxName="Template category"
      placeholder="Select a category…"
      disabled={disabled}
      onChange={onTagsCsvChange}
    />
  );
}
