import { render, screen, waitFor } from "@testing-library/react";
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
      http.get("*/draft-assist/ready", () =>
        HttpResponse.json({ ready: true, runner: "fake" }),
      ),
    );
    render(<DraftAssistNotReadyBanner />);
    await waitFor(() => {
      expect(screen.queryByRole("status")).not.toBeInTheDocument();
    });
  });

  it("shows missing_key copy with Retry", async () => {
    server.use(
      http.get("*/draft-assist/ready", () =>
        HttpResponse.json({
          ready: false,
          runner: "sdk",
          reason: "missing_key",
        }),
      ),
    );
    render(<DraftAssistNotReadyBanner />);
    expect(
      await screen.findByText(/set CURSOR_API_KEY/i),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument();
  });

  it("shows sidecar_down copy", async () => {
    server.use(
      http.get("*/draft-assist/ready", () =>
        HttpResponse.json({
          ready: false,
          runner: "sdk",
          reason: "sidecar_down",
        }),
      ),
    );
    render(<DraftAssistNotReadyBanner />);
    expect(
      await screen.findByText(/sidecar is down/i),
    ).toBeInTheDocument();
  });
});
