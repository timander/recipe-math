import { test, expect, open, answerRight, current } from "./helpers.js";

test("Dish Pit: no clock, a wrong-way slip gets a targeted question, and a clean shift earns 3 stars", async ({
  page,
}) => {
  await open(page, { rank: "1", auto: "off" });
  await expect(page.locator("#clock")).toBeHidden();
  const o = await current(page);
  expect(o.kind).toBe("swap");
  await page.locator("#num").fill(`${o.answer.given[0]}/${o.answer.given[1]}`);
  await page.locator("#serve").click();
  await expect(page.locator("#feedback")).toContainText("That's the number you started with");
  await expect(page.locator("#feedback")).toContainText("3 plates left");
  for (let i = 0; i < 12 && !(await page.locator("#result").isVisible()); i++) {
    await answerRight(page);
    await page.locator("#next-btn").click();
  }
  await expect(page.locator("#result")).toBeVisible();
  await expect(page.locator("#r-title")).toContainText("Chef is impressed");
  await expect(page.locator("#r-review")).toContainText("Study these");
});

test("clean answers auto-advance and build a hot-line streak", async ({ page }) => {
  await open(page, { rank: "1" });
  for (let i = 0; i < 3; i++) {
    const before = (await current(page)).id;
    await answerRight(page);
    await expect.poll(async () => (await current(page)).id, { timeout: 4000 }).toBe(before + 1);
  }
  await expect(page.locator("#streak")).toHaveText("🔥 ×3");
});
