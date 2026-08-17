import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useRef, useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { TEST_SCENARIOS, type TestScenario } from "@/tasks/test-scenarios";
import { TestScenariosPopover } from "./TestScenariosPopover";

function Harness({ onPick }: { onPick: (s: TestScenario) => void }) {
  const ref = useRef<HTMLButtonElement>(null);
  const [open, setOpen] = useState(false);
  return (
    <>
      <button
        ref={ref}
        type="button"
        data-testid="harness-trigger"
        onClick={() => setOpen((v) => !v)}
      >
        Open
      </button>
      {open ? (
        <TestScenariosPopover
          anchor={ref.current}
          onPick={(s) => {
            onPick(s);
            setOpen(false);
          }}
          onClose={() => setOpen(false)}
        />
      ) : null}
    </>
  );
}

describe("TestScenariosPopover", () => {
  it("renders every sample-task pick row", async () => {
    const user = userEvent.setup();
    render(<Harness onPick={vi.fn()} />);
    await user.click(screen.getByTestId("harness-trigger"));
    expect(screen.getByTestId("test-scenarios-popover")).toBeInTheDocument();
    for (const scenario of TEST_SCENARIOS) {
      expect(
        screen.getByTestId(`test-scenarios-pick-${scenario.id}`),
      ).toBeInTheDocument();
    }
  });

  it("invokes onPick with the chosen scenario when a row is clicked", async () => {
    const user = userEvent.setup();
    const onPick = vi.fn();
    render(<Harness onPick={onPick} />);
    await user.click(screen.getByTestId("harness-trigger"));
    const target = TEST_SCENARIOS[0]!;
    await user.click(
      screen.getByTestId(`test-scenarios-pick-${target.id}`),
    );
    expect(onPick).toHaveBeenCalledTimes(1);
    expect(onPick.mock.calls[0]?.[0]?.id).toBe(target.id);
  });

  it("Escape closes the popover without invoking onPick", async () => {
    const user = userEvent.setup();
    const onPick = vi.fn();
    render(<Harness onPick={onPick} />);
    await user.click(screen.getByTestId("harness-trigger"));
    expect(screen.getByTestId("test-scenarios-popover")).toBeInTheDocument();
    await user.keyboard("{Escape}");
    expect(
      screen.queryByTestId("test-scenarios-popover"),
    ).not.toBeInTheDocument();
    expect(onPick).not.toHaveBeenCalled();
  });
});
