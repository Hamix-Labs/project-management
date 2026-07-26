import type { VerificationSnapshot } from "../../task-events/parseVerificationSnapshot";

/**
 * Operator-facing failure copy for a `phase_failed` activity row.
 * Mirrors Go `FailureSurfaceMessage` precedence, plus verification
 * counts and a phase-generic fallback so the Timeline never shows a
 * blank detail line.
 */
export function phaseFailureDetail(
  data: Record<string, unknown>,
  phaseLabel: string,
  snapshot: VerificationSnapshot | null,
): string {
  const failureSummary = trimStr(data.failure_summary);
  if (failureSummary) return failureSummary;

  const details =
    data.details && typeof data.details === "object" && !Array.isArray(data.details)
      ? (data.details as Record<string, unknown>)
      : null;

  const standardized = trimStr(details?.standardized_message);
  if (standardized) return standardized;

  const summary = trimStr(data.summary);
  if (summary) return summary;

  const failureKind = trimStr(details?.failure_kind);
  if (failureKind) {
    const humanized = humanizeFailureKind(failureKind);
    if (humanized) return humanized;
    return failureKind;
  }

  if (snapshot && snapshot.failedCount > 0) {
    return `${snapshot.failedCount} criteria failed`;
  }

  return `${phaseLabel} phase failed.`;
}

/** First failed criterion id for a short title highlight, if any. */
export function firstFailedCriterionHighlight(
  snapshot: VerificationSnapshot | null,
): string {
  if (!snapshot) return "";
  const failed = snapshot.criteria.find((c) => !c.verified);
  return failed?.criterionId ?? "";
}

function trimStr(v: unknown): string {
  return typeof v === "string" ? v.trim() : "";
}

function humanizeFailureKind(kind: string): string {
  switch (kind) {
    case "cursor_usage_limit":
      return "Cursor usage limit reached";
    case "cursor_missing_session_id":
      return "Cursor chat session id missing";
    case "cursor_resume_session":
      return "Cursor could not resume the prior chat";
    default:
      return "";
  }
}
