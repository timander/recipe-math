import { test, expect, open, playUntil, answerRight, current } from "./helpers.js";

test("recipe card: the factor question highlights no ingredient row, then the card fills in", async ({ page }) => {
  await open(page, { rank: "5", auto: "off" });
  await playUntil(page, (o) => o.kind === "factor" && o.card);
  // Regression (/code-review): the factor order used to highlight line 0.
  await expect(page.locator("#card .card-lines li.active")).toHaveCount(0);
  await expect(page.locator("#card .factor")).toHaveText("×?");
  await answerRight(page);
  await page.locator("#next-btn").click();
  await expect(page.locator("#card .factor")).not.toHaveText("×?");
  await expect(page.locator("#card .card-lines li.active")).toHaveCount(1);
  let o = await current(page);
  while (o.card && o.kind !== "fix") {
    await answerRight(page);
    const done = await page.locator("#feedback").textContent();
    await page.locator("#next-btn").click();
    if (done.includes("Recipe card complete")) break;
    o = await current(page);
  }
});

test("Chef de Cuisine: Fix the Mistake is a side-by-side ledger", async ({ page }) => {
  await open(page, { rank: "6", auto: "off" });
  const o = await playUntil(page, (x) => x.kind === "fix");
  await expect(page.locator("#card")).toBeHidden();
  await expect(page.locator("#in-choice.ledger [data-choice]")).toHaveCount(o.card.lines.length);
  const row = page.locator(`#in-choice [data-choice="${o.answer.choices[0]}"]`);
  await expect(row.locator(".lg-orig")).toHaveText(o.card.lines[o.answer.choices[0]].text);
  await row.click();
  await page.locator("#serve").click();
  await expect(page.locator("#feedback")).toContainText("Correct");
});
