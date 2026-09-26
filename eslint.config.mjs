// Copyright NU Cybernetics. p(DOOM) — research prototype.
import js from "@eslint/js";
import tseslint from "typescript-eslint";
import reactHooks from "eslint-plugin-react-hooks";

export default tseslint.config(
  {
    ignores: [
      "**/node_modules/**",
      "**/.next/**",
      "**/dist/**",
      "**/coverage/**",
      "**/playwright-report/**",
      "**/test-results/**",
      "data/**",
      "**/next-env.d.ts",
    ],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ["**/*.{ts,tsx,mts,cts}"],
    plugins: { "react-hooks": reactHooks },
    rules: {
      ...reactHooks.configs.recommended.rules,
      "@typescript-eslint/no-unused-vars": ["error", { argsIgnorePattern: "^_", varsIgnorePattern: "^_" }],
      "@typescript-eslint/no-explicit-any": "warn",
      "no-console": ["warn", { allow: ["warn", "error", "info"] }],
    },
  },
  {
    files: ["**/*.{js,mjs,cjs}", "**/scripts/**/*.ts", "tools/**/*.ts"],
    languageOptions: {
      globals: {
        process: "readonly", console: "readonly", URL: "readonly", Buffer: "readonly",
        setTimeout: "readonly", clearTimeout: "readonly", fetch: "readonly", globalThis: "readonly",
      },
    },
    rules: { "no-console": "off" },
  },
);
