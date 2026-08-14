import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { server } from "@/test/server";
import { DraftAssistNotReadyBanner } from "./DraftAssistNotReadyBanner";

describe("DraftAssistNotReadyBanner", () => {
  beforeEach(() => {
    server.resetHandlers();
  });

  it("stays hidden when ready", async () => {
    server.use(
      http.get("/draft-assist/ready", () =>
        HttpResponse.json({ ready: true, runner: "fake" }),
      ),
    );
    render(<DraftAssistNotReadyBanner />);
    await waitFor(() => {
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });
  });

  it("shows missing_key copy and retries", async () => {
    let attempts = 0;
    server.use(
      http.get("/draft-assist/ready", () => {
        attempts += 1;
        if (attempts === 1) {
          return HttpResponse.json({
            ready: false,
            runner: "sdk",
            reason: "missing_key",
          });
        }
        return HttpResponse.json({ ready: true, runner: "sdk" });
      }),
    );
    const user = userEvent.setup();
    render(<DraftAssistNotReadyBanner />);
    expect(
      await screen.findByText(/set CURSOR_API_KEY/i),
    ).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /retry/i }));
    await waitFor(() => {
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });
  });
});
