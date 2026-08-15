import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { TaskComposeLayout } from "./TaskComposeLayout";

describe("TaskComposeLayout", () => {
  it("marks the page as v2 when a handoff rail is present", () => {
    render(
      <MemoryRouter>
        <TaskComposeLayout
          title="New task"
          backTo="/"
          rightRail={<div>Destination</div>}
        >
          <p>Form body</p>
        </TaskComposeLayout>
      </MemoryRouter>,
    );

    expect(document.querySelector(".task-compose-page")).toHaveClass(
      "task-compose-page--v2",
    );
  });

  it("portals the sticky footer to document.body so it can stay viewport-fixed", () => {
    render(
      <MemoryRouter>
        <div data-testid="compose-host">
          <TaskComposeLayout
            title="New task"
            backTo="/"
            stickyFooter={<button type="button">Create task</button>}
          >
            <p>Form body</p>
          </TaskComposeLayout>
        </div>
      </MemoryRouter>,
    );

    const footer = screen.getByTestId("task-compose-sticky-footer");
    expect(footer).toBeInTheDocument();
    expect(footer.parentElement).toBe(document.body);
    expect(
      screen.getByTestId("compose-host").querySelector(
        ".task-compose-page__sticky-footer",
      ),
    ).toBeNull();
  });
});
