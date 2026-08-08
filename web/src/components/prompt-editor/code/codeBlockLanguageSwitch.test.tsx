import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeAll, describe, expect, it } from "vitest";
import { BlockNotePromptEditor } from "../BlockNotePromptEditor";

const CODE_HTML =
  '<pre><code class="language-javascript" data-language="javascript">const total = 1;</code></pre>';

function blockLanguage(): string | undefined {
  const select = document.querySelector(
    '[data-content-type="codeBlock"] select',
  );
  return (select as HTMLSelectElement | null)?.value;
}

async function pickLanguage(
  user: ReturnType<typeof userEvent.setup>,
  current: RegExp,
  next: RegExp,
) {
  await user.click(await screen.findByRole("button", { name: current }));
  await user.click(screen.getByRole("button", { name: next }));
}

/**
 * Guards the real wiring: picking a language must reach the block props, not just
 * the node view's hidden <select>, which BlockNote destroys and recreates on every
 * update — a stale reference used to swallow the second switch.
 */
describe("code block language switching", () => {
  beforeAll(() => {
    // jsdom lacks elementsFromPoint, which BlockNote's side menu calls on hover.
    Object.defineProperty(document, "elementsFromPoint", {
      configurable: true,
      value: () => [],
    });
  });

  it("writes each chosen language to the block and follows it in the label", async () => {
    const user = userEvent.setup();
    render(
      <BlockNotePromptEditor
        id="prompt"
        initialHtml={CODE_HTML}
        onChange={() => {}}
      />,
    );

    await pickLanguage(user, /javascript/i, /^go$/i);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /^go/i })).toBeInTheDocument();
    });
    await waitFor(() => expect(blockLanguage()).toBe("go"));

    await pickLanguage(user, /^go/i, /^python$/i);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /^python/i }),
      ).toBeInTheDocument();
    });
    await waitFor(() => expect(blockLanguage()).toBe("python"));
  });
});
