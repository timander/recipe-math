// Recipe Math UI. All the math lives in the Go engine (window.recipeMath);
// this file only draws the screens and the printable pages.
"use strict";

const ROUND_SIZE = 10;
const STORE_KEY = "recipeMath.v1";
const $ = (id) => document.getElementById(id);

// el("p", {class: "x"}, "text", child) builds DOM without innerHTML.
function el(tag, attrs, ...children) {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs || {})) {
    if (v === false || v == null) continue;
    if (k === "class") node.className = v;
    else node.setAttribute(k, v === true ? "" : v);
  }
  for (const c of children.flat()) {
    if (c != null && c !== false) node.append(c);
  }
  return node;
}

function call(name, ...args) {
  const out = JSON.parse(window.recipeMath[name](...args));
  if (out && out.error) throw new Error(out.error);
  return out;
}

const store = {
  data: (() => {
    try { return JSON.parse(localStorage.getItem(STORE_KEY)) || {}; } catch { return {}; }
  })(),
  get(key, fallback) { return this.data[key] ?? fallback; },
  set(key, value) {
    this.data[key] = value;
    try { localStorage.setItem(STORE_KEY, JSON.stringify(this.data)); } catch { /* private mode */ }
  },
};

const randomSeed = () => crypto.getRandomValues(new Uint32Array(1))[0];
const randomVersion = () => 1000 + (randomSeed() % 9000);
const starText = (n) => "★".repeat(n) + "☆".repeat(3 - n);
const formatPoints = (p) => (p % 1 ? `${Math.floor(p)}½` : String(p)).replace(/^0½$/, "½");

let LEVELS = [];
let REF = null;

// ---------- Loading the Go engine ----------

async function loadEngine() {
  const bin = atob(WASM_GZIP_BASE64);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
  const stream = new Blob([bytes]).stream().pipeThrough(new DecompressionStream("gzip"));
  const wasm = await new Response(stream).arrayBuffer();
  const go = new Go();
  const ready = new Promise((resolve) => { window.recipeMathReady = resolve; });
  const { instance } = await WebAssembly.instantiate(wasm, go.importObject);
  go.run(instance);
  await ready;
}

// ---------- Tabs ----------

function showTab(name) {
  for (const btn of document.querySelectorAll("[role=tab]")) {
    const on = btn.dataset.tab === name;
    btn.setAttribute("aria-selected", String(on));
    $(btn.dataset.tab).hidden = !on;
  }
  store.set("tab", name);
}

// ---------- Practice ----------

const round = { level: null, problems: [], i: 0, tries: 0, hinted: false, done: false, results: [] };

function renderLevelList() {
  const best = store.get("best", {});
  $("level-list").replaceChildren(...LEVELS.map((l) =>
    el("li", {},
      el("button", { type: "button", class: "level-card", "data-level": l.id },
        el("span", { class: "level-num", "aria-hidden": "true" }, String(l.id)),
        el("span", { class: "level-text" },
          el("strong", {}, `Level ${l.id}: ${l.name}`),
          el("span", { class: "level-goal" }, l.goal),
          el("span", { class: "level-example" }, l.example)),
        el("span", { class: "level-stars", "aria-label": `Best: ${best[l.id] || 0} of 3 stars` }, starText(best[l.id] || 0))))));
}

function startRound(levelId) {
  const level = LEVELS.find((l) => l.id === levelId);
  Object.assign(round, {
    level,
    problems: call("generate", JSON.stringify([levelId]), ROUND_SIZE, randomSeed()).problems,
    i: 0, results: [],
  });
  $("level-picker").hidden = true;
  $("summary").hidden = true;
  $("round").hidden = false;
  $("round-level").textContent = `Level ${level.id}: ${level.name}`;
  showProblem();
}

function showProblem() {
  const p = round.problems[round.i];
  Object.assign(round, { tries: 0, hinted: false, done: false });
  $("round-progress").textContent = `${round.i + 1} of ${round.problems.length}`;
  $("given").textContent = p.givenText;
  $("equals").textContent = p.equals;
  $("to-label").textContent = p.toLabel;
  const input = $("answer");
  input.value = "";
  input.disabled = false;
  input.removeAttribute("aria-invalid");
  setFeedback("", "");
  $("work").hidden = true;
  $("check-btn").hidden = false;
  $("hint-btn").hidden = false;
  $("hint-btn").disabled = false;
  $("next-btn").hidden = true;
  renderDots();
  // Don't pop up the phone keyboard over the problem on the first screen.
  if (round.i > 0 || matchMedia("(pointer: fine)").matches) input.focus();
}

