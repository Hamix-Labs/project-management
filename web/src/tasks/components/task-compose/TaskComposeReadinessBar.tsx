import type { ChecklistItemDraft } from "@/types";
import { nonEmptyChecklistCount } from "@/tasks/task-compose/checklistRequirement";

export type ReadinessCheck = {
  label: string;
  ok: boolean;
};

export function buildComposeReadinessChecks(input: {
  title: string;
  brief: string;
  repositoryId: string;
  checklistItems: ChecklistItemDraft[];
}): ReadinessCheck[] {
  return [
    { label: "Title", ok: input.title.trim().length > 0 },
    { label: "Brief", ok: input.brief.trim().length > 0 },
    { label: "Repository", ok: input.repositoryId.trim().length > 0 },
    {
      label: "Done criteria",
      ok: nonEmptyChecklistCount(input.checklistItems) >= 1,
    },
  ];
}

type Props = {
  title: string;
  brief: string;
  repositoryId: string;
  checklistItems: ChecklistItemDraft[];
};

/**
 * Visual readiness mirror for the sticky footer.
 * Submit enablement stays owned by TaskCreateModalFooterActions —
 * this bar must never gate mutations.
 */
export function TaskComposeReadinessBar({
  title,
  brief,
  repositoryId,
  checklistItems,
}: Props) {
  const checks = buildComposeReadinessChecks({
    title,
    brief,
    repositoryId,
    checklistItems,
  });
  const readyCount = checks.filter((c) => c.ok).length;
  const allReady = readyCount === checks.length;

  return (
    <div
      className="task-compose-readiness"
      data-testid="task-compose-readiness"
      aria-live="polite"
    >
      <span
        className="task-compose-readiness__dot"
        data-ready={allReady ? "true" : "false"}
        aria-hidden="true"
      >
        {allReady ? <CheckIcon /> : <DashedCircleIcon />}
      </span>
      <span className="task-compose-readiness__label">
        {allReady ? (
          <strong>Ready to hand off</strong>
        ) : (
          <>
            <strong>
              {readyCount}/{checks.length}
            </strong>{" "}
            essentials ready
          </>
        )}
      </span>
      <span className="task-compose-readiness__chips">
        {checks.map((c) => (
          <span
            key={c.label}
            title={c.label}
            className="task-compose-readiness__chip"
            data-ok={c.ok ? "true" : "false"}
          >
            <span className="task-compose-readiness__chip-dot" />
            {c.label}
          </span>
        ))}
      </span>
    </div>
  );
}

function CheckIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

function DashedCircleIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <circle cx="12" cy="12" r="9" strokeDasharray="3 3" />
    </svg>
  );
}
