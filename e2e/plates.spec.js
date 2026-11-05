import { test, expect, open, answerWrong, current } from "./helpers.js";

test("plates read as words: ½ plate left", async ({ page }) => {
  await open(page, { rank: "2", auto: "off", plates: "2" });
  await page.locator("#hint-btn").click(); // 1½ plates
  await answerWrong(page); // ½ plate
  await expect(page.locator("#feedback")).toContainText("(½ plate left)");
});

test("knockout freezes the ticket at once, then shows the review", async ({ page }) => {
  await open(page, { rank: "2", auto: "off", plates: "1" });
  const o = await answerWrong(page);
  // Regression (/code-review): during the pause before the result, serving or skipping still worked.
  await expect(page.locator("#serve")).toBeDisabled();
  await expect(page.locator("#skip-btn")).toBeDisabled();
  expect(await page.evaluate(() => S.phase)).toBe("out");
  await expect(page.locator("#result")).toBeVisible();
  await expect(page.locator("#r-title")).toHaveText("Chef pulled you off the line!");
  await expect(page.locator("#r-review")).toContainText(o.prompt);
});

test("leaving for another tab pauses the shift clock and keyboard", async ({ page, isMobile }) => {
  test.skip(isMobile, "keyboard shortcuts are for Chromebooks");
  await open(page, { rank: "2", auto: "off" });
  await page.locator("#tab-worksheet").click();
  const t0 = await page.evaluate(() => S.timeLeft);
  await page.waitForTimeout(2200);
  expect(await page.evaluate(() => S.timeLeft)).toBe(t0);
  const before = (await current(page)).id;
  await page.locator("body").press("a");
  await page.locator("body").press("a");
  expect(await page.evaluate(() => S.accidents)).toBe(0);
  await page.locator("#tab-kitchen").click();
  expect((await current(page)).id).toBe(before);
});
