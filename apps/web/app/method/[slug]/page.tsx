// Copyright NU Cybernetics. p(DOOM) — research prototype.
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { PageHeader } from "@/components/shell/PageHeader";
import { getDataSource, getSnapshot } from "@/lib/data";
import { renderMarkdown } from "@/lib/markdown";

export async function generateStaticParams() {
  const docs = await getDataSource().getMethodology();
  return [...docs.map((d) => ({ slug: d.slug })), { slug: "definitions" }];
}

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const { slug } = await params;
  if (slug === "definitions") return { title: "Definitions", description: "Every term used on the site, with sourced definitions and the working definition adopted here." };
  const doc = await getDataSource().getMethodologyDoc(slug);
  return { title: doc?.doc.title ?? "Method", description: doc ? `Methodology document ${doc.doc.path}` : undefined };
}

export default async function MethodDocPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  if (slug === "definitions") return <Definitions />;
  if (!/^[a-z0-9-]+$/.test(slug)) notFound();
  const doc = await getDataSource().getMethodologyDoc(slug);
  if (!doc) notFound();
  const r = renderMarkdown(doc.markdown);
  const toc = r.headings.filter((h) => h.depth === 2);
  return (
    <div className="page">
      <div className="container">
        <PageHeader crumbs={[{ href: "/method", label: "Method" }]} title={r.title || doc.doc.title} textAnchor="method">
          <p className="cite">{doc.doc.path}</p>
        </PageHeader>
        <div className="doc-layout">
          {toc.length > 1 ? (
            <nav className="doc-toc" aria-label="On this page">
              <div className="eyebrow">On this page</div>
              <ol>
                {toc.map((h) => (
                  <li key={h.id}>
                    <a href={`#${h.id}`}>{h.text}</a>
                  </li>
                ))}
              </ol>
            </nav>
          ) : null}
          <article className="prose" dangerouslySetInnerHTML={{ __html: r.html.replace(/^<h1[^>]*>[\s\S]*?<\/h1>\n?/, "") }} />
        </div>
      </div>
    </div>
  );
}

async function Definitions() {
  const snap = await getSnapshot();
  const defs = [...snap.definitions].sort((a, b) => a.term.localeCompare(b.term));
  const src = new Map(snap.sources.map((s) => [s.id, s]));
  return (
    <div className="page">
      <div className="container stack">
        <PageHeader crumbs={[{ href: "/method", label: "Method" }]} title="Definitions" lede="Each term shows the definitions found in the literature, with attribution, and the working definition adopted here. Where sources disagree, the disagreement is shown rather than resolved silently." textAnchor="method" />
        <div className="stack">
          {defs.map((d) => (
            <article key={d.id} id={d.id} className="card stack">
              <h2 style={{ margin: 0 }}>{d.term}</h2>
              <p>
                <strong>Working definition.</strong> {d.short_definition}
              </p>
              {d.definitions.length ? (
                <details>
                  <summary>{d.definitions.length} definitions in the literature</summary>
                  <ul>
                    {d.definitions.map((x, i) => (
                      <li key={i}>
                        “{x.text}” — {x.attribution}
                        {x.source_id ? (
                          <>
                            {" "}
                            (<Link href={`/evidence/sources/${x.source_id}`}>{src.get(x.source_id)?.title ?? x.source_id}</Link>)
                          </>
                        ) : null}
                        {x.note ? <span className="cite"> {x.note}</span> : null}
                      </li>
                    ))}
                  </ul>
                </details>
              ) : null}
              <p className="cite">Consensus: {d.consensus.replace(/_/g, " ")}</p>
              {d.see_also_urls.length ? (
                <p className="cite">
                  See also:{" "}
                  {d.see_also_urls.map((u, i) => (
                    <span key={u}>
                      {i ? ", " : ""}
                      <a href={u} rel="noopener noreferrer">
                        {new URL(u).hostname}
                      </a>
                    </span>
                  ))}
                </p>
              ) : null}
              {d.related_ids.length ? (
                <p className="cite">
                  Related:{" "}
                  {d.related_ids.map((r, i) => (
                    <span key={r}>
                      {i ? ", " : ""}
                      <a href={`#${r}`}>{r.replace(/^def-/, "")}</a>
                    </span>
                  ))}
                </p>
              ) : null}
            </article>
          ))}
        </div>
      </div>
    </div>
  );
}
