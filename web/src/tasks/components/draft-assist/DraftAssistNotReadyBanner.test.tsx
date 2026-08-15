import { http, HttpResponse } from "msw";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { ensureMswListening } from "@/test/mswLifecycle";
import { server } from "@/test/server";
import {
  draftAssistReadyMissing,
  draftAssistReadyOk,
} from "@/test/handlers/draftAssist";
import { DraftAssistNotReadyBanner } from "./DraftAssistNotReadyBanner";

ensureMswListening("error");

describe("DraftAssistNotReadyBanner", () => {
  beforeEach(() => {
    server.resetHandlers();
  });

  it("stays hidden when ready", async () => {
    server.use(draftAssistReadyOk("fake"));
    render(<DraftAssistNotReadyBanner />);
    await expect(
      screen.findByRole("status", {}, { timeout: 300 }),
    ).rejects.toThrow();
  });

  it("shows missing_key copy with Retry", async () => {
    server.use(draftAssistReadyMissing("missing_key", "sdk"));
    render(<DraftAssistNotReadyBanner />);
    expect(
      await screen.findByText(/set CURSOR_API_KEY/i),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument();
  });

  it("shows sidecar_down copy", async () => {
    server.use(draftAssistReadyMissing("sidecar_down", "sdk"));
    render(<DraftAssistNotReadyBanner />);
    expect(
      await screen.findByText(/sidecar is down/i),
    ).toBeInTheDocument();
  });

  it("shows taskapi-down copy when the probe cannot reach the server", async () => {
    server.use(
      http.get("/draft-assist/ready", () =>
        HttpResponse.json({ error: "down" }, { status: 503 }),
      ),
    );
    render(<DraftAssistNotReadyBanner />);
    expect(
      await screen.findByText(/taskapi is not running/i),
    ).toBeInTheDocument();
  });
});
