// Copyright NU Cybernetics. p(DOOM) — research prototype.
import Link from "next/link";

export default function NotFound() {
  return (
    <div className="page">
      <div className="container stack" style={{ maxWidth: "60ch" }}>
        <h1>Not found</h1>
        <p className="lede">There is no page at this address. Nothing is hidden: everything published is reachable from the pages below.</p>
        <ul>
          <li>
            <Link href="/meter">The meter</Link>
          </li>
          <li>
            <Link href="/text">Plain-text observatory</Link>
          </li>
          <li>
            <Link href="/method">Methodology</Link>
          </li>
        </ul>
      </div>
    </div>
  );
}