function renderDots() {
  $("dots").replaceChildren(...round.problems.map((_, i) => {
    const r = round.results[i];
    const state = r ? (r.points === 1 ? "full" : r.points > 0 ? "half" : "miss") : i === round.i ? "now" : "todo";
    const label = { full: "correct", half: "correct with help", miss: "missed", now: "current", todo: "not yet" }[state];
    return el("li", { class: `dot ${state}`, "aria-label": `Problem ${i + 1}: ${label}` });
  }));
}

function setFeedback(kind, text) {
  const fb = $("feedback");
  fb.className = `feedback ${kind}`;
  fb.textContent = text;
}

function showWork(p) {
  $("steps").replaceChildren(...p.steps.map((s) => el("li", {}, s)));
  $("work").hidden = false;
}

function finishProblem(points) {
  const p = round.problems[round.i];
  round.done = true;
  round.results[round.i] = { points, problem: p };
  $("answer").disabled = true;
  $("check-btn").hidden = true;
  $("hint-btn").hidden = true;
  $("next-btn").hidden = false;
  $("next-btn").textContent = round.i + 1 < round.problems.length ? "Next →" : "See score →";
  $("next-btn").focus();
  showWork(p);
  renderDots();
}

function checkAnswer() {
  if (round.done) return nextProblem();
  const p = round.problems[round.i];
  const input = $("answer");
  const r = call("check", JSON.stringify(p), input.value);
  if (r.status === "invalid") {
    input.setAttribute("aria-invalid", "true");
    setFeedback("warn", r.message);
    input.focus();
    return;
  }
  input.removeAttribute("aria-invalid");
  if (r.status === "correct") {
    const points = round.tries === 0 && !round.hinted ? 1 : 0.5;
    setFeedback("good", points === 1 ? `✅ ${r.message}` : `✅ ${r.message} (half point for using help)`);
    return finishProblem(points);
  }
  round.tries++;
  if (round.tries === 1) {
    setFeedback("bad", `❌ ${r.message}`);
    input.select();
    return;
  }
  // Second miss: show the answer and the steps instead of "try again".
  const why = r.message.startsWith("Not quite") ? "Not quite." : r.message;
  setFeedback("bad", `❌ ${why} The answer is ${p.answerText}.`);
  finishProblem(0);
}

function showHint() {
  const p = round.problems[round.i];
  round.hinted = true;
  $("hint-btn").disabled = true;
  setFeedback("hint", `💡 ${p.hint}`);
  $("answer").focus();
}

function nextProblem() {
  round.i++;
  if (round.i < round.problems.length) return showProblem();
  showSummary();
}

function showSummary() {
  const total = round.results.reduce((s, r) => s + r.points, 0);
  const stars = total >= 9 ? 3 : total >= 7 ? 2 : total >= 5 ? 1 : 0;
  const best = store.get("best", {});
  if (stars > (best[round.level.id] || 0)) {
    best[round.level.id] = stars;
    store.set("best", best);
  }
  $("round").hidden = true;
  $("summary").hidden = false;
  $("summary-stars").textContent = starText(stars);
  $("summary-title").textContent = [
    "Keep practicing. Try a hint when you're stuck!",
    "Getting there!",
    "Nice work!",
    "Kitchen pro!",
  ][stars];
  $("summary-score").textContent = `You scored ${formatPoints(total)} out of ${round.problems.length}.`;
  const missed = round.results.filter((r) => r.points < 1);
  $("summary-missed").replaceChildren(missed.length === 0 ? "" :
    el("div", { class: "missed" },
      el("h3", {}, "Review these"),
      el("ul", {}, missed.map((r) => el("li", {}, `${r.problem.givenText} ${r.problem.equals} `, el("strong", {}, r.problem.answerText))))));
  const next = LEVELS.find((l) => l.id === round.level.id + 1);
  $("next-level-btn").hidden = !next;
  $("again-btn").focus();
}

function backToLevels() {
  $("round").hidden = true;
  $("summary").hidden = true;
  $("level-picker").hidden = false;
  renderLevelList();
}

// ---------- Worksheet ----------

function wsSettings() {
  return {
    levels: [...document.querySelectorAll("#ws-levels input:checked")].map((c) => Number(c.value)),
    count: Number($("ws-count").value),
    title: $("ws-title").value.trim() || "Kitchen Unit Conversions",
    version: Math.min(9999, Math.max(1000, Number($("ws-version").value) || 1000)),
    cheat: $("ws-cheat").checked,
    work: $("ws-work").checked,
  };
}

function saveWs() {
  const s = wsSettings();
  store.set("ws", { ...s, title: $("ws-title").value });
}

