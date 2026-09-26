// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import type { ReactNode } from "react";
import { BRAND, PUBLIC_LABEL, TAGLINES } from "@pdoom/schemas";
import "./globals.css";
import { ModeProvider } from "@/components/mode/ModeProvider";
import { Header } from "@/components/shell/Header";
import { Footer } from "@/components/shell/Footer";
import { getRelease } from "@/lib/data";

export const metadata: Metadata = {
  title: { default: `${BRAND} — ${PUBLIC_LABEL}`, template: `%s · ${BRAND}` },
  description: `${TAGLINES[0]} ${BRAND} monitors evidence relevant to catastrophic and existential risk from advanced AI, with every number shown with its horizon, outcome, interval and sources.`,
  robots: { index: true, follow: true },
  openGraph: { title: BRAND, description: TAGLINES[0], type: "website" },
};

export default async function RootLayout({ children }: { children: ReactNode }) {
  let releaseId: string | undefined;
  let cutoff: string | undefined;
  let published: string | null = null;
  try {
    const rel = await getRelease();
    releaseId = rel.manifest.release_id;
    cutoff = rel.manifest.source_cutoff;
    published = rel.manifest.published;
  } catch {
    /* no release: the footer says so */
  }
  return (
    <html lang="en">
      <body>
        <a className="skip-link" href="#main">
          Skip to content
        </a>
        <ModeProvider>
          <Header />
          <main id="main">{children}</main>
          <Footer releaseId={releaseId} dataCutoff={cutoff} published={published} />
        </ModeProvider>
      </body>
    </html>
  );
}
