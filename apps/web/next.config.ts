// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { NextConfig } from "next";

const csp = [
  "default-src 'self'",
  "script-src 'self' 'unsafe-inline'",
  "style-src 'self' 'unsafe-inline'",
  "img-src 'self' data: blob:",
  "font-src 'self' data:",
  "worker-src 'self' blob:",
  "connect-src 'self'",
  "frame-ancestors 'none'",
  "base-uri 'self'",
  "form-action 'self'",
  "object-src 'none'",
].join("; ");

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // Linting runs once at the workspace root (ESLint 9 flat config); `next build` skips its own pass.
  eslint: { ignoreDuringBuilds: true },
  poweredByHeader: false,
  transpilePackages: ["@pdoom/schemas", "@pdoom/model-core", "@pdoom/sdk", "@pdoom/design-system"],
  outputFileTracingRoot: new URL("../../", import.meta.url).pathname,
  experimental: { optimizePackageImports: ["three", "@react-three/drei"] },
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          { key: "Content-Security-Policy", value: csp },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "no-referrer" },
          { key: "X-Frame-Options", value: "DENY" },
          { key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=(), interest-cohort=()" },
        ],
      },
    ];
  },
};

export default nextConfig;
