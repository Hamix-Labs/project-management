import type { ChecklistVerifyCommandInput, Priority } from "@/types";

/** One done criterion pre-filled when a scenario is applied to the create form. */
export type TestScenarioCriterion = {
  text: string;
  verify_commands?: ChecklistVerifyCommandInput[];
};

function shellCommand(command: string): ChecklistVerifyCommandInput {
  return { command };
}

/**
 * One ready-to-run sample task. Picking it fills the compose form so the
 * operator can dispatch a real agent run with zero typing.
 */
export type TestScenario = {
  id: string;
  /** Menu label in the sample-task picker. */
  name: string;
  /** Short title shown in the picker; becomes the task `title`. */
  title: string;
  /** One-line description shown beneath the name in the picker. */
  description: string;
  /**
   * Plain-text body inserted into `initial_prompt`. The picker wraps this
   * in `<p>` blocks via `plainTextToInitialHtml` so the rich editor renders
   * paragraphs instead of one mega-line.
   */
  prompt: string;
  priority: Priority;
  /** Done criteria written into the form on apply (text + optional verify commands). */
  criteria: TestScenarioCriterion[];
  /** Comma-separated tags applied via `setNewTagsCsv`. Never overwrites destination. */
  tags?: string;
};

export const TEST_SCENARIOS: TestScenario[] = [
  {
    id: "observability",
    name: "Observability pass",
    title: "Add structured observability to a critical flow",
    description:
      "Structured logging, counters, and correlation IDs on a critical flow.",
    prompt: [
      "Pick a critical end-to-end flow in the codebase — a request lifecycle, a CLI run, a worker job. Justify the pick in one sentence.",
      "Add structured logging at each meaningful state transition: start, every key decision point, terminal states (success / failure).",
      "Add a counter that records success and failure counts for the flow. Wire it through whatever metrics surface the project already uses.",
      "Propagate a request / correlation ID through every log line in the flow, generating one at the entry point if no upstream ID is present.",
      "Document the new instrumentation in the project's observability doc (or create OBSERVABILITY.md if none exists).",
    ].join("\n\n"),
    priority: "high",
    tags: "observability, logging",
    criteria: [
      { text: "The chosen flow is named with a one-line justification." },
      {
        text: "Every meaningful state transition logs structured fields including the correlation ID.",
      },
      {
        text: "A success / failure counter (or equivalent) exists and is documented.",
      },
      {
        text: "Operator-facing observability documentation was updated or created.",
        verify_commands: [shellCommand("npm run docs:check")],
      },
    ],
  },
  {
    id: "flaky-test",
    name: "Flaky test hunt",
    title: "Stabilize the flaky integration test suite",
    description: "Reproduce, isolate, and stabilize an intermittently failing test.",
    prompt: [
      "Identify the most frequently flaking test in CI over the last 30 runs. Name it and quote its failure rate.",
      "Reproduce the flake locally by running the isolated test in a loop until it fails at least twice.",
      "Determine the root cause — shared state, timing, ordering, or an external dependency — and fix it at the source rather than adding retries.",
      "Add a regression guard so the specific failure mode cannot silently return.",
    ].join("\n\n"),
    priority: "medium",
    tags: "testing, ci",
    criteria: [
      { text: "The flakiest test is identified with its measured failure rate." },
      { text: "The flake is reproduced deterministically before any fix." },
      { text: "Root cause is fixed at the source, not masked with retries." },
      { text: "A regression guard covers the original failure mode." },
    ],
  },
  {
    id: "dep-upgrade",
    name: "Dependency upgrade",
    title: "Upgrade the primary framework to the latest major",
    description: "Bump a major dependency and resolve the resulting breakage.",
    prompt: [
      "Upgrade the primary framework dependency to its latest stable major version.",
      "Work through every breaking change in the migration guide, updating call sites as you go.",
      "Run the full test suite and the build; resolve all failures introduced by the upgrade.",
      "Summarize notable behavioral changes engineers should know about in the PR description.",
    ].join("\n\n"),
    priority: "low",
    tags: "maintenance, deps",
    criteria: [
      {
        text: "The dependency is on the latest stable major and lockfile is updated.",
      },
      {
        text: "The build and full test suite pass.",
        verify_commands: [shellCommand("npm run build && npm test")],
      },
      { text: "Breaking changes are addressed at every affected call site." },
    ],
  },
];

export function findTestScenarioById(id: string): TestScenario | undefined {
  return TEST_SCENARIOS.find((scenario) => scenario.id === id);
}
