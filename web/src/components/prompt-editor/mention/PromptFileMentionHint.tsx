import {
  describeMentionSearchStatus,
  type MentionSearchStatus,
} from "./promptFileMentionStatus";

type Props = {
  status: MentionSearchStatus;
};

/** Status copy under the prompt editor for `@` file mentions. */
export function PromptFileMentionHint({ status }: Props) {
  const hint = describeMentionSearchStatus(status);
  if (!hint) return null;

  const className =
    hint.tone === "error"
      ? "mention-repo-hint"
      : "mention-repo-hint mention-repo-hint--pending";

  return (
    <p className={className} role="status" aria-live="polite">
      {hint.message}
      {hint.action ? (
        <>
          {" "}
          <a href={hint.action.href} target="_blank" rel="noopener noreferrer">
            {hint.action.label}
          </a>{" "}
          {hint.trailing}
        </>
      ) : null}
    </p>
  );
}
