// Merges the per-test V8 coverage of the page's inline scripts into one
// number for web/app.js (bytes of code that ran at least once), and writes
// reports/e2e-coverage.json and reports/e2e-coverage.md.
import { existsSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");

// rangeHits maps V8's nested ranges onto app.js offsets: outer function
// ranges count 1+, inner blocks that never ran count 0, and later (inner)
// ranges override earlier (outer) ones.
function rangeHits(entry, start, length) {
  const hits = new Int32Array(length).fill(-1);
  for (const fn of entry.functions) {
    for (const r of fn.ranges) {
      hits.fill(r.count, Math.max(r.startOffset - start, 0), Math.max(Math.min(r.endOffset - start, length), 0));
    }
  }
  return hits;
}

export default function teardown() {
  const dir = join(root, "reports", "e2e-coverage-raw");
  if (!existsSync(dir)) return;
  const app = readFileSync(join(root, "web", "app.js"), "utf8");
  let covered = null;
  for (const file of readdirSync(dir)) {
    for (const entry of JSON.parse(readFileSync(join(dir, file), "utf8"))) {
      // The bundle inlines several scripts; app.js is the one containing its header.
      const start = entry.source.indexOf(app.slice(0, 200));
      if (start < 0) continue;
      covered ??= new Uint8Array(app.length);
      const hits = rangeHits(entry, start, app.length);
      for (let i = 0; i < app.length; i++) if (hits[i] > 0) covered[i] = 1;
    }
  }
  if (!covered) return;
  // Count only non-whitespace characters so blank lines and indentation don't skew it.
  let total = 0;
  let ran = 0;
  for (let i = 0; i < app.length; i++) {
    if (/\s/.test(app[i])) continue;
    total++;
    ran += covered[i];
  }
  const pct = ((100 * ran) / total).toFixed(1);
  writeFileSync(
    join(root, "reports", "e2e-coverage.json"),
    JSON.stringify({ file: "web/app.js", percent: Number(pct) }),
  );
  writeFileSync(
    join(root, "reports", "e2e-coverage.md"),
    `## Browser code coverage (Playwright, Chromium)\n\n**web/app.js: ${pct}%** of code ran during the UI tests.\n`,
  );
}
