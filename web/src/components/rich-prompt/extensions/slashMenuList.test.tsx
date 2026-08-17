import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { SlashMenuList } from "./slashMenuList";
import { SLASH_ITEMS } from "./slashMenu";

describe("SlashMenuList", () => {
  it("shows the header placeholder when the query is empty", () => {
    render(
      <SlashMenuList items={SLASH_ITEMS} command={vi.fn()} query="" />,
    );
    expect(screen.getByText(/insert a block/i)).toBeInTheDocument();
  });

  it("uses a product-command header when emptyQueryHint is overridden", () => {
    render(
      <SlashMenuList
        items={SLASH_ITEMS.filter((i) => i.kind === "command")}
        command={vi.fn()}
        query=""
        emptyQueryHint="Mention a file or Ask AI"
      />,
    );
    expect(screen.getByText("Mention a file or Ask AI")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /mention a file/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /ask ai/i })).toBeInTheDocument();
    expect(screen.queryByText("Heading 2")).not.toBeInTheDocument();
    expect(screen.queryByText("Heading 3")).not.toBeInTheDocument();
    expect(screen.queryByText("Bulleted list")).not.toBeInTheDocument();
    expect(screen.queryByText("Numbered list")).not.toBeInTheDocument();
    expect(screen.queryByText("Quote")).not.toBeInTheDocument();
  });

  it("renders the live query in the header", () => {
    render(
      <SlashMenuList items={SLASH_ITEMS} command={vi.fn()} query="head" />,
    );
    expect(screen.getByText("/head")).toBeInTheDocument();
  });

  it("marks the selected row via aria-selected", () => {
    render(
      <SlashMenuList
        items={SLASH_ITEMS.slice(0, 3)}
        command={vi.fn()}
        selectedIndex={1}
      />,
    );
    const options = screen.getAllByRole("option");
    expect(options[0]).toHaveAttribute("aria-selected", "false");
    expect(options[1]).toHaveAttribute("aria-selected", "true");
    expect(options[1]).toHaveClass("mention-option--active");
  });

  it("invokes command when a row is clicked", async () => {
    const user = userEvent.setup();
    const command = vi.fn();
    render(
      <SlashMenuList
        items={SLASH_ITEMS.filter((i) => i.id === "ai")}
        command={command}
      />,
    );
    await user.click(screen.getByRole("button", { name: /ask ai/i }));
    expect(command).toHaveBeenCalledWith(
      expect.objectContaining({ id: "ai" }),
    );
  });

  it("shows an empty state when nothing matches", () => {
    render(
      <SlashMenuList items={[]} command={vi.fn()} query="zzz" />,
    );
    expect(screen.getByText(/no commands match .zzz./i)).toBeInTheDocument();
  });
});
