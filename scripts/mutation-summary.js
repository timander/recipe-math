// Turns Gremlins' JSON results into a markdown summary (for the CI job
// summary and the published report page).
import { readFileSync } from "node:fs";

const data = JSON.parse(readFileSync(process.argv[2], "utf8"));
const byFile = new Map();
const totals = { KILLED: 0, LIVED: 0, "NOT COVERED": 0, "TIMED OUT": 0 };
for (const f of data.files) {
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
const lines = [
  "## Mutation testing (Gremlins)",
  "",
  `**Test efficacy: ${pct(totals.KILLED, tested)}%** (${totals.KILLED} of ${tested} planted bugs caught) · ` +
    `mutant coverage ${pct(tested, tested + totals["NOT COVERED"])}% · timed out ${totals["TIMED OUT"]}`,
  "",
  "| File | Killed | Lived | Not covered | Efficacy |",
  "|---|---:|---:|---:|---:|",
];
for (const [file, r] of [...byFile].sort()) {
  lines.push(`| ${file} | ${r.KILLED} | ${r.LIVED} | ${r["NOT COVERED"]} | ${pct(r.KILLED, r.KILLED + r.LIVED)}% |`);
}
lines.push("", "<details><summary>Surviving mutants</summary>", "");
for (const [file, r] of [...byFile].sort()) {
  if (r.lived.length) lines.push(`- **${file}**: ${r.lived.join(", ")}`);
}
lines.push("", "</details>");
process.stdout.write(lines.join("\n") + "\n");
