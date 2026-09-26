// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { expect, test } from "@playwright/test";

test.describe("immersive scenes", () => {
  test.skip(({ contextOptions, isMobile }) => contextOptions.reducedMotion === "reduce" || !!isMobile, "desktop, motion allowed");

  test("the event horizon canvas mounts over the static disk and stays free of numbers", async ({ page }) => {
    await page.addInitScript(() => window.localStorage.setItem("pdoom.mode", "event-horizon"));
    const errors: string[] = [];
    page.on("pageerror", (e) => errors.push(e.message));
    await page.goto("/");
    await expect(page.locator(".hero-stage")).toHaveAttribute("data-scene", "event-horizon");
    await expect(page.locator(".hero-stage .static-disk")).toBeVisible();
    await expect(page.locator(".scene-canvas[data-ready='true']")).toBeVisible({ timeout: 20_000 });
    await expect(page.locator(".scene-canvas")).toHaveAttribute("aria-hidden", "true");
    expect(errors).toEqual([]);
    // The meter below the stage is unaffected by the scene.
    await expect(page.getByRole("heading", { level: 2, name: /Insufficiently calibrated/ })).toBeVisible();
  });

  for (const mode of ["orrery", "branching"] as const) {
    test(`the ${mode} scene is an SVG with a conceptual label and no percentages`, async ({ page }) => {
      await page.addInitScript((m) => window.localStorage.setItem("pdoom.mode", m), mode);
      await page.goto("/");
      await expect(page.locator(".hero-stage")).toHaveAttribute("data-scene", mode);
      const svg = page.locator(".hero-stage svg.scene-svg");
      await expect(svg).toBeVisible();
      await expect(svg).toHaveAttribute("aria-label", /not a simulation/i);
      const text = (await svg.textContent()) ?? "";
      expect(text).not.toMatch(/\d+\s?%/);
    });
  }

  test("switching modes never changes the published numbers", async ({ page }) => {
    await page.goto("/");
    const before = await page.locator(".meter").innerText();
    for (const name of ["Observatory", "Event Horizon"]) {
      await page.getByRole("group", { name: "Presentation mode" }).getByRole("button", { name }).click();
      await page.waitForTimeout(300);
      expect(await page.locator(".meter").innerText()).toBe(before);
    }
  });
});
