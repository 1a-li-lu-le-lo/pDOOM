// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { expect, test } from "@playwright/test";

test.describe("shell", () => {
  test("all five modes are reachable from the home page", async ({ page }) => {
    await page.goto("/");
    const full = page.locator(".hero-modes .mode-switcher");
    for (const name of ["Event Horizon", "Orrery", "Branching Futures", "Observatory"]) {
      await expect(full.getByRole("button", { name })).toBeVisible();
    }
    await expect(full.getByRole("link", { name: "Plain text" })).toBeVisible();
  });

  test("display settings apply a theme and palette and persist locally", async ({ page }) => {
    await page.goto("/meter");
    const settings = page.getByRole("form", { name: "Display settings" });
    await settings.getByLabel("Theme").selectOption("light");
    await settings.getByLabel("Palette").selectOption("cvd");
    await expect(page.locator("html")).toHaveAttribute("data-theme", "light");
    await expect(page.locator("html")).toHaveAttribute("data-palette", "cvd");
    await page.reload();
    await expect(page.locator("html")).toHaveAttribute("data-theme", "light");
    await expect(page.locator("html")).toHaveAttribute("data-palette", "cvd");
    expect(await page.context().cookies()).toEqual([]);
    await settings.getByLabel("Theme").selectOption("system");
    await expect(page.locator("html")).not.toHaveAttribute("data-theme", /.+/);
  });

  test("robots and sitemap are served and name the release pages", async ({ request }) => {
    const robots = await request.get("/robots.txt");
    expect(robots.ok()).toBeTruthy();
    expect(await robots.text()).toContain("Sitemap:");
    const sitemap = await request.get("/sitemap.xml");
    expect(sitemap.ok()).toBeTruthy();
    const xml = await sitemap.text();
    expect(xml).toContain("/meter");
    expect(xml).toContain("/futures/S1");
    expect(xml).toContain("/releases/rel-");
  });

  test("the evidence ledger says so when a filter matches nothing", async ({ page }) => {
    await page.goto("/evidence?q=zzzz-no-such-source");
    await expect(page.getByText(/0 of \d+ sources match these filters/)).toBeVisible();
  });

  test("the header keeps to two rows and the nav never wraps on a phone", async ({ page, isMobile }) => {
    test.skip(!isMobile, "mobile only");
    await page.goto("/");
    const nav = page.locator(".site-nav");
    const box = await nav.boundingBox();
    expect(box).not.toBeNull();
    expect(box!.height).toBeLessThan(60);
    const header = await page.locator(".site-header").boundingBox();
    expect(header!.height).toBeLessThan(170);
  });
});

test.describe("console hygiene", () => {
  test.skip(({ isMobile }) => !!isMobile, "desktop is enough for console checks");
  for (const route of ["/", "/meter", "/futures", "/capabilities", "/forecasts", "/compare", "/lab", "/text"]) {
    test(`${route} logs no errors or hydration warnings`, async ({ page }) => {
      const problems: string[] = [];
      page.on("pageerror", (e) => problems.push(`pageerror: ${e.message}`));
      page.on("console", (m) => {
        if (m.type() === "error") problems.push(`console.error: ${m.text()}`);
        if (m.type() === "warning" && /hydrat|did not match|Minified React error/i.test(m.text())) problems.push(`console.warning: ${m.text()}`);
      });
      await page.goto(route, { waitUntil: "networkidle" });
      await page.waitForTimeout(500);
      expect(problems).toEqual([]);
    });
  }
});
