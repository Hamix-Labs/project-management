import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { CustomSelect, type CustomSelectOption } from "./CustomSelect";

const OPTIONS: CustomSelectOption[] = [
  { value: "ready", label: "Ready" },
  { value: "running", label: "Running" },
];

describe("CustomSelect", () => {
  it("closes the listbox when tabbing away", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <>
        <CustomSelect
          id="status"
          label="Status"
          value="ready"
          options={OPTIONS}
          onChange={onChange}
        />
        <button type="button">After field</button>
      </>,
    );

    await user.click(screen.getByRole("combobox", { name: /status/i }));
    expect(screen.getByRole("listbox", { name: /status/i })).toBeInTheDocument();

    await user.tab();
    await waitFor(() => {
      expect(
        screen.queryByRole("listbox", { name: /status/i }),
      ).not.toBeInTheDocument();
    });
  });

  it("selects the highlighted option when pressing Space in the listbox", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <CustomSelect
        id="status"
        label="Status"
        value="ready"
        options={OPTIONS}
        onChange={onChange}
      />,
    );

    const trigger = screen.getByRole("combobox", { name: /status/i });
    await user.click(trigger);
    const listbox = screen.getByRole("listbox", { name: /status/i });
    await user.click(listbox);

    await user.keyboard("{ArrowDown}");
    await user.keyboard(" ");

    expect(onChange).toHaveBeenCalledWith("running");
  });

  it("does not show the first option label when unset and no placeholder", () => {
    render(
      <CustomSelect
        id="status"
        label="Status"
        value=""
        options={OPTIONS}
        onChange={() => {}}
      />,
    );

    const trigger = screen.getByRole("combobox", { name: /status/i });
    expect(trigger).not.toHaveTextContent("Ready");
    expect(trigger).not.toHaveTextContent("Running");
  });

  it("shows placeholder on trigger when unset and options exist", () => {
    render(
      <CustomSelect
        id="status"
        label="Status"
        value=""
        options={OPTIONS}
        placeholder="Select a status"
        onChange={() => {}}
      />,
    );

    expect(screen.getByRole("combobox", { name: /status/i })).toHaveTextContent(
      "Select a status",
    );
  });

  it("shows placeholder on trigger when unset and options are empty", () => {
    render(
      <CustomSelect
        id="status"
        label="Status"
        value=""
        options={[]}
        placeholder="No statuses available"
        onChange={() => {}}
      />,
    );

    expect(screen.getByRole("combobox", { name: /status/i })).toHaveTextContent(
      "No statuses available",
    );
  });

  it("shows placeholder once in empty listbox when opened with zero options", async () => {
    const user = userEvent.setup();

    render(
      <CustomSelect
        id="status"
        label="Status"
        value=""
        options={[]}
        placeholder="No statuses available"
        onChange={() => {}}
      />,
    );

    await user.click(screen.getByRole("combobox", { name: /status/i }));
    const listbox = screen.getByRole("listbox", { name: /status/i });
    expect(listbox).toHaveTextContent("No statuses available");
    expect(screen.queryAllByRole("option")).toHaveLength(0);
  });

  it("renders a leading icon inside the trigger when leadingIcon is set", () => {
    render(
      <CustomSelect
        id="repo"
        label="Repository"
        value="ready"
        options={OPTIONS}
        onChange={() => {}}
        leadingIcon={<span data-testid="repo-icon">icon</span>}
      />,
    );

    expect(screen.getByTestId("repo-icon")).toBeInTheDocument();
    expect(
      screen.getByRole("combobox", { name: /repository/i }).closest(
        ".field--custom-select--leading-icon",
      ),
    ).not.toBeNull();
  });

  it("exposes option title on the trigger and listbox options", async () => {
    const user = userEvent.setup();
    const titled: CustomSelectOption[] = [
      { value: "hamix", label: "hamix", title: "C:/Users/dev/Documents/hamix" },
      {
        value: "other",
        label: "Hamix-project-management",
        title: "C:/Users/dev/Documents/Hamix-project-management",
      },
    ];

    render(
      <CustomSelect
        id="repo"
        label="Repository"
        value="hamix"
        options={titled}
        onChange={() => {}}
      />,
    );

    const trigger = screen.getByRole("combobox", { name: /repository/i });
    expect(trigger).toHaveTextContent("hamix");
    expect(trigger).toHaveAttribute("title", "C:/Users/dev/Documents/hamix");

    await user.click(trigger);
    expect(screen.getByRole("option", { name: "hamix" })).toHaveAttribute(
      "title",
      "C:/Users/dev/Documents/hamix",
    );
    expect(
      screen.getByRole("option", { name: "Hamix-project-management" }),
    ).toHaveAttribute("title", "C:/Users/dev/Documents/Hamix-project-management");
  });
});
