import { test, expect, open, playUntil, answerWrong } from "./helpers.js";

test("the bowl groups scoops and a wrong amount asks a question", async ({ page }) => {
  await open(page, { rank: "3", auto: "off" });
  const o = await playUntil(page, (x) => x.input === "build" && x.drawer === "dry");
  const spoon = o.tools.find((t) => t.id === "T1");
  for (let i = 0; i < 3; i++) await page.getByRole("button", { name: `Add ${spoon.name}` }).click();
  await expect(page.locator("#bowl .scoop")).toHaveCount(1);
  await expect(page.locator("#bowl .scoop")).toContainText("×3");
  await page.locator("#bowl .scoop").click(); // take one out
  await expect(page.locator("#bowl .scoop")).toContainText("×2");
  await page.locator("#bowl-clear").click();
  await answerWrong(page);
  await expect(page.locator("#feedback")).toContainText("Your scoops add up to");
});

test("on phones the bowl is a sticky tray with its own Serve button", async ({ page, isMobile }) => {
  test.skip(!isMobile, "phone layout only");
  await open(page, { rank: "3", auto: "off" });
  await playUntil(page, (x) => x.input === "build");
  await expect(page.locator(".tray-serve")).toBeVisible();
  await expect(page.locator("#serve")).toBeHidden();
});
