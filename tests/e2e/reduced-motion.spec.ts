// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { expect, test } from "@playwright/test";

test.describe("reduced motion", () => {
  test.skip(({ contextOptions }) => contextOptions.reducedMotion !== "reduce", "runs in the reduced-motion project only");

  test("defaults to the observatory mode and renders no canvas", async ({ page }) => {
    await page.goto("/");
    await expect(page.locator("header").getByRole("group", { name: "Presentation mode" }).getByRole("button", { name: "Observatory" })).toHaveAttribute("aria-pressed", "true");
    await expect(page.locator("canvas")).toHaveCount(0);
    await expect(page.locator(".hero-stage .static-disk")).toBeVisible();
  });

  test("choosing an immersive mode still avoids animation", async ({ page }) => {
    await page.goto("/");
    await page.locator("header").getByRole("group", { name: "Presentation mode" }).getByRole("button", { name: "Event Horizon" }).click();
    await page.waitForTimeout(500);
    await expect(page.locator("canvas")).toHaveCount(0);
  });
});
