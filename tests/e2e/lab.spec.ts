// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { expect, test } from "@playwright/test";

test.describe("scenario lab", () => {
  test("results are labelled as the user's scenario and the URL reproduces them", async ({ page }) => {
    await page.goto("/lab");
    const result = page.locator(".lab-result");
    await expect(result.locator(".badge", { hasText: "Your scenario" })).toBeVisible();
    await expect(result).toContainText("not the p(DOOM) official model");
    await expect(result.locator(".lab-value")).not.toHaveText("…");
    const baseline = await result.locator(".lab-value").innerText();

    const slider = page.locator("#s-safety_progress");
    await slider.focus();
    await page.keyboard.press("ArrowRight");
    await page.keyboard.press("ArrowRight");
    await expect(page).toHaveURL(/sp=2/);
    await expect(result.locator(".badge", { hasText: /dials moved/ })).toBeVisible();

    const url = page.url();
    const context2 = await page.context().browser()!.newContext();
    const page2 = await context2.newPage();
    await page2.goto(url);
    await expect(page2.locator("#s-safety_progress")).toHaveValue("2");
    await expect(page2.locator(".lab-result .lab-value")).not.toHaveText("…");
    const value2 = await page2.locator(".lab-result .lab-value").innerText();
    const value1 = await result.locator(".lab-value").innerText();
    expect(value2).toBe(value1);
    expect(typeof baseline).toBe("string");
    await context2.close();
  });

  test("implausible dial combinations are flagged, not refused", async ({ page }) => {
    await page.goto("/lab?h=10y&sp=2&gs=-2&ic=-2");
    await expect(page.locator(".flags .badge").first()).toContainText(/implausible/);
    await expect(page.locator(".lab-result .lab-value")).not.toHaveText("…");
  });

  test("the published research-mode run is shown beside the user scenario and does not change", async ({ page }) => {
    await page.goto("/lab?h=10y&ct=2&ag=2");
    await expect(page.locator(".lab-result")).toContainText(/Published research-mode run for this horizon/);
    await expect(page.locator(".lab-result")).toContainText(/does not change it/);
  });
});
