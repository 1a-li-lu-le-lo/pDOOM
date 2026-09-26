// Copyright NU Cybernetics. PDUM — research prototype.
import type { ReactNode } from "react";

export const metadata = { title: "PDUM", description: "AI Existential and Civilizational Risk Observatory" };

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