function renderWorksheet(mode) {
  mode = mode || document.querySelector("input[name=ws-preview]:checked").value;
  const s = wsSettings();
  const sheet = $("ws-sheet");
  sheet.classList.toggle("is-key", mode === "key");
  sheet.classList.toggle("dense", s.count >= 24);
  if (s.levels.length === 0) {
    sheet.replaceChildren(el("p", { class: "empty" }, "Pick at least one level."));
    return;
  }
  const problems = call("generate", JSON.stringify(s.levels), s.count, s.version).problems;
  const levelNames = s.levels.map((id) => `${id}`).join(", ");
  const hasApprox = problems.some((p) => p.approx);

  const head = el("header", { class: "sheet-head" },
    el("h2", {}, mode === "key" ? `${s.title}: Answer Key` : s.title,
      mode === "key" ? "" : el("small", { class: "version" }, `Version ${s.version}`)),
    mode === "key"
      ? el("p", { class: "meta" }, `Version ${s.version} · Level${s.levels.length > 1 ? "s" : ""} ${levelNames}`)
      : el("div", { class: "name-line" },
          el("span", {}, "Name ", el("i", { class: "blank wide" })),
          el("span", {}, "Date ", el("i", { class: "blank" })),
          el("span", {}, "Period ", el("i", { class: "blank short" }))));

  const instructions = el("p", { class: "instructions" },
    "Convert each amount. Write fractions like 1 1/2 or decimals like 0.75.",
    hasApprox ? " ≈ means “about”: US ↔ metric answers are rounded." : "");

  const groups = [];
  let n = 1;
  for (const id of s.levels) {
    const level = LEVELS.find((l) => l.id === id);
    const items = problems.filter((p) => p.level === id);
    if (items.length === 0) continue;
    groups.push(el("section", { class: "ws-group" },
      s.levels.length > 1 ? el("h3", {}, `Level ${id}: ${level.name}`) : "",
      el("ol", { class: "ws-list", start: String(n) }, items.map((p) =>
        el("li", {},
          el("span", { class: "ws-q" },
            `${p.givenText} ${p.equals} `,
            mode === "key" ? el("b", { class: "ans" }, p.answer) : el("i", { class: "blank ans-blank" }),
            ` ${p.toLabel}`),
          mode === "key" && s.work
            ? el("span", { class: "ws-work" }, p.steps.filter((t) => !/^An? /.test(t) && !t.startsWith("Shortcut")).join(" "))
            : "")))));
    n += items.length;
  }

  const footer = el("footer", { class: "sheet-foot" },
    s.cheat && mode !== "key" ? miniCheat(s.levels) : "");

  sheet.replaceChildren(head, instructions, ...groups, footer);
}

// miniCheat lists only the facts the chosen levels use.
function miniCheat(levelIds) {
  const wantApprox = levelIds.includes(6);
  const units = new Set();
  const uses = {
    1: ["Tbsp", "tsp"], 2: ["cup", "Tbsp", "floz"], 3: ["cup", "pint", "quart", "gallon"],
    4: ["lb", "oz", "L", "mL", "kg", "g"], 5: ["tsp", "Tbsp", "floz", "cup", "pint", "quart", "gallon"],
    6: ["tsp", "Tbsp", "floz", "cup", "oz", "lb", "kg", "mL", "g"],
  };
  levelIds.forEach((id) => (uses[id] || []).forEach((u) => units.add(u)));
  const facts = REF.facts.filter((f) => units.has(f.big) && units.has(f.small) && (!f.approx || wantApprox));
  return el("aside", { class: "mini-cheat" },
    el("strong", {}, "Cheat sheet: "),
    facts.map((f) => el("span", { class: "fact" }, f.text)));
}

function printSheet(mode) {
  saveWs();
  renderWorksheet(mode);
  const restore = () => { renderWorksheet(); window.removeEventListener("afterprint", restore); };
  window.addEventListener("afterprint", restore);
  window.print();
}

// ---------- Cheat sheet ----------

