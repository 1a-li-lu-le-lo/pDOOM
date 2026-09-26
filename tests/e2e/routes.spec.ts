// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { expect, test } from "@playwright/test";

const ROUTES = ["/", "/meter", "/futures", "/futures/S1", "/evidence", "/capabilities", "/agents", "/incidents", "/forecasts", "/safeguards", "/act", "/method", "/method/model", "/method/definitions", "/lab", "/changelog", "/compare", "/text"];

for (const route of ROUTES) {
  test(`${route} renders with one h1, breadcrumbs or hero, and a plain-text link`, async ({ page }) => {
    const res = await page.goto(route);
    expect(res?.status()).toBe(200);
    await expect(page.locator("h1")).toHaveCount(1);
    if (route === "/text") await expect(page.locator("main a[href='/']").first()).toBeVisible();
    else await expect(page.locator("main a[href^='/text']").first()).toBeVisible();
    await expect(page.locator("header .brand")).toContainText("p(DOOM)");
    // The footer says which release the page was rendered from.
    await expect(page.locator("footer")).toContainText(/Release rel-\d{4}-\d{2}-\d{2}-\d{3}/);
  });
}

test("every probability shown on the meter travels with its status and horizon", async ({ page }) => {
  await page.goto("/meter");
  const cards = page.locator("article.estimate-card");
  expect(await cards.count()).toBeGreaterThan(0);
  for (const card of await cards.all()) {
    await expect(card.locator(".badge").first()).toBeVisible();
    const label = await card.getAttribute("aria-label");
    expect(label).toMatch(/Official value withheld|External forecast aggregate|Research mode/);
    expect(label).toMatch(/year|2100|eventual|Eventual/i);
  }
  await expect(page.getByRole("heading", { level: 2, name: /Insufficiently calibrated/ })).toBeVisible();
});

test("horizon tabs change the meter view and the URL is shareable", async ({ page }) => {
  await page.goto("/meter");
  await page.getByRole("navigation", { name: "Horizon" }).getByRole("link", { name: /2100/ }).click();
  await expect(page).toHaveURL(/horizon=2100/);
  await expect(page.locator(".meter-headline .eyebrow")).toContainText(/2100/);
});

test("the plain-text route is complete without JavaScript", async ({ browser }) => {
  const context = await browser.newContext({ javaScriptEnabled: false });
  const page = await context.newPage();
  const res = await page.goto("/text");
  expect(res?.status()).toBe(200);
  for (const id of ["estimate", "definition", "uncertainty", "decomposition", "forecasts", "scenarios", "safeguards", "act", "method", "sources", "limitations", "governance"]) {
    await expect(page.locator(`h2#${id}`)).toHaveCount(1);
  }
  await expect(page.locator("body")).toContainText("Insufficiently calibrated");
  await expect(page.locator("canvas")).toHaveCount(0);
  await context.close();
});

test("exports are machine readable and name their release", async ({ request }) => {
  const rel = await request.get("/api/export/release.json");
  expect(rel.ok()).toBeTruthy();
  expect(rel.headers()["x-pdoom-release"]).toMatch(/^rel-/);
  const body = (await rel.json()) as { manifest: { release_id: string }; estimates: { status: string; quantiles: unknown }[] };
  expect(body.manifest.release_id).toBe(rel.headers()["x-pdoom-release"]);
  for (const e of body.estimates.filter((x) => x.status === "insufficiently_calibrated")) expect(e.quantiles).toBeNull();
  const csv = await request.get("/api/export/estimates.csv");
  expect(csv.headers()["content-type"]).toContain("text/csv");
  expect((await csv.text()).split("\r\n")[0]).toContain("estimate_id,status,outcome_set,horizon");
  const missing = await request.get("/api/export/nothing.csv");
  expect(missing.status()).toBe(404);
});

test("the share card renders as a PNG", async ({ request }) => {
  const res = await request.get("/opengraph-image");
  expect(res.ok()).toBeTruthy();
  expect(res.headers()["content-type"]).toContain("image/png");
});

test("unknown routes show the not-found page with a way back", async ({ page }) => {
  const res = await page.goto("/does-not-exist");
  expect(res?.status()).toBe(404);
  await expect(page.getByRole("heading", { level: 1 })).toContainText("Not found");
  await expect(page.getByRole("link", { name: "The meter" })).toBeVisible();
});
