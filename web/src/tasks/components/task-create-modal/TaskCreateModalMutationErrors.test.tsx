import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { TaskCreateModalMutationErrors } from "./TaskCreateModalMutationErrors";

describe("TaskCreateModalMutationErrors", () => {
  it("rewrites invalid tag createErrors without request ids", () => {
    render(
      <TaskCreateModalMutationErrors
        isTaskEdit={false}
        createError={
          new Error(
            'invalid tag "a a a a a" (request f25133d1-f58f-4362-82e4-aad920e79fdf)',
          )
        }
      />,
    );
    const callout = document.querySelector(".task-create-modal-err--create");
    expect(callout).toHaveTextContent(/Tag "a a a a a" is invalid/);
    expect(callout).toHaveTextContent(/no spaces/i);
    expect(callout).not.toHaveTextContent(/request /i);
    expect(screen.getByRole("alert")).toBeInTheDocument();
  });
});
