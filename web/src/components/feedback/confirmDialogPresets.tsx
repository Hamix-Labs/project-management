import type { ReactNode } from "react";

export type DeleteEntityConfirmPreset = {
  noun: string;
  name: string;
  titleId: string;
  descriptionId: string;
  footnote?: string;
  confirmLabel?: string;
};

export type DeleteManyConfirmPreset = {
  noun: string;
  count: number;
  titleId: string;
  descriptionId: string;
  footnote?: string;
  confirmLabel?: string;
};

export function deleteEntityConfirmProps(preset: DeleteEntityConfirmPreset) {
  return {
    title: `Delete this ${preset.noun}?`,
    description: (
      <>
        <strong>{preset.name}</strong> will be permanently deleted.
      </>
    ) as ReactNode,
    footnote: preset.footnote ?? "This action cannot be undone.",
    confirmLabel: preset.confirmLabel ?? "Delete",
    confirmVariant: "danger" as const,
    titleId: preset.titleId,
    descriptionId: preset.descriptionId,
  };
}

export function deleteManyConfirmProps(preset: DeleteManyConfirmPreset) {
  const noun = preset.count === 1 ? preset.noun : `${preset.noun}s`;
  return {
    title: `Delete ${preset.count} ${noun}?`,
    description: (
      <>
        {preset.count} {noun} will be permanently deleted.
      </>
    ) as ReactNode,
    footnote: preset.footnote ?? "This action cannot be undone.",
    confirmLabel: preset.confirmLabel ?? "Delete",
    confirmVariant: "danger" as const,
    titleId: preset.titleId,
    descriptionId: preset.descriptionId,
  };
}
