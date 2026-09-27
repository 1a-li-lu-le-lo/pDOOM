// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { MetadataRoute } from "next";

const base = process.env.PDOOM_PUBLIC_URL ?? "http://localhost:3000";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [{ userAgent: "*", allow: "/", disallow: ["/api/"] }],
    sitemap: `${base}/sitemap.xml`,
  };
}
