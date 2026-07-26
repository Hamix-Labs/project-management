import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaskCreateTagsPillsField } from "./TaskCreateTagsPillsField";
import {
  formatTagsCsv,
  isValidTaskTag,
  parseTagsFromCsv,
} from "@/tasks/create/composePayload";

describe("isValidTaskTag / parseTagsFromCsv", () => {
  it("accepts server-shaped tags and rejects leading punctuation", () => {
    expect(isValidTaskTag("backend")).toBe(true);
    expect(isValidTaskTag("api.v2")).toBe(true);
    expect(isValidTaskTag("a".repeat(32))).toBe(true);
    expect(isValidTaskTag("a".repeat(33))).toBe(false);
    expect(isValidTaskTag(".leading")).toBe(false);
    expect(isValidTaskTag("has space")).toBe(false);
  });

  it("round-trips CSV via parse + format", () => {
    expect(formatTagsCsv(parseTagsFromCsv("Backend, API;extra"))).toBe(
      "backend, api, extra",
    );
  });
});

describe("TaskCreateTagsPillsField", () => {
  it("renders existing tags as removable pills", () => {
    render(
      <TaskCreateTagsPillsField
        tagsCsv="backend, api"
        onTagsCsvChange={vi.fn()}
      />,
    );
    expect(screen.getByText("backend")).toBeInTheDocument();
    expect(screen.getByText("api")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Remove backend" })).toBeInTheDocument();
  });

  it("commits a tag on Enter and lowercases it", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <TaskCreateTagsPillsField tagsCsv="" onTagsCsvChange={onChange} />,
    );
    const input = screen.getByLabelText(/^tags$/i);
    await user.type(input, "Backend{Enter}");
    expect(onChange).toHaveBeenCalledWith("backend");
  });

  it("shows an error for invalid tags and does not commit them", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <TaskCreateTagsPillsField tagsCsv="" onTagsCsvChange={onChange} />,
    );
    const input = screen.getByLabelText(/^tags$/i);
    await user.type(input, ".bad{Enter}");
    expect(onChange).not.toHaveBeenCalled();
    expect(screen.getByRole("alert")).toHaveTextContent(/invalid/i);
  });

  it("removes a tag via the pill button", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <TaskCreateTagsPillsField
        tagsCsv="backend, api"
        onTagsCsvChange={onChange}
      />,
    );
    await user.click(screen.getByRole("button", { name: "Remove api" }));
    expect(onChange).toHaveBeenCalledWith("backend");
  });

  it("removes the last tag on Backspace when the draft is empty", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <TaskCreateTagsPillsField
        tagsCsv="backend, api"
        onTagsCsvChange={onChange}
      />,
    );
    const input = screen.getByLabelText(/^tags$/i);
    await user.click(input);
    await user.keyboard("{Backspace}");
    expect(onChange).toHaveBeenCalledWith("backend");
  });
});
