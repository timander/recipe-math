// Turns Gremlins' JSON results into a markdown summary (for the CI job
// summary and the published report page). CI runs Gremlins in parallel
// shards, so this also merges several result files into one.
//
//   node scripts/mutation-summary.js [--threshold 90] [--merge all.json] a.json [b.json ...]
//
// Exits 1 if test efficacy is under the threshold.
import { readFileSync, writeFileSync } from "node:fs";
import { parseArgs } from "node:util";

const { values: opts, positionals: inputs } = parseArgs({
  allowPositionals: true,
  options: { threshold: { type: "string", default: "0" }, merge: { type: "string" } },
});
const results = inputs.map((path) => JSON.parse(readFileSync(path, "utf8")));
const files = results.flatMap((r) => r.files ?? []);
const elapsed = Math.max(0, ...results.map((r) => r.elapsed_time ?? 0));

const byFile = new Map();
const totals = { KILLED: 0, LIVED: 0, "NOT COVERED": 0, "TIMED OUT": 0 };
for (const f of files) {
  const row = byFile.get(f.file_name) ?? { KILLED: 0, LIVED: 0, "NOT COVERED": 0, "TIMED OUT": 0, lived: [] };
  for (const m of f.mutations) {
    if (m.status in row) row[m.status]++;
    if (m.status in totals) totals[m.status]++;
    if (m.status === "LIVED") row.lived.push(`${m.line}:${m.column} ${m.type}`);
  }
  byFile.set(f.file_name, row);
}
const pct = (a, b) => (b ? ((100 * a) / b).toFixed(1) : "0.0");
const tested = totals.KILLED + totals.LIVED;
const efficacy = tested ? (100 * totals.KILLED) / tested : 0;

if (opts.merge) {
  const merged = {
    ...results[0],
    files,
    test_efficacy: efficacy,
    mutants_total: tested,
    mutants_killed: totals.KILLED,
    mutants_lived: totals.LIVED,
    mutants_not_covered: totals["NOT COVERED"],
    elapsed_time: elapsed,
  };
  delete merged.mutations_coverage;
  delete merged.mutator_statistics;
  writeFileSync(opts.merge, JSON.stringify(merged, null, 2) + "\n");
}

const lines = [
  "## Mutation testing (Gremlins)",
  "",
  `**Test efficacy: ${pct(totals.KILLED, tested)}%** (${totals.KILLED} of ${tested} planted bugs caught) · ` +
    `mutant coverage ${pct(tested, tested + totals["NOT COVERED"])}% · timed out ${totals["TIMED OUT"]}` +
    (inputs.length > 1 ? ` · ${inputs.length} parallel shards, slowest ${Math.round(elapsed)}s` : ""),
  "",
  "| File | Killed | Lived | Not covered | Timed out | Efficacy |",
  "|---|---:|---:|---:|---:|---:|",
];
for (const [file, r] of [...byFile].sort()) {
  lines.push(
    `| ${file} | ${r.KILLED} | ${r.LIVED} | ${r["NOT COVERED"]} | ${r["TIMED OUT"]} | ${pct(r.KILLED, r.KILLED + r.LIVED)}% |`,
  );
}
lines.push("", "<details><summary>Surviving mutants</summary>", "");
for (const [file, r] of [...byFile].sort()) {
  if (r.lived.length) lines.push(`- **${file}**: ${r.lived.join(", ")}`);
}
lines.push("", "</details>");

const threshold = Number(opts.threshold);
if (efficacy < threshold) {
  lines.push("", `**Failed:** efficacy ${efficacy.toFixed(1)}% is under the ${threshold}% threshold.`);
}
process.stdout.write(lines.join("\n") + "\n");
process.exitCode = efficacy < threshold ? 1 : 0;
