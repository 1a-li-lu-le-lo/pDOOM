// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { expect, test } from "@playwright/test";

test.describe("presentation modes", () => {
  test("the mode switcher is in the first viewport and remembers the choice locally", async ({ page }) => {
    await page.goto("/");
    const switcher = page.getByRole("group", { name: "Presentation mode" });
    await expect(switcher).toBeInViewport();
    await switcher.getByRole("button", { name: "Observatory" }).click();
    await expect(switcher.getByRole("button", { name: "Observatory" })).toHaveAttribute("aria-pressed", "true");
    expect(await page.evaluate(() => window.localStorage.getItem("pdoom.mode"))).toBe("observatory");
    await page.reload();
    await expect(switcher.getByRole("button", { name: "Observatory" })).toHaveAttribute("aria-pressed", "true");
    // No cookie is ever set.
    expect(await page.context().cookies()).toEqual([]);
  });

  test("the plain-text link in the switcher leads to /text", async ({ page }) => {
    await page.goto("/");
    await page.getByRole("group", { name: "Presentation mode" }).getByRole("link", { name: "Plain text" }).click();
    await expect(page).toHaveURL(/\/text$/);
    await expect(page.locator("h1")).toContainText(/plain text/i);
  });

  test("the meter is readable before and without any 3D asset", async ({ page }) => {
    await page.route("**/*.js", (route) => (route.request().url().includes("three") ? route.abort() : route.continue()));
    await page.goto("/");
    await expect(page.getByRole("heading", { level: 2, name: /Insufficiently calibrated/ })).toBeVisible();
    await expect(page.locator(".static-disk").first()).toBeVisible();
  });

  test("when WebGL is unavailable the static disk is shown and nothing breaks", async ({ page }) => {
    await page.addInitScript(() => {
      const orig = HTMLCanvasElement.prototype.getContext;
      // @ts-expect-error test shim
      HTMLCanvasElement.prototype.getContext = function (type: string, ...rest: unknown[]) {
        if (typeof type === "string" && type.includes("webgl")) return null;
        return orig.call(this, type as never, ...(rest as never[]));
      };
      window.localStorage.setItem("pdoom.mode", "event-horizon");
    });
    const errors: string[] = [];
    page.on("pageerror", (e) => errors.push(e.message));
    await page.goto("/");
    await expect(page.locator(".hero-stage .static-disk")).toBeVisible();
    await expect(page.locator(".hero-stage")).toHaveAttribute("data-scene", /static|event-horizon/);
    expect(errors).toEqual([]);
  });

  test("the hero labels the scene as conceptual and never shows a countdown", async ({ page }) => {
    await page.goto("/");
    await expect(page.locator(".scene-label")).toContainText(/not a simulation/i);
    const text = await page.locator("main").innerText();
    expect(text).not.toMatch(/years left|countdown|doom is certain/i);
  });
});
