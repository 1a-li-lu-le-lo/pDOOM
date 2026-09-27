// Copyright NU Cybernetics. p(DOOM) — research prototype.
import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "@playwright/test";

const PAGES = ["/", "/meter", "/futures", "/forecasts", "/lab", "/act", "/text", "/method/model"];

test.describe("accessibility", () => {
  test.skip(({ browserName, isMobile }) => browserName !== "chromium" || !!isMobile, "axe runs once, on desktop chromium");
  for (const path of PAGES) {
    test(`${path} has no serious or critical axe violations`, async ({ page }) => {
      await page.goto(path);
      const results = await new AxeBuilder({ page }).withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"]).exclude("canvas").analyze();
      const serious = results.violations.filter((v) => v.impact === "serious" || v.impact === "critical");
      expect(serious.map((v) => `${v.id}: ${v.nodes.map((n) => n.target.join(" ")).slice(0, 3).join(" | ")}`)).toEqual([]);
    });
  }

  test("the skip link is the first focusable element and lands on main", async ({ page }) => {
    await page.goto("/meter");
    await page.keyboard.press("Tab");
    await expect(page.locator(".skip-link")).toBeFocused();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(/#main$/);
  });

  test("every chart has a caption and an accessible data table", async ({ page }) => {
    await page.goto("/meter");
    const figures = page.locator("figure");
    const n = await figures.count();
    expect(n).toBeGreaterThan(0);
    for (let i = 0; i < n; i++) {
      const fig = figures.nth(i);
      await expect(fig.locator("figcaption")).toHaveCount(1);
      await expect(fig.locator("details table").first()).toBeAttached();
    }
    expect(await page.locator("figure svg[role='img'][aria-label]").count()).toBeGreaterThan(0);
  });

  test("colour is never the only channel on badges", async ({ page }) => {
    await page.goto("/meter");
    const badges = page.locator(".badge");
    expect(await badges.count()).toBeGreaterThan(0);
    for (const b of (await badges.all()).slice(0, 12)) {
      expect((await b.innerText()).trim().length).toBeGreaterThan(0);
    }
  });
});
