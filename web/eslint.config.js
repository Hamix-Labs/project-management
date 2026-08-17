import eslint from "@eslint/js";
import { defineConfig } from "eslint/config";
import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";
import globals from "globals";

export default defineConfig(
  {
    ignores: [
      "dist/**",
      "node_modules/**",
      "scripts/**",
      "src/components/tiptap-*/**",
      "src/components/tiptap-templates/**",
      "src/components/tiptap-icons/**",
      "src/components/tiptap-node/**",
      "src/components/tiptap-extension/**",
      "src/components/tiptap-ui-primitive/**",
      "src/components/tiptap-ui-utils/**",
      "src/lib/tiptap-*.ts",
      "src/hooks/use-composed-ref.ts",
      "src/hooks/use-isomorphic-layout-effect.ts",
      "src/hooks/use-on-click-outside.ts",
      "src/hooks/use-menu-navigation.ts",
      "src/hooks/use-floating-element.ts",
      "src/hooks/use-element-rect.ts",
      "src/hooks/use-unmount.ts",
      "src/hooks/use-throttled-callback.ts",
      "src/hooks/use-ui-editor-state.ts",
      "src/hooks/use-is-breakpoint.ts",
      "src/hooks/use-tiptap-editor.ts",
      "src/hooks/use-floating-toolbar-visibility.ts",
      "src/hooks/use-scrolling.ts",
      "src/hooks/use-window-size.ts",
      "src/hooks/use-cursor-visibility.ts",
      "src/contexts/**",
    ],
  },
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ["**/*.{ts,tsx}"],
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
      parserOptions: {
        ecmaFeatures: { jsx: true },
      },
      globals: {
        ...globals.browser,
      },
    },
    plugins: {
      "react-hooks": reactHooks,
    },
    rules: {
      // ESLint 10 recommended adds these; defer until a dedicated lint pass.
      "preserve-caught-error": "off",
      "no-useless-assignment": "off",
      // Core hooks only — v7 recommended also enables React Compiler rules
      // (set-state-in-effect, refs, preserve-manual-memoization, etc.) that
      // this codebase has not been migrated to yet.
      "react-hooks/rules-of-hooks": "error",
      "react-hooks/exhaustive-deps": "warn",
    },
  },
  {
    files: ["**/*.{test,spec}.{ts,tsx}"],
    languageOptions: {
      globals: {
        ...globals.browser,
        describe: "readonly",
        it: "readonly",
        expect: "readonly",
        vi: "readonly",
        beforeEach: "readonly",
        afterEach: "readonly",
        beforeAll: "readonly",
        afterAll: "readonly",
      },
    },
    rules: {
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/no-unused-vars": [
        "error",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
        },
      ],
      "max-lines": ["error", { max: 500, skipBlankLines: true, skipComments: true }],
      "no-restricted-syntax": [
        "error",
        {
          selector: "CallExpression[callee.name='setTimeout']",
          message: "Use vi.useFakeTimers() in tests — see docs/domain/web-testing.md",
        },
      ],
    },
  },
  {
    files: [
      "src/api/parseTaskApi.test.ts",
      "src/tasks/components/task-detail/cycles/TaskCyclesPanel.test.tsx",
      "src/tasks/components/task-list/section/TaskListSection.test.tsx",
    ],
    rules: {
      "max-lines": ["error", { max: 1100, skipBlankLines: true, skipComments: true }],
    },
  },
);
