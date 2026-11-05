import { test, expect, open, playUntil, answerRight } from "./helpers.js";

test("the last hint shows the cup wall with the amount shaded, and Chef says shades the answer", async ({ page }) => {
  await open(page, { rank: "1", auto: "off" });
  let o;
  for (let tries = 0; tries < 6 && !o; tries++) {
    try {
      o = await playUntil(page, (x) => x.visual && x.visual.kind === "wall", 10);
    } catch {
      await open(page, { rank: "1", auto: "off" });
    }
  }
  expect(o, "a cup-wall ticket").toBeTruthy();
  await page.evaluate(() => {
    curRec().shown = cur().hints.length;
    renderHints(cur(), curRec());
  });
  const wall = page.locator("#hints .cup-wall");
  await expect(wall).toBeVisible();
  await expect(wall.locator(".wall-row")).toHaveCount(6);
  await expect(wall.locator(".wall-shade")).toBeVisible();
  await answerRight(page);
  await expect(page.locator("#explain .cup-wall .wall-shade")).toBeVisible();
});
