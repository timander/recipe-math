import { test, expect, open, playUntil } from "./helpers.js";

test("a choice is picked first, then served, so a slip doesn't cost a plate", async ({ page }) => {
  await open(page, { rank: "2", auto: "off" });
  const o = await playUntil(page, (x) => x.input === "choice");
  const plates = await page.evaluate(() => S.plates);
  const right = o.answer.choices[0];
  const wrong = o.choices.findIndex((_, i) => !o.answer.choices.includes(i));
  await page.locator(`#in-choice [data-choice="${wrong}"]`).click();
  await expect(page.locator(`#in-choice [data-choice="${wrong}"]`)).toHaveAttribute("aria-pressed", "true");
  await expect(page.locator("#feedback")).not.toContainText("Accident");
  expect(await page.evaluate(() => S.plates)).toBe(plates);
  await page.locator(`#in-choice [data-choice="${right}"]`).click();
  await expect(page.locator("#serve")).toHaveText(`🛎 Serve ${"ABCDE"[right]}`);
  await page.locator(`#in-choice [data-choice="${right}"]`).click(); // second tap serves
  await expect(page.locator("#feedback")).toContainText("Correct");
  // Untouched options are locked but not crossed out.
  await expect(page.locator(`#in-choice [data-choice="${wrong}"]`)).not.toHaveClass(/struck/);
});

test("keyboard letters only select; Enter serves", async ({ page, isMobile }) => {
  test.skip(isMobile, "keyboard shortcuts are for Chromebooks");
  await open(page, { rank: "2", auto: "off" });
  const o = await playUntil(page, (x) => x.input === "choice");
  const right = o.answer.choices[0];
  await page.locator("body").press("abcde"[right]);
  await expect(page.locator(`#in-choice [data-choice="${right}"]`)).toHaveAttribute("aria-pressed", "true");
  await expect(page.locator("#feedback")).not.toContainText("Correct");
  await page.keyboard.press("Enter");
  await expect(page.locator("#feedback")).toContainText("Correct");
});
