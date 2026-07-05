export const TEMPLATE_CATEGORY_LABELS = [
  "Refactor",
  "Testing",
  "Docs",
  "Infra",
  "Review",
] as const;

export type TemplateCategoryLabel = (typeof TEMPLATE_CATEGORY_LABELS)[number];

export function isTemplateCategoryLabel(value: string): value is TemplateCategoryLabel {
  return (TEMPLATE_CATEGORY_LABELS as readonly string[]).includes(value);
}
