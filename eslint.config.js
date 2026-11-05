// ESLint for the browser code (web/app.js) and the Playwright tests.
import js from "@eslint/js";
import prettier from "eslint-config-prettier";
import { defineConfig, globalIgnores } from "eslint/config";
import noUnsanitized from "eslint-plugin-no-unsanitized";
import globals from "globals";

export default defineConfig(
  globalIgnores(["dist", "build", "reports", "coverage", "playwright-report", "test-results"]),
  js.configs.recommended,
  // Flags innerHTML/outerHTML/insertAdjacentHTML/document.write (DOM XSS sinks).
  noUnsanitized.configs.recommended,
  {
    files: ["web/**/*.js"],
    languageOptions: {
      sourceType: "script",
      globals: {
        ...globals.browser,
        Go: "readonly", // from Go's wasm_exec.js, inlined before app.js
        WASM_GZIP_BASE64: "readonly", // the engine, inlined by cmd/bundle
      },
    },
  },
  {
    files: ["**/*.js"],
    rules: {
      eqeqeq: ["error", "always", { null: "ignore" }],
      "no-console": "error",
      "no-eval": "error",
      "no-implied-eval": "error",
      "no-new-func": "error",
      "no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
      complexity: ["warn", 15],
      "max-depth": ["warn", 4],
    },
  },
  {
    files: ["e2e/**", "scripts/**", "*.config.js"],
    languageOptions: { globals: { ...globals.node, ...globals.browser } },
  },
  prettier,
);
