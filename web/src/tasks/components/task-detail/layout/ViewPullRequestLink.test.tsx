import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ViewPullRequestLink } from "./ViewPullRequestLink";

describe("ViewPullRequestLink", () => {
  it("renders an external link when url is set", () => {
    render(
      <ViewPullRequestLink url="https://github.com/example/repo/pull/42" />,
    );
    const link = screen.getByTestId("task-detail-view-pr");
    expect(link).toHaveAttribute(
      "href",
      "https://github.com/example/repo/pull/42",
    );
    expect(link).toHaveAttribute("target", "_blank");
    expect(link).toHaveAttribute("rel", "noopener noreferrer");
    expect(link).toHaveTextContent("View PR");
  });

  it("renders nothing for empty url", () => {
    const { container } = render(<ViewPullRequestLink url="  " />);
    expect(container).toBeEmptyDOMElement();
  });
});
