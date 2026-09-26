// Copyright NU Cybernetics. p(DOOM) — research prototype.
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: [
      "packages/*/test/**/*.test.ts",
      "packages/*/src/**/*.test.ts",
      "services/mcp/test/**/*.test.ts",
      "apps/web/test/**/*.test.{ts,tsx}",
    ],
    environment: "node",
    passWithNoTests: true,
  },
});
