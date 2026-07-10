/**
 * Secondary line under the task title: a one-line prompt preview.
 */
export function taskListRowSubtitle(input: {
  promptPreview: string;
}): string | undefined {
  const { promptPreview } = input;
  const pv = promptPreview.replace(/\s+/g, " ").trim();
  const tail = pv.length > 80 ? `${pv.slice(0, 77)}…` : pv;

  if (tail) {
    return tail;
  }
  return undefined;
}
