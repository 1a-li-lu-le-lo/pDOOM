// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { ReactNode } from "react";

export const metadata = { title: "P_DOOM", description: "AI Existential and Civilizational Risk Observatory" };

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
