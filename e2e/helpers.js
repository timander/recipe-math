// Shared helpers: open the app, read the current ticket, answer it, and
// collect JavaScript coverage of the page's own code (Chromium only).
import { test as base, expect } from "@playwright/test";
import { mkdirSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
export const APP = "file://" + join(root, "dist", "recipe-math.html");

export const test = base.extend({
  page: async ({ page, browserName }, use, testInfo) => {
    const errors = [];
    page.on("pageerror", (e) => errors.push(e.message));
    page.on("console", (m) => m.type() === "error" && errors.push(m.text()));
    const coverage = browserName === "chromium";
    if (coverage) await page.coverage.startJSCoverage({ resetOnNavigation: false });
    await use(page);
    if (coverage) {
      const entries = await page.coverage.stopJSCoverage();
      const dir = join(root, "reports", "e2e-coverage-raw");
      mkdirSync(dir, { recursive: true });
      writeFileSync(
        join(dir, `${testInfo.testId}.json`),
        JSON.stringify(entries.filter((e) => e.url.startsWith("file:"))),
      );
    }
    expect(errors, "no errors in the page").toEqual([]);
  },
});
export { expect };

// open loads the app with optional URL settings (sound is always off in tests).
export async function open(page, settings = {}) {
  const q = new URLSearchParams({ sound: "off", seed: "11", ...settings });
  await page.goto(`${APP}?${q}`);
  await expect(page.locator("#loading")).toBeHidden();
}

export const current = (page) => page.evaluate(() => cur());

// serveButton is the visible Serve button (on phones, measuring uses the tray's).
async function serveButton(page) {
  return (await page.locator("#serve").isVisible()) ? page.locator("#serve") : page.locator(".tray-serve");
}

// answerRight serves the correct answer for the current ticket.
export async function answerRight(page) {
  const o = await current(page);
  if (o.input === "choice") {
    await page.locator(`#in-choice [data-choice="${o.answer.choices[0]}"]`).click();
    await page.locator("#serve").click();
  } else if (o.input === "number") {
    await page.locator("#num").fill(`${o.answer.num[0]}/${o.answer.num[1]}`);
    await page.locator("#serve").click();
  } else {
    await page.evaluate(() => {
      S.bowl = cur().answer.best.slice();
      renderBowl();
      renderGlass(cur());
    });
    await (await serveButton(page)).click();
  }
  await expect(page.locator("#next-btn")).toBeVisible();
  return o;
}

// answerWrong serves a wrong answer for the current ticket.
export async function answerWrong(page) {
  const o = await current(page);
  if (o.input === "choice") {
    const used = await page.evaluate(() => curRec().wrongChoices);
    const bad = o.choices.findIndex((_, i) => !o.answer.choices.includes(i) && !used.includes(i));
    await page.locator(`#in-choice [data-choice="${bad}"]`).click();
    await page.locator("#serve").click();
  } else if (o.input === "number") {
    await page.locator("#num").fill("999");
    await page.locator("#serve").click();
  } else {
    await page.evaluate(() => {
      S.bowl = [cur().tools.at(-1).id];
      renderBowl();
    });
    await (await serveButton(page)).click();
  }
  return o;
}

// playUntil answers tickets correctly until pred(order) is true; returns that order.
export async function playUntil(page, pred, limit = 30) {
  for (let i = 0; i < limit; i++) {
    const o = await current(page);
    if (pred(o)) return o;
    await answerRight(page);
    await page.locator("#next-btn").click();
  }
  throw new Error("no matching ticket");
}
