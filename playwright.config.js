// Playwright UI integration tests against the built single-file app
// (dist/recipe-math.html), opened straight from disk like a teacher would.
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI
    ? [
        ["github"],
        ["html", { open: "never", outputFolder: "reports/playwright" }],
        ["json", { outputFile: "reports/playwright.json" }],
      ]
    : [
        ["list"],
        ["html", { open: "never", outputFolder: "reports/playwright" }],
        ["json", { outputFile: "reports/playwright.json" }],
      ],
  globalTeardown: "./e2e/coverage-teardown.js",
  use: { trace: "retain-on-failure", screenshot: "only-on-failure" },
  projects: [
    { name: "desktop", use: { ...devices["Desktop Chrome"], viewport: { width: 1366, height: 900 } } },
    { name: "phone", use: { ...devices["Pixel 7"] } },
  ],
});
