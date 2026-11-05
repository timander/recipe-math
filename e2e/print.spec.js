import { test, expect, open } from "./helpers.js";

const pages = (pdf) => (pdf.toString("latin1").match(/\/Type\s*\/Page[^s]/g) || []).length;

test.describe("paper", () => {
  test.skip(({ isMobile }) => isMobile, "printing is tested once, on desktop");

  test("worksheet and answer key share a version", async ({ page }) => {
    await open(page);
    await page.locator("#tab-worksheet").click();
    await page.locator('#ws-ranks input[value="1"]').check();
    await page.locator("#ws-version").fill("4821");
    const sheet = page.locator("#ws-sheet");
    await expect(sheet).toContainText("Version 4821");
    await expect(sheet.locator(".cup-wall")).toHaveCount(0);
    const firstQ = await sheet.locator(".ws-item").first().textContent();
    await page.locator('input[name="ws-preview"][value="key"]').check();
    await expect(sheet).toContainText("Answer Key");
    await expect(sheet.locator(".ws-item").first()).toContainText(firstQ.split("Answer")[0].trim().slice(0, 20));
    await page.emulateMedia({ media: "print" });
    expect(pages(await page.pdf({ format: "Letter", preferCSSPageSize: true }))).toBe(1);
  });

  test("the cheat sheet prints on one page", async ({ page }) => {
    await open(page);
    await page.locator("#tab-cheatsheet").click();
    await expect(page.locator("#cheat-sheet .cup-wall")).toBeVisible();
    await page.emulateMedia({ media: "print" });
    expect(pages(await page.pdf({ format: "Letter", preferCSSPageSize: true }))).toBe(1);
  });
});