function renderCheatSheet() {
  const facts = (pred) => el("ul", { class: "fact-list" }, REF.facts.filter(pred).map((f) => el("li", {}, f.text)));
  const isUS = (id) => REF.units.find((u) => u.id === id).system === "US";
  const vol = (id) => REF.units.find((u) => u.id === id).dimension === "volume";
  const unitRows = REF.units.map((u) =>
    el("tr", {},
      el("td", {}, u.name),
      el("td", {}, [u.abbr, ...(u.otherAbbr || [])].join(", ")),
      el("td", {}, u.dimension === "volume" ? "volume (space)" : "weight"),
      el("td", {}, u.tool)));

  $("cheat-sheet").replaceChildren(
    el("header", { class: "sheet-head" }, el("h2", {}, "Kitchen Units Cheat Sheet")),
    el("div", { class: "cheat-grid" },
      el("section", {},
        el("h3", {}, "The Gallon Kingdom"),
        gallonDiagram(),
        el("p", { class: "caption" }, "1 gallon = 4 quarts = 8 pints = 16 cups")),
      el("section", {},
        el("h3", {}, "Spoons & cups (exact)"),
        facts((f) => !f.approx && isUS(f.big) && vol(f.big)),
        el("h3", {}, "Weight (exact)"),
        facts((f) => !f.approx && isUS(f.big) && !vol(f.big)),
        el("h3", {}, "Metric (exact)"),
        facts((f) => !f.approx && !isUS(f.big)),
        el("h3", {}, "US ↔ metric (≈ about)"),
        facts((f) => f.approx))),
    el("h3", {}, "Remember"),
    el("ul", { class: "tips" }, REF.tips.map((t) => el("li", {}, t))),
    el("h3", {}, "All the units"),
    el("table", { class: "unit-table" },
      el("thead", {}, el("tr", {}, el("th", {}, "Unit"), el("th", {}, "Written as"), el("th", {}, "Measures"), el("th", {}, "Measure with"))),
      el("tbody", {}, unitRows)));
}

// The classic "gallon man" picture: a gallon holds 4 quarts, each quart
// holds 2 pints, each pint holds 2 cups.
function gallonDiagram() {
  const cup = () => el("div", { class: "g-cup" }, "C");
  const pint = () => el("div", { class: "g-pint" }, el("span", {}, "P"), el("div", { class: "g-row" }, cup(), cup()));
  const quart = () => el("div", { class: "g-quart" }, el("span", {}, "Q"), el("div", { class: "g-row" }, pint(), pint()));
  return el("div", { class: "g-gallon", role: "img", "aria-label": "A gallon holds 4 quarts. Each quart holds 2 pints. Each pint holds 2 cups." },
    el("span", {}, "G"),
    el("div", { class: "g-quarts" }, quart(), quart(), quart(), quart()));
}

// ---------- Startup ----------

function wireUp() {
  for (const btn of document.querySelectorAll("[role=tab]")) {
    btn.addEventListener("click", () => showTab(btn.dataset.tab));
  }
  $("level-list").addEventListener("click", (e) => {
    const card = e.target.closest("[data-level]");
    if (card) startRound(Number(card.dataset.level));
  });
  $("answer-form").addEventListener("submit", (e) => { e.preventDefault(); checkAnswer(); });
  $("hint-btn").addEventListener("click", showHint);
  $("next-btn").addEventListener("click", nextProblem);
  $("quit").addEventListener("click", backToLevels);
  $("levels-btn").addEventListener("click", backToLevels);
  $("again-btn").addEventListener("click", () => startRound(round.level.id));
  $("next-level-btn").addEventListener("click", () => startRound(round.level.id + 1));

  // Worksheet controls
  const saved = store.get("ws", {});
  $("ws-levels").replaceChildren(...LEVELS.map((l) =>
    el("label", { class: "check" },
      el("input", { type: "checkbox", value: String(l.id), checked: (saved.levels || [1]).includes(l.id) }),
      ` ${l.id}. ${l.name}`)));
  $("ws-count").value = String(saved.count || 20);
  $("ws-title").value = saved.title ?? "";
  $("ws-title").placeholder = "Kitchen Unit Conversions";
  $("ws-version").value = String(saved.version || randomVersion());
  $("ws-cheat").checked = saved.cheat ?? true;
  $("ws-work").checked = saved.work ?? true;
  $("ws-form").addEventListener("input", () => { saveWs(); renderWorksheet(); });
  $("ws-new").addEventListener("click", () => { $("ws-version").value = String(randomVersion()); saveWs(); renderWorksheet(); });
  $("ws-print").addEventListener("click", () => printSheet("sheet"));
  $("ws-print-key").addEventListener("click", () => printSheet("key"));
  $("cheat-print").addEventListener("click", () => window.print());
}

(async function main() {
  try {
    await loadEngine();
    LEVELS = call("levels");
    REF = call("reference");
  } catch (err) {
    $("loading").hidden = true;
    $("load-error").hidden = false;
    $("load-error").textContent = `Sorry, this browser couldn't start the game (${err.message}). Try an up-to-date Chrome, Edge, Safari, or Firefox.`;
    return;
  }
  wireUp();
  renderLevelList();
  renderWorksheet();
  renderCheatSheet();
  $("loading").hidden = true;
  showTab(store.get("tab", "practice"));
})();
