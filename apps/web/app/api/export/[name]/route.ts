// Copyright NU Cybernetics. p(DOOM) — research prototype.
// Machine-readable exports of the promoted release and its snapshot. Every
// file is generated from the same objects the pages render; nothing is
// recomputed here. CSV cells are quoted and formula-injection-safe.
import { NextResponse } from "next/server";
import { getDataSource, getRelease, getSnapshot } from "@/lib/data";

export const dynamic = "force-dynamic";

const NAMES = ["release.json", "estimates.csv", "sources.csv", "forecasts.csv", "incidents.csv", "snapshot.json", "estimates.jsonl"] as const;
type Name = (typeof NAMES)[number];

function csvCell(v: unknown): string {
  if (v === null || v === undefined) return "";
  let s = Array.isArray(v) ? v.join("|") : typeof v === "object" ? JSON.stringify(v) : String(v);
  if (/^[=+\-@\t\r]/.test(s)) s = `'${s}`;
  return /[",\n\r]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s;
}

function csv(rows: Record<string, unknown>[], columns: string[]): string {
  const head = columns.map(csvCell).join(",");
  const body = rows.map((r) => columns.map((c) => csvCell(r[c])).join(","));
  return [head, ...body].join("\r\n") + "\r\n";
}

export async function generateStaticParams() {
  return NAMES.map((name) => ({ name }));
}

export async function GET(req: Request, ctx: { params: Promise<{ name: string }> }) {
  const { name } = await ctx.params;
  if (!(NAMES as readonly string[]).includes(name)) return NextResponse.json({ error: "unknown export", available: NAMES }, { status: 404 });
  const url = new URL(req.url);
  const releaseParam = url.searchParams.get("release");
  const rel = releaseParam && /^rel-\d{4}-\d{2}-\d{2}-\d{3}$/.test(releaseParam) ? await getDataSource().getReleaseById(releaseParam).catch(() => null) : await getRelease();
  if (!rel) return NextResponse.json({ error: "release not found" }, { status: 404 });
  const headers = (type: string, filename: string) => ({
    "content-type": type,
    "content-disposition": `attachment; filename="pdoom-${rel.manifest.release_id}-${filename}"`,
    "cache-control": "public, max-age=300",
    "x-pdoom-release": rel.manifest.release_id,
  });
  switch (name as Name) {
    case "release.json":
      return new NextResponse(JSON.stringify({ manifest: rel.manifest, estimates: rel.estimates, indexes: rel.indexes, aggregations: rel.aggregations, sensitivity: rel.sensitivity, delta: rel.delta, drivers_explained: rel.drivers_explained, approvals: rel.approvals }, null, 2), { headers: headers("application/json; charset=utf-8", "release.json") });
    case "estimates.jsonl":
      return new NextResponse(rel.estimates.map((e) => JSON.stringify(e)).join("\n") + "\n", { headers: headers("application/x-ndjson; charset=utf-8", "estimates.jsonl") });
    case "estimates.csv":
      return new NextResponse(
        csv(
          rel.estimates.map((e) => ({
            estimate_id: e.estimate_id,
            status: e.status,
            outcome_set: e.outcome_set,
            horizon: e.horizon,
            producer: e.producer,
            p05: e.quantiles?.p05 ?? "",
            p25: e.quantiles?.p25 ?? "",
            p50: e.quantiles?.p50 ?? "",
            p75: e.quantiles?.p75 ?? "",
            p95: e.quantiles?.p95 ?? "",
            mean: e.mean ?? "",
            display_central: e.display.central,
            display_interval: e.display.interval,
            uncertainty: e.uncertainty,
            disagreement: e.disagreement ?? "",
            rounding_rule: e.rounding_rule,
            forecast_origin_date: e.forecast_origin_date,
            last_evidence_date: e.last_evidence_date,
            conditioning: e.conditioning,
            method_ref: e.method_ref,
            source_ids: e.source_coverage.source_ids,
          })),
          ["estimate_id", "status", "outcome_set", "horizon", "producer", "p05", "p25", "p50", "p75", "p95", "mean", "display_central", "display_interval", "uncertainty", "disagreement", "rounding_rule", "forecast_origin_date", "last_evidence_date", "conditioning", "method_ref", "source_ids"],
        ),
        { headers: headers("text/csv; charset=utf-8", "estimates.csv") },
      );
    default: {
      const snap = await getSnapshot();
      if (name === "snapshot.json") {
        return new NextResponse(JSON.stringify(snap, null, 2), { headers: headers("application/json; charset=utf-8", "snapshot.json") });
      }
      if (name === "sources.csv") {
        return new NextResponse(
          csv(
            snap.sources.map((s) => ({ id: s.id, title: s.title, publisher: s.publisher, authors: s.authors, date_published: s.date_published, source_tier: s.source_tier, source_type: s.source_type, topic: s.topic, canonical_url: s.canonical_url, license: s.license, robots_status: s.robots_status, verification_status: s.verification.status, verification_checked_at: s.verification.checked_at, human_review_status: s.human_review_status, model_use_status: s.model_use_status, retraction_status: s.retraction_status, conflicts: s.conflicts, citation: s.citation })),
            ["id", "title", "publisher", "authors", "date_published", "source_tier", "source_type", "topic", "canonical_url", "license", "robots_status", "verification_status", "verification_checked_at", "human_review_status", "model_use_status", "retraction_status", "conflicts", "citation"],
          ),
          { headers: headers("text/csv; charset=utf-8", "sources.csv") },
        );
      }
      if (name === "forecasts.csv") {
        return new NextResponse(
          csv(
            snap.forecasts.map((f) => ({ id: f.id, forecaster_or_survey: f.forecaster_or_survey, source_id: f.source_id, date: f.date, population: f.population, sample_size: f.sample_size, outcome_set: f.outcome_set, horizon: f.horizon, horizon_note: f.horizon_note, horizon_end_year: f.horizon_end_year, conditions: f.conditions, mean: f.mean, median: f.median, p05: f.quantiles?.p05, p25: f.quantiles?.p25, p75: f.quantiles?.p75, p95: f.quantiles?.p95, group_id: f.group_id, question_wording_original: f.question_wording_original, paraphrase: f.paraphrase, status: f.status, verification_status: f.verification.status, model_use_status: f.model_use_status })),
            ["id", "forecaster_or_survey", "source_id", "date", "population", "sample_size", "outcome_set", "horizon", "horizon_note", "horizon_end_year", "conditions", "mean", "median", "p05", "p25", "p75", "p95", "group_id", "question_wording_original", "paraphrase", "status", "verification_status", "model_use_status"],
          ),
          { headers: headers("text/csv; charset=utf-8", "forecasts.csv") },
        );
      }
      return new NextResponse(
        csv(
          snap.incidents.map((i) => ({ id: i.id, title: i.title, date: i.date, date_precision: i.date_precision, severity: i.severity, pdoom_relevance: i.pdoom_relevance, evidence_level: i.evidence_level, near_miss: i.near_miss, novelty: i.novelty, cause: i.cause, harm: i.harm, systems_involved: i.systems_involved, jurisdiction: i.jurisdiction, external_ids: i.external_ids, source_ids: i.source_ids, summary: i.summary, verification_status: i.verification.status, model_use_status: i.model_use_status })),
          ["id", "title", "date", "date_precision", "severity", "pdoom_relevance", "evidence_level", "near_miss", "novelty", "cause", "harm", "systems_involved", "jurisdiction", "external_ids", "source_ids", "summary", "verification_status", "model_use_status"],
        ),
        { headers: headers("text/csv; charset=utf-8", "incidents.csv") },
      );
    }
  }
}
