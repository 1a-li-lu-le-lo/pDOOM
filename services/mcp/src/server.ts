#!/usr/bin/env tsx
// Copyright NU Cybernetics. p(DOOM) — research prototype.
/**
 * p(DOOM) MCP server (stdio). Read tools over the promoted release and its
 * snapshot, plus three submission tools that append to the human review queue.
 * Configuration: PDOOM_DATA_DIR (file mode, default: the repository's data/)
 * or PDOOM_API_URL (http mode against cmd/pdoom-api).
 */
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { createDefaultDataSource, resolveDataDir } from "@pdoom/sdk";
import { BRAND } from "@pdoom/schemas";
import { registerTools, resolveContext } from "./tools";

export function createServer(env: NodeJS.ProcessEnv = process.env) {
  const server = new McpServer(
    { name: "pdoom", version: "0.1.0", title: `${BRAND} observatory` },
    {
      instructions: [
        `${BRAND} publishes eight separate outputs that are never blended: external forecast aggregates, a research-mode model estimate, five indexes and an editorial level.`,
        "The official p(DOOM) value is withheld (insufficiently_calibrated) in this release line; never present an index or a research-mode number as an official probability.",
        "Quote every probability with its horizon, outcome set (O3–O8 for p(DOOM)), status and model version, rounded as the release displays it. Do not imply a date or inevitability.",
        "Submission tools only append to a human review queue; they cannot change published data.",
      ].join(" "),
    },
  );
  const source = createDefaultDataSource();
  const ctx = resolveContext({ ...env, PDOOM_DATA_DIR: env.PDOOM_DATA_DIR ?? resolveDataDir() }, source);
  registerTools(server, ctx);
  return server;
}

const isMain = process.argv[1] && /server\.(ts|js)$/.test(process.argv[1]);
if (isMain) {
  const server = createServer();
  const transport = new StdioServerTransport();
  server.connect(transport).catch((err: unknown) => {
    process.stderr.write(`pdoom-mcp: ${err instanceof Error ? err.message : String(err)}\n`);
    process.exit(1);
  });
}
