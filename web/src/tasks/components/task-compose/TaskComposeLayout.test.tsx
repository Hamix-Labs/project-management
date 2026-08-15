import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { TaskComposeLayout } from "./TaskComposeLayout";

describe("TaskComposeLayout", () => {
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
