import { test, expect, open } from "./helpers.js";

test("home lists six stations, beginner first, with a start badge and gear", async ({ page }) => {
  await open(page);
  const cards = page.locator(".rank-card");
  await expect(cards).toHaveCount(6);
  await expect(cards.first()).toContainText("Dish Pit");
  await expect(cards.first()).toContainText("Start here");
  await expect(cards.first()).toContainText("No clock");
  await expect(cards.last()).toContainText("Chef de Cuisine");
  await expect(page.locator("#gear")).toContainText("Paper cap");
});

test("sound button cycles soft, full, off", async ({ page }) => {
  await page.goto((await import("./helpers.js")).APP);
  const btn = page.locator("#sound-btn");
  await expect(btn).toContainText("Soft");
  await btn.click();
  await expect(btn).toContainText("Full");
  await btn.click();
  await expect(btn).toContainText("Off");
});

test("?rank= opens straight into a station", async ({ page }) => {
  await open(page, { rank: "4" });
  await expect(page.locator("#shift")).toBeVisible();
  await expect(page.locator("#clock")).toBeVisible();
});
