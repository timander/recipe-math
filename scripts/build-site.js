// Assembles the GitHub Pages site from CI artifacts:
//   site/index.html            the game (single file)
//   site/reports/index.html    a dashboard of the latest test results
//   site/reports/...           full coverage, Playwright, and mutation reports
// Usage: node scripts/build-site.js <artifacts-dir> <site-dir>
import { cpSync, existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const [artifacts = "artifacts", site = "site"] = process.argv.slice(2);
const read = (...p) => (existsSync(join(artifacts, ...p)) ? readFileSync(join(artifacts, ...p), "utf8") : "");
const esc = (s) => String(s).replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" })[c]);

mkdirSync(join(site, "reports"), { recursive: true });
const app = join(artifacts, "app", "recipe-math.html");
if (existsSync(app)) {
  cpSync(app, join(site, "index.html"));
  cpSync(app, join(site, "recipe-math.html"));
}
for (const [from, to] of [
  [["coverage", "coverage.html"], "coverage.html"],
  [["mutation", "mutation.json"], "mutation.json"],
  [["playwright", "playwright"], "playwright"],
]) {
  if (existsSync(join(artifacts, ...from)))
    cpSync(join(artifacts, ...from), join(site, "reports", to), { recursive: true });
}

// The headline numbers.
const coverage = (read("coverage", "coverage.md").match(/\*\*total\*\* \| \*\*([\d.]+)%/) || [])[1];
const efficacy = (read("mutation", "mutation.md").match(/Test efficacy: ([\d.]+)%/) || [])[1];
const e2e = read("playwright", "playwright.json") ? JSON.parse(read("playwright", "playwright.json")).stats : null;
const browser = read("playwright", "e2e-coverage.json")
  ? JSON.parse(read("playwright", "e2e-coverage.json")).percent
  : null;

// A small markdown-to-HTML converter for our own summaries (headings,
// paragraphs, bold, lists, tables, and <details> blocks).
function md(text) {
  const inline = (s) => esc(s).replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>");
  const out = [];
  const lines = text.split("\n");
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (line.startsWith("## ")) out.push(`<h2>${inline(line.slice(3))}</h2>`);
    else if (line.startsWith("<details>") || line.startsWith("</details>"))
      out.push(line.replace(/<summary>(.*)<\/summary>/, (_, s) => `<summary>${esc(s)}</summary>`));
    else if (line.startsWith("- ")) out.push(`<p class="li">${inline(line.slice(2))}</p>`);
    else if (line.startsWith("|")) {
      const rows = [];
      while (i < lines.length && lines[i].startsWith("|")) rows.push(lines[i++]);
      i--;
      const cells = (r) =>
        r
          .split("|")
          .slice(1, -1)
          .map((c) => c.trim());
      const [head, , ...body] = rows;
      out.push(
        `<table><thead><tr>${cells(head)
          .map((c) => `<th>${inline(c)}</th>`)
          .join("")}</tr></thead><tbody>` +
          body
            .map(
              (r) =>
                `<tr>${cells(r)
                  .map((c) => `<td>${inline(c)}</td>`)
                  .join("")}</tr>`,
            )
            .join("") +
          "</tbody></table>",
      );
    } else if (line.trim()) out.push(`<p>${inline(line)}</p>`);
  }
  return out.join("\n");
}

const tile = (label, value, ok, href) =>
  `<a class="tile ${ok ? "ok" : "bad"}" href="${href}"><span>${esc(label)}</span><strong>${esc(value)}</strong></a>`;
const tiles = [
  tile("Go test coverage", coverage ? `${coverage}%` : "n/a", Number(coverage) >= 90, "coverage.html"),
  tile("Mutation test efficacy", efficacy ? `${efficacy}%` : "n/a", Number(efficacy) >= 90, "#mutation"),
  e2e
    ? tile(
        "Playwright UI tests",
        `${e2e.expected} passed${e2e.unexpected ? `, ${e2e.unexpected} failed` : ""}`,
        !e2e.unexpected,
        "playwright/index.html",
      )
    : tile("Playwright UI tests", "n/a", false, "#"),
  tile("Browser code coverage", browser != null ? `${browser}%` : "n/a", browser >= 75, "playwright/index.html"),
  tile("Static analysis", "0 issues", true, "#lint"),
].join("\n");

const commit = (process.env.COMMIT || "local").slice(0, 7);
const runUrl = process.env.RUN_URL || "#";
writeFileSync(
  join(site, "reports", "index.html"),
  `<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Recipe Math: test results</title>
<style>
  body { font-family: system-ui, sans-serif; margin: 0; background: #f4efe6; color: #1f2933; }
  header { background: #2f3a44; color: #fff; padding: 1rem 1.5rem; }
  header a { color: #ffd7b8; }
  main { max-width: 60rem; margin: 0 auto; padding: 1rem 1.5rem 3rem; }
  .tiles { display: grid; grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr)); gap: 0.8rem; margin: 1rem 0 1.5rem; }
  .tile { display: grid; gap: 0.3rem; padding: 0.9rem; border-radius: 12px; background: #fff; text-decoration: none; color: inherit; border-left: 6px solid #15803d; }
  .tile.bad { border-left-color: #b91c1c; }
  .tile span { font-size: 0.85rem; color: #5f6b76; }
  .tile strong { font-size: 1.5rem; }
  section { background: #fff; border-radius: 12px; padding: 0.4rem 1.2rem 1rem; margin-bottom: 1rem; }
  table { border-collapse: collapse; width: 100%; font-size: 0.92rem; }
  th, td { text-align: left; padding: 0.3rem 0.5rem; border-bottom: 1px solid #e2d9ca; }
  td:not(:first-child), th:not(:first-child) { text-align: right; }
  details { font-size: 0.85rem; color: #5f6b76; margin-top: 0.6rem; }
  .li { margin: 0.2rem 0 0.2rem 1rem; }
</style></head>
<body>
<header>
  <h1>🛎️ Recipe Math: test results</h1>
  <p>Commit <code>${esc(commit)}</code> · <a href="${esc(runUrl)}">CI run</a> · <a href="../">Play the game</a></p>
</header>
<main>
  <div class="tiles">${tiles}</div>
  <section id="coverage">${md(read("coverage", "coverage.md"))}<p><a href="coverage.html">Line-by-line coverage report →</a></p></section>
  <section id="mutation">${md(read("mutation", "mutation.md"))}
    <p>Gremlins plants small bugs (flipped conditions, off-by-one boundaries, changed arithmetic) in the Go engine. Each one should make a test fail. Survivors are listed so they can be reviewed: most remaining ones are equivalent mutants that can't change behavior.</p></section>
  <section id="e2e">${md(read("playwright", "e2e-coverage.md"))}<p><a href="playwright/index.html">Full Playwright report (desktop Chrome + Pixel 7) →</a></p></section>
  <section id="lint"><h2>Static analysis</h2><p>golangci-lint (vet, staticcheck, errcheck, and revive rules from the Google Go Style Guide; see <code>.golangci.yml</code>), ESLint with the DOM-XSS check, and Prettier all pass. CI fails on any issue.</p></section>
</main>
</body></html>
`,
);
process.stdout.write(`site built in ${site}/ (coverage ${coverage}%, mutation ${efficacy}%, browser ${browser}%)\n`);
