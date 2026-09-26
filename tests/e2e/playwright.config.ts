// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { defineConfig, devices } from "@playwright/test";

const PORT = Number(process.env.PDOOM_E2E_PORT ?? 3117);
const baseURL = `http://127.0.0.1:${PORT}`;

export default defineConfig({
  testDir: ".",
  testMatch: /.*\.spec\.ts/,
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: true,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [["list"], ["html", { open: "never" }]] : "list",
  use: {
    baseURL,
    trace: "retain-on-failure",
    launchOptions: { executablePath: process.env.PDOOM_CHROMIUM ?? "/opt/pw-browsers/chromium" },
  },
  webServer: {
    command: `pnpm --filter @pdoom/web exec next start -p ${PORT}`,
    url: baseURL,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
  projects: [
    { name: "desktop", use: { ...devices["Desktop Chrome"] } },
    { name: "mobile", use: { ...devices["Pixel 7"] } },
    { name: "reduced-motion", use: { ...devices["Desktop Chrome"], contextOptions: { reducedMotion: "reduce" } } },
  ],
});
