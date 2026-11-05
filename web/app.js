// Recipe Math: The Line. All the math lives in the Go engine
// (window.recipeMath); this file draws the screens, runs the shift
// (plates, clock, tips, sounds), and lays out the printable pages.
"use strict";

const STORE_KEY = "recipeMath.v2";
const $ = (id) => document.getElementById(id);

// el("p", {class: "x"}, "text", child) builds DOM without innerHTML.
function el(tag, attrs, ...children) {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs || {})) {
    if (v === false || v == null) continue;
    if (k === "class") node.className = v;
    else if (k.startsWith("on")) node.addEventListener(k.slice(2), v);
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
    try {
      return JSON.parse(localStorage.getItem(STORE_KEY)) || {};
    } catch {
      return {};
    }
  })(),
  get(key, fallback) {
    return this.data[key] ?? fallback;
  },
  set(key, value) {
    this.data[key] = value;
    try {
      localStorage.setItem(STORE_KEY, JSON.stringify(this.data));
    } catch {
      /* private mode */
    }
  },
};

// Teacher settings in the URL: ?plates=4&timer=off&sound=off
const params = new URLSearchParams(location.search);
const MAX_PLATES = Math.min(10, Math.max(1, Number(params.get("plates")) || 4));
const TIMER_ON = params.get("timer") !== "off";
const AUTO_NEXT = params.get("auto") !== "off"; // clean answers move on by themselves
// ?seed=4821 makes every shift the same, like a worksheet version, so a whole
// class (or a test) gets identical tickets.
const SEED = Number(params.get("seed")) || null;

// Tips add up over every shift and unlock chef gear (cosmetic, just for fun).
const GEAR = [
  { tips: 0, icon: "🧢", name: "Paper cap" },
  { tips: 200, icon: "🥼", name: "Apron" },
  { tips: 500, icon: "🧣", name: "Side towel" },
  { tips: 1000, icon: "🔪", name: "Chef's knife" },
  { tips: 2000, icon: "👨‍🍳", name: "Toque" },
  { tips: 3500, icon: "🍳", name: "Copper pan" },
  { tips: 5000, icon: "🏆", name: "Golden whisk" },
];
const gearFor = (tips) => GEAR.filter((g) => tips >= g.tips).at(-1);
const nextGear = (tips) => GEAR.find((g) => g.tips > tips);

const randomSeed = () => crypto.getRandomValues(new Uint32Array(1))[0];
const randomVersion = () => 1000 + (randomSeed() % 9000);
const starText = (n) => "★".repeat(n) + "☆".repeat(3 - n);
const mmss = (s) => `${Math.floor(Math.abs(s) / 60)}:${String(Math.abs(s) % 60).padStart(2, "0")}`;
const letters = "ABCDEFGHIJ";
const cardKey = (c) => (c ? `${c.id}:${c.name}:${c.factor}` : "");

let RANKS = [];
let REF = null;

// ---------- Sound (synthesized, so the file stays offline and small) ----------

const sound = {
  modes: ["soft", "full", "off"],
  mode: params.get("sound") || store.get("sound", "soft"),
  ctx: null,
  master: null,
  ensure() {
    if (this.mode === "off") return null;
    if (!this.ctx) {
      const AC = window.AudioContext || window.webkitAudioContext;
      if (!AC) return null;
      this.ctx = new AC();
      this.master = this.ctx.createGain();
      this.master.connect(this.ctx.destination);
    }
    if (this.ctx.state === "suspended") this.ctx.resume();
    this.master.gain.value = this.mode === "full" ? 0.9 : 0.3;
    return this.ctx;
  },
  tone(freq, start, dur, { type = "sine", gain = 0.3, attack = 0.005 } = {}) {
    const ctx = this.ctx;
    const osc = ctx.createOscillator();
    const g = ctx.createGain();
    osc.type = type;
    osc.frequency.value = freq;
    g.gain.setValueAtTime(0, start);
    g.gain.linearRampToValueAtTime(gain, start + attack);
    g.gain.exponentialRampToValueAtTime(0.0001, start + dur);
    osc.connect(g).connect(this.master);
    osc.start(start);
    osc.stop(start + dur + 0.05);
  },
  noise(start, dur, { freq = 3000, q = 1, gain = 0.4 } = {}) {
    const ctx = this.ctx;
    const len = Math.ceil(ctx.sampleRate * dur);
    const buf = ctx.createBuffer(1, len, ctx.sampleRate);
    const data = buf.getChannelData(0);
    for (let i = 0; i < len; i++) data[i] = (Math.random() * 2 - 1) * Math.pow(1 - i / len, 3);
    const src = ctx.createBufferSource();
    src.buffer = buf;
    const filter = ctx.createBiquadFilter();
    filter.type = "bandpass";
    filter.frequency.value = freq;
    filter.Q.value = q;
    const g = ctx.createGain();
    g.gain.value = gain;
    src.connect(filter).connect(g).connect(this.master);
    src.start(start);
  },
  // "Order up!": two taps on a service bell.
  bell() {
    const ctx = this.ensure();
    if (!ctx) return;
    const t = ctx.currentTime + 0.02;
    for (const at of [t, t + 0.22]) {
      [
        [2093, 0.35],
        [4186, 0.08],
        [5777, 0.12],
        [11302, 0.04],
      ].forEach(([f, g]) => this.tone(f, at, 1.4, { gain: g, attack: 0.002 }));
    }
  },
  // A small, quiet "tick" for each correct answer.
  correct() {
    const ctx = this.ensure();
    if (!ctx) return;
    const t = ctx.currentTime + 0.01;
    this.tone(1318, t, 0.12, { gain: 0.12 });
    this.tone(1976, t + 0.06, 0.16, { gain: 0.1 });
  },
  // An accident: a plate hits the floor and clatters.
  crash() {
    const ctx = this.ensure();
    if (!ctx) return;
    const t = ctx.currentTime + 0.01;
    this.noise(t, 0.35, { freq: 2400, q: 0.8, gain: 0.55 });
    this.noise(t, 0.12, { freq: 600, q: 0.7, gain: 0.35 });
    [0.09, 0.17, 0.26, 0.31].forEach((d, i) => {
      this.noise(t + d, 0.08, { freq: 3500 + i * 700, q: 6, gain: 0.25 });
      this.tone(2900 + i * 610, t + d, 0.09, { gain: 0.05, type: "triangle" });
    });
  },
  // A recipe card is finished.
  chime() {
    const ctx = this.ensure();
    if (!ctx) return;
    const t = ctx.currentTime + 0.01;
    [1047, 1319, 1568].forEach((f, i) => this.tone(f, t + i * 0.09, 0.35, { type: "triangle", gain: 0.15 }));
  },
  // The shift is cleared.
  fanfare() {
    const ctx = this.ensure();
    if (!ctx) return;
    const t = ctx.currentTime + 0.01;
    [784, 988, 1175, 1568].forEach((f, i) => this.tone(f, t + i * 0.11, 0.5, { type: "triangle", gain: 0.16 }));
    setTimeout(() => this.bell(), 520);
  },
  // Knocked out: a sad slide down.
  womp() {
    const ctx = this.ensure();
    if (!ctx) return;
    const t = ctx.currentTime + 0.01;
    const osc = ctx.createOscillator();
    const g = ctx.createGain();
    osc.type = "triangle";
    osc.frequency.setValueAtTime(330, t);
    osc.frequency.exponentialRampToValueAtTime(120, t + 0.8);
    g.gain.setValueAtTime(0.25, t);
    g.gain.exponentialRampToValueAtTime(0.001, t + 0.9);
    osc.connect(g).connect(this.master);
    osc.start(t);
    osc.stop(t + 1);
  },
};

function renderSoundButton() {
  const icon = { soft: "🔉", full: "🔊", off: "🔇" }[sound.mode];
  const name = { soft: "Soft", full: "Full", off: "Off" }[sound.mode];
  $("sound-btn").textContent = `${icon} ${name}`;
  $("sound-btn").setAttribute("aria-label", `Sound: ${name}. Click to change.`);
}

// ---------- Loading the Go engine ----------

async function loadEngine() {
  const bin = atob(WASM_GZIP_BASE64);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
  const stream = new Blob([bytes]).stream().pipeThrough(new DecompressionStream("gzip"));
  const wasm = await new Response(stream).arrayBuffer();
  const go = new Go();
  const ready = new Promise((resolve) => {
    window.recipeMathReady = resolve;
  });
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
  // A shift pauses (clock, keys, auto-advance) while another tab is open.
  S.away = name !== "kitchen";
}

// ---------- Kitchen home ----------

function renderGear() {
  const life = store.get("lifetime", 0);
  const now = gearFor(life);
  const next = nextGear(life);
  const pct = next ? Math.round(((life - now.tips) / (next.tips - now.tips)) * 100) : 100;
  $("gear").replaceChildren(
    el("span", { class: "gear-icon", "aria-hidden": "true" }, now.icon),
    el(
      "span",
      { class: "gear-text" },
      el("strong", {}, `Your gear: ${now.name}`),
      el(
        "span",
        {},
        `💰 ${life} lifetime tips`,
        next ? ` · next: ${next.icon} ${next.name} at ${next.tips}` : " · you've earned it all!",
      ),
      el(
        "span",
        {
          class: "gear-bar",
          role: "progressbar",
          "aria-valuenow": String(pct),
          "aria-valuemin": "0",
          "aria-valuemax": "100",
          "aria-label": "Progress to next gear",
        },
        el("span", { style: `width:${pct}%` }),
      ),
    ),
  );
}

function renderRanks() {
  renderGear();
  const best = store.get("best", {});
  const pb = store.get("pb", {});
  const fresh = Object.keys(best).length === 0;
  $("rank-list").replaceChildren(
    ...RANKS.map((r) =>
      el(
        "li",
        {},
        el(
          "button",
          { type: "button", class: "rank-card", "data-rank": r.id },
          el("span", { class: "rank-num", "aria-hidden": "true" }, String(r.id)),
          el(
            "span",
            { class: "rank-text" },
            el(
              "strong",
              {},
              r.name,
              fresh && r.id === 1 ? el("span", { class: "badge start" }, "Start here") : "",
              !r.timed ? el("span", { class: "badge" }, "No clock") : "",
            ),
            el("span", { class: "rank-goal" }, r.goal),
            el("span", { class: "rank-example" }, r.example),
          ),
          el(
            "span",
            { class: "rank-side" },
            el(
              "span",
              { class: "rank-stars", "aria-label": `Best: ${best[r.id] || 0} of 3 stars` },
              starText(best[r.id] || 0),
            ),
            pb[r.id] ? el("span", { class: "rank-pb" }, `Best: 💰 ${pb[r.id].tips}`) : "",
            el("span", { class: "rank-go" }, "Start shift →"),
          ),
        ),
      ),
    ),
  );
}

// ---------- The shift ----------

const S = {}; // the shift in progress

function startShift(rankId) {
  sound.ensure(); // first user gesture: allowed to start audio
  const orders = call("shift", rankId, SEED ?? randomSeed()).orders;
  clearInterval(S.timer);
  Object.assign(S, {
    id: (S.id ?? 0) + 1, // stale timers compare this before acting
    rank: RANKS.find((r) => r.id === rankId),
    orders,
    queue: orders.map((_, i) => i), // queue[0] is the current ticket
    doneOrder: [],
    rec: orders.map(() => ({ accidents: 0, hints: 0, shown: 0, skips: 0, solved: false, wrongChoices: [] })),
    plates: MAX_PLATES,
    tips: 0,
    accidents: 0,
    hintsUsed: 0,
    timeLeft: orders.reduce((s, o) => s + o.seconds, 0),
    late: false,
    paused: false,
    phase: "answer",
    bowl: [],
    over: false,
    streak: 0,
    bestStreak: 0,
    lastPlates: MAX_PLATES,
  });
  S.timer = setInterval(tick, 1000);
  $("home").hidden = true;
  $("result").hidden = true;
  $("shift").hidden = false;
  S.timed = TIMER_ON && S.rank.timed;
  $("clock").hidden = !S.timed;
  renderHUD();
  showOrder();
  sound.bell();
  flash("🛎 Order up! First ticket's on the rail.", "info");
}

const cur = () => S.orders[S.queue[0]];
const curRec = () => S.rec[S.queue[0]];

function tick() {
  if (S.over || S.paused || S.away || !S.timed) return;
  S.timeLeft--;
  if (S.timeLeft === 0) {
    S.late = true;
    flash("⏰ Tickets are late! No more time bonus, but keep cooking.", "warn");
  }
  renderClock();
}

function renderClock() {
  const c = $("clock");
  c.textContent = S.timeLeft >= 0 ? `⏱ ${mmss(S.timeLeft)}` : `⏱ +${mmss(S.timeLeft)}`;
  c.classList.toggle("late", S.timeLeft < 0);
  c.classList.toggle("hurry", S.timeLeft >= 0 && S.timeLeft <= 20);
}

function renderHUD() {
  const p = $("plates");
  const icons = [];
  for (let i = 0; i < MAX_PLATES; i++) {
    const fill = Math.max(0, Math.min(1, S.plates - i));
    icons.push(el("span", { class: `plate ${fill === 1 ? "full" : fill > 0 ? "half" : "empty"}` }));
  }
  // Juice: the plate that was just lost falls; one that was won back pops.
  const changed = Math.floor(Math.min(S.plates, S.lastPlates));
  if (icons[changed] && S.plates < S.lastPlates) icons[changed].classList.add("drop");
  if (icons[changed] && S.plates > S.lastPlates) icons[changed].classList.add("heal");
  S.lastPlates = S.plates;
  p.replaceChildren(...icons);
  const st = $("streak");
  st.hidden = S.streak < 3;
  if (S.streak >= 3) {
    st.textContent = `🔥 ×${S.streak}`;
    st.setAttribute("aria-label", `Hot line: ${S.streak} clean answers in a row`);
    st.classList.remove("pop");
    void st.offsetWidth;
    st.classList.add("pop");
  }
  p.setAttribute("aria-label", `Plates: ${S.plates} of ${MAX_PLATES}`);
  $("tips").textContent = `💰 ${S.tips}`;
  renderClock();
  renderRail();
}

function renderRail() {
  const order = [...S.doneOrder, ...S.queue];
  $("rail").replaceChildren(
    ...order.map((i, pos) => {
      const r = S.rec[i];
      const o = S.orders[i];
      const isCur = pos === S.doneOrder.length && !S.over;
      const state = r.solved
        ? r.accidents || r.hints
          ? "rough"
          : "clean"
        : isCur
          ? "now"
          : r.skips
            ? "skipped"
            : "todo";
      return el(
        "li",
        { class: `chip ${state}${isCur ? " now" : ""}`, title: o.station },
        o.card && o.kind !== "fix" ? "📋" : "",
        String(o.id),
      );
    }),
  );
  const now = $("rail").querySelector(".now");
  if (now && now.scrollIntoView) now.scrollIntoView({ block: "nearest", inline: "center" });
}

function flash(text, kind) {
  const fb = $("feedback");
  fb.className = `feedback ${kind}`;
  fb.textContent = text;
}

function showOrder() {
  const o = cur();
  const r = curRec();
  S.phase = "answer";
  S.paused = false;
  S.bowl = [];
  S.selected = null;
  $("t-num").textContent = `#${o.id}`;
  $("t-station").textContent = o.station;
  $("t-prompt").textContent = o.prompt;
  $("t-context").textContent = o.context || "";
  $("t-context").hidden = !o.context;
  flash("", "");
  $("explain").hidden = true;
  $("ticket").classList.remove("shake", "solved");

  $("in-choice").hidden = o.input !== "choice";
  $("in-number").hidden = o.input !== "number";
  $("in-build").hidden = o.input !== "build";
  $("ticket").classList.toggle("is-build", o.input === "build");
  if (o.input === "choice") renderChoices(o, r);
  if (o.input === "number") {
    $("num").value = "";
    $("num").disabled = false;
    $("num-unit").textContent = o.unit === "×" ? "" : o.unit;
    $("num").placeholder = o.unit === "×" ? "×?" : "";
  }
  if (o.input === "build") {
    renderDrawer(o);
    renderBowl();
  }
  renderHints(o, r);
  renderCard(o);

  $("serve").hidden = false;
  $("serve").textContent = "🛎 Serve";
  $("serve").disabled = o.input === "choice";
  $("hint-btn").hidden = false;
  $("skip-btn").hidden = false;
  $("next-btn").hidden = true;
  updateButtons();
  renderRail();
  if (o.input === "number" && matchMedia("(pointer: fine)").matches) $("num").focus();
}

function updateButtons() {
  const o = cur();
  const r = curRec();
  const moreHints = r.shown < o.hints.length;
  $("hint-btn").disabled = !moreHints || S.plates <= 0.5;
  $("hint-btn").title = !moreHints
    ? "No more hints for this ticket"
    : S.plates <= 0.5
      ? "Not enough plates for a hint"
      : "";
  const skipCosts = r.skips >= 1;
  $("skip-btn").disabled = S.queue.length < 2 || (skipCosts && S.plates <= 0.5);
  $("skip-btn").textContent = skipCosts ? "↪ Skip (−½ 🍽)" : "↪ Skip";
}

// Tap once to pick, tap again (or press Serve) to serve, so one slip of a
// finger doesn't cost a plate.
function pickChoice(i) {
  const o = cur();
  if (S.phase !== "answer" || curRec().wrongChoices.includes(i)) return;
  if (S.selected === i) return serve({});
  S.selected = i;
  $("serve").disabled = false;
  $("serve").textContent = `🛎 Serve ${letters[i]}`;
  for (const b of $("in-choice").querySelectorAll("[data-choice]")) {
    b.setAttribute("aria-pressed", String(Number(b.dataset.choice) === i));
  }
  if (o.kind !== "fix") $("serve").focus();
}

function renderChoices(o, r) {
  const box = $("in-choice");
  box.classList.toggle("ledger", o.kind === "fix");
  const state = (i) => ({
    type: "button",
    class: `choice${r.wrongChoices.includes(i) ? " struck" : ""}`,
    "data-choice": i,
    "aria-pressed": String(S.selected === i),
    disabled: r.wrongChoices.includes(i) || S.phase !== "answer",
    onclick: () => pickChoice(i),
  });
  if (o.kind === "fix") {
    // Fix the Mistake: one table, original next to what the cook wrote.
    const c = o.card;
    box.replaceChildren(
      el(
        "div",
        { class: "ledger-head", "aria-hidden": "true" },
        el("span", {}, ""),
        el("span", {}, "Ingredient"),
        el("span", {}, `Recipe (makes ${c.makes})`),
        el("span", {}, `Cook wrote (×${c.factor} → ${c.want})`),
      ),
      ...c.lines.map((l, i) =>
        el(
          "button",
          { ...state(i), "aria-label": `${letters[i]}: ${o.choices[i]}` },
          el("span", { class: "choice-key", "aria-hidden": "true" }, letters[i]),
          el("span", { class: "lg-ing" }, l.tool),
          el("span", { class: "lg-orig" }, l.text),
          el("span", { class: "lg-cook" }, el("span", { class: "lg-arrow", "aria-hidden": "true" }, "→ "), l.scaled),
        ),
      ),
    );
    return;
  }
  box.replaceChildren(
    ...o.choices.map((text, i) =>
      el(
        "button",
        state(i),
        el("span", { class: "choice-key", "aria-hidden": "true" }, letters[i]),
        el("span", {}, text),
      ),
    ),
  );
}

function renderHints(o, r) {
  $("hints").replaceChildren(
    ...o.hints
      .slice(0, r.shown)
      .map((h, i) =>
        el(
          "li",
          {},
          el("span", { class: "hint-num" }, `💡 ${i + 1}`),
          " ",
          h,
          o.visual && i === o.hints.length - 1 ? picture(o.visual, o.visual.given) : "",
        ),
      ),
  );
}

// ---------- The recipe card (mise en place board) ----------

function renderCard(o) {
  const box = $("card");
  if (!o.card || o.kind === "fix") {
    box.hidden = true;
    return;
  }
  box.hidden = false;
  const key = cardKey(o.card);
  const isFix = o.kind === "fix";
  // Which lines (and the factor) are already solved on this card?
  const solvedLines = new Set();
  let factorSolved = isFix;
  S.orders.forEach((x, i) => {
    if (cardKey(x.card) !== key || !S.rec[i].solved) return;
    if (x.kind === "factor") factorSolved = true;
    else solvedLines.add(x.line);
  });
  const c = o.card;
  box.replaceChildren(
    el(
      "div",
      { class: "card-head" },
      el("span", { class: "card-tag" }, isFix ? "Original recipe" : "Recipe card"),
      el("h3", {}, c.name),
      el(
        "p",
        { class: "card-yield" },
        `Makes ${c.makes} ${c.makesWhat} → need ${c.want} `,
        el("span", { class: `factor ${factorSolved ? "set" : ""}` }, factorSolved ? `×${c.factor}` : "×?"),
      ),
    ),
    el(
      "ol",
      { class: "card-lines" },
      c.lines.map((l, i) =>
        el(
          "li",
          { class: `${o.line === i ? "active" : ""} ${solvedLines.has(i) ? "done" : ""}` },
          el("span", { class: "orig" }, l.text),
          !isFix
            ? el("span", { class: "scaled" }, solvedLines.has(i) ? `✓ ${l.scaled}` : o.line === i ? "◀ now" : "")
            : "",
        ),
      ),
    ),
  );
}

// ---------- Measure It: drawer and bowl ----------

function toolById(o, id) {
  return o.tools.find((t) => t.id === id);
}

function addTool(o, t) {
  if (S.phase !== "answer") return;
  if (t.kind === "liquid") {
    // One fill line at a time; tapping the same line pours it out.
    const had = S.bowl.find((id) => toolById(o, id).kind === "liquid");
    S.bowl = S.bowl.filter((id) => toolById(o, id).kind !== "liquid");
    if (had !== t.id) S.bowl.unshift(t.id);
  } else {
    if (S.bowl.length >= 24) return;
    S.bowl.push(t.id);
  }
  renderBowl();
  renderGlass(o);
}

function renderDrawer(o) {
  const toolBtn = (t) =>
    el(
      "button",
      { type: "button", class: `tool ${t.kind}`, "aria-label": `Add ${t.name}`, onclick: () => addTool(o, t) },
      el("span", { class: `tool-icon ${t.kind}`, style: `--s:${toolScale(t)}` }),
      el("span", { class: "tool-label" }, t.label),
    );
  const group = (title, tools) =>
    tools.length
      ? el("div", { class: "tool-group" }, el("h4", {}, title), el("div", { class: "tool-row" }, tools.map(toolBtn)))
      : "";
  const spoons = o.tools.filter((t) => t.kind === "spoon");
  const d = $("drawer");
  if (o.drawer === "liquid") {
    d.replaceChildren(el("div", { class: "liquid-wrap" }, glass(o), group("Measuring spoons", spoons)));
  } else if (o.drawer === "butter") {
    d.replaceChildren(group("Butter", o.tools));
  } else {
    d.replaceChildren(
      group(
        "Dry measuring cups",
        o.tools.filter((t) => t.kind === "dry"),
      ),
      group("Measuring spoons", spoons),
    );
  }
  renderGlass(o);
}

// Icon size grows with the tool, so students see ⅛ t is tiny and 1 c is big.
function toolScale(t) {
  return (0.45 + 0.55 * Math.sqrt(Math.min(t.size, 384) / 384)).toFixed(2);
}

// The glass liquid measuring cup: cups in ¼s on the left, ⅓s on the right.
function glass(o) {
  const big = o.answer.vol > 768;
  const top = big ? 1536 : 768;
  const lines = o.tools.filter((t) => t.kind === "liquid" && t.size <= top);
  const g = el(
    "div",
    { class: "glass", id: "glass", "data-top": String(top) },
    el("div", { class: "glass-fill", id: "glass-fill" }),
  );
  for (const t of lines) {
    const third = t.size % 96 !== 0;
    g.append(
      el(
        "button",
        {
          type: "button",
          class: `gline ${third ? "third" : "quarter"}`,
          "data-tool": t.id,
          style: `bottom:${(t.size / top) * 100}%`,
          "aria-label": `Fill to the ${t.label} line`,
          onclick: () => addTool(o, t),
        },
        el("span", {}, t.label.replace(" c", "")),
      ),
    );
  }
  return el(
    "div",
    { class: "glass-box" },
    el("h4", {}, `${big ? "4" : "2"}-cup liquid measuring cup`),
    el("p", { class: "fine" }, "Tap a line to fill to it. Tap it again to pour out."),
    g,
  );
}

function renderGlass(o) {
  const fill = $("glass-fill");
  if (!fill || !$("glass")) return;
  const id = S.bowl.find((x) => toolById(o, x).kind === "liquid");
  const top = Number($("glass").dataset.top);
  fill.style.height = id ? `${(toolById(o, id).size / top) * 100}%` : "0";
  for (const b of document.querySelectorAll(".gline")) {
    b.classList.toggle("on", b.dataset.tool === id);
    b.setAttribute("aria-pressed", String(b.dataset.tool === id));
  }
}

function renderBowl() {
  const o = cur();
  const bowl = $("bowl");
  if (S.bowl.length === 0) {
    bowl.replaceChildren(el("span", { class: "bowl-empty" }, "Empty bowl. Tap tools below to add scoops."));
    return;
  }
  // The bowl groups scoops ("1 T ×3") but never adds them up: that's the
  // student's job. Tapping a group takes one scoop out.
  const groups = [];
  for (const id of S.bowl) {
    const g = groups.find((x) => x.id === id);
    if (g) g.n++;
    else groups.push({ id, n: 1, t: toolById(o, id) });
  }
  groups.sort((a, b) => b.t.size - a.t.size);
  bowl.replaceChildren(
    ...groups.map(({ id, n, t }) =>
      el(
        "button",
        {
          type: "button",
          class: `scoop ${t.kind}`,
          "aria-label": `${t.name}, ${n} scoop${n > 1 ? "s" : ""}. Remove one.`,
          disabled: S.phase !== "answer",
          onclick: () => {
            S.bowl.splice(S.bowl.lastIndexOf(id), 1);
            renderBowl();
            renderGlass(o);
          },
        },
        t.kind === "liquid" ? `${t.label} line` : t.label,
        n > 1 ? el("b", { class: "scoop-n" }, ` ×${n}`) : "",
        el("span", { class: "scoop-minus", "aria-hidden": "true" }, "−"),
      ),
    ),
    el("span", { class: "scoop-count" }, `${S.bowl.length} scoop${S.bowl.length === 1 ? "" : "s"}`),
  );
}

// ---------- Serving, hints, skipping ----------

// "½ plate", "1 plate", "2½ plates".
function platesText(n) {
  const whole = Math.floor(n);
  const num = n % 1 ? (whole ? `${whole}½` : "½") : String(whole);
  return `${num} plate${n === 1 || n === 0.5 ? "" : "s"}`;
}

// What the student is serving, or null (with a nudge) if nothing is picked.
function readResponse(o, extra) {
  const resp = { choice: -1, tools: [], text: "" };
  if (o.input === "choice") {
    resp.choice = extra.choice ?? S.selected;
    if (resp.choice == null) {
      flash("Tap a choice first.", "warn");
      return null;
    }
  }
  if (o.input === "number") resp.text = $("num").value;
  if (o.input === "build") resp.tools = S.bowl;
  return resp;
}

function serve(extra) {
  if (S.over || S.phase !== "answer") return;
  const o = cur();
  const resp = readResponse(o, extra);
  if (!resp) return;
  const res = call("check", JSON.stringify(o), JSON.stringify(resp));
  if (res.invalid) flash(res.message, "warn");
  else if (res.correct) onCorrect(o, curRec(), res);
  else onAccident(o, curRec(), res, resp);
}

function shakeTicket() {
  const t = $("ticket");
  t.classList.remove("shake");
  void t.offsetWidth; // restart the animation
  t.classList.add("shake");
}

function onAccident(o, r, res, resp) {
  r.accidents++;
  S.accidents++;
  S.streak = 0; // quietly: the lost bonus is penalty enough
  S.plates = Math.max(0, S.plates - 1);
  if (o.input === "choice") {
    r.wrongChoices.push(resp.choice);
    S.selected = null;
    $("serve").disabled = true;
    $("serve").textContent = "🛎 Serve";
  }
  if (res.freeHint && r.shown === 0) r.shown = 1; // the message already asked hint 1
  sound.crash();
  shakeTicket();
  flash(`💥 Accident! (${platesText(S.plates)} left) ${res.message}`, "bad");
  $("feedback").focus({ preventScroll: true });
  renderHUD();
  renderHints(o, r);
  if (o.input === "choice") renderChoices(o, r);
  updateButtons();
  if (S.plates <= 0) {
    // Out of plates: freeze the ticket now, then show the result. The id
    // check stops a stale timer from ending a shift started in the meantime.
    S.phase = "out";
    for (const id of ["serve", "hint-btn", "skip-btn"]) $(id).disabled = true;
    const shift = S.id;
    setTimeout(() => {
      if (S.id === shift) endShift(false);
    }, 900);
  }
}

// Tips for a correct serve: more for a clean first try, a bonus for the
// fewest scoops, and a streak bonus from 3 clean answers in a row.
function tipsFor(r, res) {
  const clean = r.accidents === 0 && r.hints === 0;
  let earned = clean ? 10 : 5;
  let note = "";
  if (res.bestScoops && res.scoops === res.bestScoops) {
    earned += 5;
    note = " Fewest scoops: +5!";
  }
  S.streak = clean ? S.streak + 1 : 0;
  S.bestStreak = Math.max(S.bestStreak, S.streak);
  if (S.streak >= 3) {
    earned += 2;
    note += ` 🔥 Hot line ×${S.streak}: +2!`;
  }
  return { clean, earned, note };
}

// Did this serve finish its recipe card?
function finishesCard(o) {
  if (!o.card || o.kind === "fix") return false;
  const key = cardKey(o.card);
  return S.orders.every((x, i) => cardKey(x.card) !== key || S.rec[i].solved);
}

function onCorrect(o, r, res) {
  r.solved = true;
  S.phase = "review";
  S.paused = true;
  S.plates = Math.min(MAX_PLATES, S.plates + 0.5);
  const { clean, earned, note } = tipsFor(r, res);
  S.tips += earned;
  sound.correct();
  const cardDone = finishesCard(o);
  if (cardDone) sound.chime();
  flash(
    `✅ ${res.message} +${earned} tips.${note}${cardDone ? " 📋 Recipe card complete. Mise en place is ready!" : ""}`,
    "good",
  );
  showSolved(o);
  scheduleAutoNext(o, clean, cardDone);
}

// Lock the ticket, explain the answer, and offer Next.
function showSolved(o) {
  $("ticket").classList.add("solved");
  $("explain").replaceChildren(
    el("h4", {}, "Chef says"),
    ...o.explain.map((e) => el("p", {}, e)),
    o.visual && o.visual.answer ? el("div", { class: "explain-pic" }, picture(o.visual, o.visual.answer)) : "",
  );
  $("explain").hidden = false;
  if (o.input === "number") $("num").disabled = true;
  if (o.input === "choice") {
    for (const b of $("in-choice").querySelectorAll("[data-choice]")) {
      b.disabled = true;
      b.classList.add("locked"); // locked, not crossed out: still readable for review
      if (o.answer.choices.includes(Number(b.dataset.choice))) b.classList.add("right");
    }
  }
  if (o.input === "build") renderBowl();
  renderCard(o);
  renderHUD();
  for (const id of ["serve", "hint-btn", "skip-btn"]) $(id).hidden = true;
  $("next-btn").hidden = false;
  $("next-btn").focus();
}

// A clean first-try answer moves on by itself; after an accident or a hint
// the student taps Next, because that's when the explanation matters.
function scheduleAutoNext(o, clean, cardDone) {
  const nb = $("next-btn");
  nb.classList.remove("auto");
  if (!clean || !AUTO_NEXT) return;
  const ms = cardDone ? 2600 : 1800;
  nb.style.setProperty("--auto", `${ms}ms`);
  void nb.offsetWidth; // restart the fill animation
  nb.classList.add("auto");
  clearTimeout(S.autoTimer);
  S.autoTimer = setTimeout(() => {
    if (!S.over && !S.away && S.phase === "review" && cur() === o) nextOrder();
  }, ms);
}

function nextOrder() {
  clearTimeout(S.autoTimer);
  $("next-btn").classList.remove("auto");
  if (S.phase !== "review") return;
  S.doneOrder.push(S.queue.shift());
  if (S.queue.length === 0) return endShift(true);
  showOrder();
}

function useHint() {
  const o = cur();
  const r = curRec();
  if (r.shown >= o.hints.length || S.plates <= 0.5 || S.phase !== "answer") return;
  r.shown++;
  r.hints++;
  S.hintsUsed++;
  S.streak = 0;
  S.plates -= 0.5;
  renderHints(o, r);
  renderHUD();
  updateButtons();
}

function skipOrder() {
  const r = curRec();
  if (S.queue.length < 2 || S.phase !== "answer") return;
  if (r.skips >= 1) {
    if (S.plates <= 0.5) return;
    S.plates -= 0.5;
  }
  r.skips++;
  S.queue.push(S.queue.shift());
  renderHUD();
  showOrder();
  flash(
    r.skips === 1
      ? "↪ Sent to the back of the rail. It'll come back around."
      : "↪ Skipped again: that cost half a plate.",
    "info",
  );
}

// ---------- End of shift ----------

function endShift(cleared) {
  if (S.over) return;
  S.over = true;
  clearInterval(S.timer);
  clearTimeout(S.autoTimer);
  const score = settleScore(cleared);
  if (cleared) sound.fanfare();
  else sound.womp();
  $("shift").hidden = true;
  $("result").hidden = false;
  renderResultHeader(cleared, score);
  renderUnlocks(score);
  renderStats();
  renderReview(cleared);
  const next = RANKS.find((r) => r.id === S.rank.id + 1);
  $("up-btn").hidden = !cleared || !next;
  $("again-btn").focus();
}

// Add the time bonus, then save lifetime tips, personal bests, and stars.
function settleScore(cleared) {
  const onTime = S.timed && !S.late;
  const bonus = cleared && onTime ? Math.floor(Math.max(0, S.timeLeft) / 3) : 0; // 1 tip per 3 seconds left
  S.tips += bonus;
  const lifeBefore = store.get("lifetime", 0);
  const lifeAfter = lifeBefore + S.tips;
  store.set("lifetime", lifeAfter);
  const pb = store.get("pb", {});
  const prevBest = pb[S.rank.id];
  const newBest = cleared && (!prevBest || S.tips > prevBest.tips);
  if (newBest) {
    pb[S.rank.id] = { tips: S.tips, plates: S.plates };
    store.set("pb", pb);
  }
  const half = S.plates >= MAX_PLATES / 2;
  const stars = cleared ? 1 + (half ? 1 : 0) + (half && (onTime || !S.timed) ? 1 : 0) : 0;
  const best = store.get("best", {});
  if (stars > (best[S.rank.id] || 0)) {
    best[S.rank.id] = stars;
    store.set("best", best);
  }
  const unlocked = GEAR.filter((g) => g.tips > lifeBefore && g.tips <= lifeAfter);
  return { onTime, bonus, stars, lifeAfter, unlocked, prevBest, newBest };
}

function resultTitle(cleared, stars) {
  if (!cleared) return "Chef pulled you off the line!";
  if (stars === 3 && S.accidents === 0 && S.hintsUsed === 0) {
    return "Flawless service! Chef wants you on the line every night.";
  }
  return ["", "Shift cleared. Rough night, but you made it.", "Solid shift!", "Chef is impressed!"][stars];
}

function resultSubtitle(cleared, { onTime, bonus }) {
  if (!cleared) return "You dropped all your plates. Study what tripped you up, then get back in there.";
  if (onTime) return `On-time bonus: +${bonus} tips.`;
  return S.timed ? "Tickets ran late, so no time bonus." : "";
}

function renderResultHeader(cleared, score) {
  $("r-stars").textContent = cleared ? starText(score.stars) : "🍽💥";
  $("r-title").textContent = resultTitle(cleared, score.stars);
  $("r-sub").textContent = resultSubtitle(cleared, score);
}

function renderUnlocks({ newBest, prevBest, unlocked, lifeAfter }) {
  const best = prevBest
    ? `🏅 New personal best: ${S.tips} tips (was ${prevBest.tips})`
    : `🏅 First clear! Personal best: ${S.tips} tips`;
  const nxt = nextGear(lifeAfter);
  $("r-unlock").replaceChildren(
    newBest ? el("p", { class: "pb" }, best) : "",
    ...unlocked.map((g) => el("p", { class: "unlock" }, `🎉 Unlocked: ${g.icon} ${g.name}`)),
    el(
      "p",
      { class: "next-gear" },
      nxt ? `${nxt.tips - lifeAfter} more tips to ${nxt.icon} ${nxt.name}` : "All chef gear unlocked. 🏆",
    ),
  );
}

function renderStats() {
  const stats = [
    ["Tips", `💰 ${S.tips}`],
    ["Best streak", `🔥 ${S.bestStreak}`],
    ["Plates left", `🍽 ${platesText(S.plates)} of ${MAX_PLATES}`],
    ["Accidents", String(S.accidents)],
    ["Hints", String(S.hintsUsed)],
    ["Tickets", `${S.rec.filter((r) => r.solved).length} of ${S.orders.length}`],
  ];
  $("r-stats").replaceChildren(...stats.map(([k, v]) => el("div", {}, el("dt", {}, k), el("dd", {}, v))));
}

// Review anything that needed help, plus (on a knockout) the ticket that ended it.
function renderReview(cleared) {
  const ender = S.queue[0];
  const review = S.orders
    .map((o, i) => ({ o, r: S.rec[i], i }))
    .filter(({ r, i }) => r.accidents || r.hints || r.skips > 1 || (!cleared && i === ender));
  if (review.length === 0) {
    $("r-review").replaceChildren(el("p", { class: "clean" }, "No accidents, no hints. Clean service! 👏"));
    return;
  }
  const item = ({ o }) =>
    el(
      "li",
      {},
      el("p", { class: "rv-q" }, el("span", { class: "rv-station" }, o.station), " ", o.prompt),
      el("p", { class: "rv-a" }, "Answer: ", el("strong", {}, o.answer.text)),
      ...o.explain.map((e) => el("p", { class: "rv-e" }, e)),
    );
  $("r-review").replaceChildren(
    el("div", { class: "review" }, el("h3", {}, "Study these"), el("ol", {}, review.map(item))),
  );
}

function goHome() {
  clearInterval(S.timer);
  clearTimeout(S.autoTimer);
  S.over = true;
  $("shift").hidden = true;
  $("result").hidden = true;
  $("home").hidden = false;
  renderRanks();
}

// ---------- Worksheet (paper parity) ----------

function wsSettings() {
  const checked = document.querySelector("#ws-ranks input:checked");
  return {
    rank: checked ? Number(checked.value) : 1,
    title: $("ws-title").value.trim() || "Recipe Math: The Line",
    version: Math.min(9999, Math.max(1000, Number($("ws-version").value) || 1000)),
  };
}

function renderWorksheet(mode) {
  mode = mode || document.querySelector("input[name=ws-preview]:checked").value;
  const key = mode === "key";
  const s = wsSettings();
  const rank = RANKS.find((r) => r.id === s.rank);
  const orders = call("shift", s.rank, s.version).orders;
  const sheet = $("ws-sheet");
  sheet.classList.toggle("is-key", key);

  const head = el(
    "header",
    { class: "sheet-head" },
    el("h2", {}, key ? `${s.title}: Answer Key` : s.title, el("small", { class: "version" }, `Version ${s.version}`)),
    el("p", { class: "meta" }, `Station ${rank.id}: ${rank.name}. ${rank.goal}`),
    key
      ? ""
      : el(
          "div",
          { class: "name-line" },
          el("span", {}, "Name ", el("i", { class: "blank wide" })),
          el("span", {}, "Date ", el("i", { class: "blank" })),
          el("span", {}, "Period ", el("i", { class: "blank short" })),
        ),
  );
  const instructions = el(
    "p",
    { class: "instructions" },
    "Use recipe notation (T, t, c). For “measure” questions, write the tools you'd use, like ¼ c + 1 T + 1 t. Circle one letter for multiple choice.",
  );

  const items = [];
  let n = 0;
  for (let i = 0; i < orders.length; i++) {
    const o = orders[i];
    if (o.card && o.kind !== "fix") {
      // Collect the whole recipe card into one table.
      const k = cardKey(o.card);
      const group = [];
      while (i < orders.length && cardKey(orders[i].card) === k) group.push(orders[i++]);
      i--;
      items.push(cardTable(++n, group, key));
      continue;
    }
    items.push(printOrder(++n, o, key));
  }
  sheet.replaceChildren(head, instructions, el("ol", { class: "ws-items" }, items));
}

const blank = (cls = "") => el("i", { class: `blank ${cls}` });
const keyAns = (text) => el("b", { class: "ans" }, text);

function printOrder(num, o, key) {
  const parts = [el("p", { class: "q" }, el("span", { class: "q-station" }, o.station), " ", o.prompt)];
  if (o.context && o.kind === "scale") parts.push(el("p", { class: "q-ctx" }, o.context));
  if (o.kind === "fix") {
    const c = o.card;
    parts.push(
      el(
        "table",
        { class: "card-table fix-table" },
        el(
          "thead",
          {},
          el(
            "tr",
            {},
            el("th", {}, ""),
            el("th", {}, "Ingredient"),
            el("th", {}, `Recipe (makes ${c.makes})`),
            el("th", {}, `Cook wrote (×${c.factor})`),
            el("th", {}, "Wrong?"),
          ),
        ),
        el(
          "tbody",
          {},
          c.lines.map((l, i) =>
            el(
              "tr",
              { class: key && o.answer.choices.includes(i) ? "right" : "" },
              el("td", { class: "ln" }, letters[i]),
              el("td", {}, l.tool),
              el("td", {}, l.text),
              el("td", {}, l.scaled),
              el("td", { class: "ln" }, key && o.answer.choices.includes(i) ? "✗" : "☐"),
            ),
          ),
        ),
      ),
    );
    if (key) parts.push(el("p", { class: "q-why" }, o.explain.join(" ")));
    return el("li", { class: "ws-item ws-card", value: String(num) }, parts);
  }
  if (o.input === "choice") {
    parts.push(
      el(
        "ol",
        { class: "q-choices", type: "A" },
        o.choices.map((c, i) => el("li", { class: key && o.answer.choices.includes(i) ? "right" : "" }, c)),
      ),
    );
  } else if (o.input === "number") {
    parts.push(
      el(
        "p",
        { class: "q-line" },
        o.unit === "×" ? "Factor: ×" : "Answer: ",
        key ? keyAns(o.answer.text.replace(/^×/, "")) : blank(),
        key || o.unit === "×" ? "" : ` ${o.unit}`,
      ),
    );
  } else {
    parts.push(el("p", { class: "q-line" }, "Tools: ", key ? keyAns(o.answer.text) : blank("xwide")));
  }
  if (key) parts.push(el("p", { class: "q-why" }, o.explain.join(" ")));
  return el("li", { class: "ws-item", value: String(num) }, parts);
}

function cardTable(num, group, key) {
  const c = group[0].card;
  const byLine = new Map(group.filter((o) => o.kind !== "factor").map((o) => [o.line, o]));
  return el(
    "li",
    { class: "ws-item ws-card", value: String(num) },
    el(
      "p",
      { class: "q" },
      el("span", { class: "q-station" }, "Recipe Card"),
      ` ${c.name} makes ${c.makes} ${c.makesWhat}. The order is for ${c.want}.`,
    ),
    el("p", { class: "q-line" }, "Conversion factor (want ÷ makes): ×", key ? keyAns(c.factor) : blank("short")),
    el(
      "table",
      { class: "card-table" },
      el(
        "thead",
        {},
        el(
          "tr",
          {},
          el("th", {}, ""),
          el("th", {}, "Recipe says"),
          el("th", {}, "Grab"),
          el("th", {}, "Scaled: measure it with"),
        ),
      ),
      el(
        "tbody",
        {},
        c.lines.map((l, i) => {
          const o = byLine.get(i);
          const isCall = o && o.kind === "call";
          return el(
            "tr",
            {},
            el("td", { class: "ln" }, letters[i].toLowerCase()),
            el("td", {}, l.text),
            el("td", { class: "tool-col" }, key ? l.tool : ""),
            el(
              "td",
              {},
              key ? keyAns(isCall ? `${l.scaled}: ${o.answer.text}` : l.scaled) : isCall ? "Chef's call: " : "",
            ),
          );
        }),
      ),
    ),
  );
}

function printSheet(mode) {
  renderWorksheet(mode);
  const restore = () => {
    renderWorksheet();
    window.removeEventListener("afterprint", restore);
  };
  window.addEventListener("afterprint", restore);
  window.print();
}

// ---------- Pictures: the cup wall and butter sticks ----------

// One cup split into equal blocks, row by row, so "4 T = ¼ c" is something
// you can see and count. Shading marks an amount without labeling it.
const WALL_ROWS = [
  ["1 cup", 1],
  ["½ cup", 2],
  ["⅓ cup", 3],
  ["¼ cup", 4],
  ["2 T", 8],
  ["1 T", 16],
];

function cupWall(shade) {
  const wall = el(
    "div",
    {
      class: "cup-wall",
      role: "img",
      "aria-label": shade
        ? "Cup wall: one cup split into halves, thirds, quarters, 2-tablespoon and 1-tablespoon blocks. The shaded part is the amount on this ticket. Count the blocks under it."
        : "Cup wall: 1 cup = 2 half cups = 3 third cups = 4 quarter cups = 8 two-tablespoon blocks = 16 tablespoons.",
    },
    ...WALL_ROWS.map(([label, n]) =>
      el(
        "div",
        { class: "wall-row" },
        Array.from({ length: n }, () => el("span", { class: "wall-block" }, label)),
      ),
    ),
  );
  if (shade) wall.append(el("div", { class: "wall-shade", style: `width:${Math.min(100, (shade / 384) * 100)}%` }));
  return wall;
}

// Butter sticks with the tablespoon marks from the wrapper; shaded marks
// show an amount.
function butterSticks(shade, count = 1) {
  const marks = Math.round(shade / 24);
  return el(
    "div",
    {
      class: "butter-pic",
      role: "img",
      "aria-label": shade
        ? `Butter sticks with ${marks} tablespoon marks shaded. Count the sticks and marks.`
        : "A stick of butter: 1 stick = ½ cup = 8 tablespoons; half a stick = ¼ cup.",
    },
    Array.from({ length: count }, (_, s) =>
      el(
        "div",
        { class: "stick" },
        el("div", { class: "stick-halves" }, el("span", {}, "½ stick = ¼ c"), el("span", {}, "½ stick = ¼ c")),
        el(
          "div",
          { class: "stick-marks" },
          Array.from({ length: 8 }, (_, m) => el("span", { class: `mark${s * 8 + m < marks ? " on" : ""}` }, "1 T")),
        ),
        el("div", { class: "stick-name" }, "1 stick = ½ c = 8 T"),
      ),
    ),
  );
}

function picture(v, amount) {
  if (v.kind === "butter") return butterSticks(amount, Math.max(1, Math.ceil(Math.max(v.given, v.answer) / 192)));
  return cupWall(amount);
}

// ---------- Cheat sheet ----------

function renderCheatSheet() {
  const two = (rows) =>
    el(
      "table",
      { class: "ref-table" },
      el(
        "tbody",
        {},
        rows.map(([a, b]) => el("tr", {}, el("th", {}, a), el("td", {}, b))),
      ),
    );
  $("cheat-sheet").replaceChildren(
    el("header", { class: "sheet-head" }, el("h2", {}, "Recipe Math Cheat Sheet")),
    el(
      "div",
      { class: "cheat-grid" },
      el(
        "section",
        {},
        el("h3", {}, "Read the recipe"),
        two(REF.abbrevs),
        el("h3", {}, "Grab the right tool"),
        two(REF.tools),
      ),
      el(
        "section",
        {},
        el("h3", {}, "Know these cold"),
        el(
          "ul",
          { class: "fact-list" },
          REF.facts.map((f) => el("li", {}, f)),
        ),
        el("h3", {}, "The Gallon Kingdom"),
        gallonDiagram(),
        el("p", { class: "caption" }, "1 gal = 4 qt = 8 pt = 16 c"),
      ),
    ),
    el(
      "div",
      { class: "cheat-grid pics" },
      el(
        "section",
        {},
        el("h3", {}, "The cup wall"),
        cupWall(0),
        el("p", { class: "caption" }, "Notice: ⅓ c doesn't line up with the T blocks. ⅓ c = 5 T + 1 t."),
      ),
      el(
        "section",
        {},
        el("h3", {}, "Butter"),
        butterSticks(0, 1),
        el("p", { class: "caption" }, "1 stick = ½ c = 8 T = 4 oz. 2 sticks = 1 c."),
      ),
    ),
    el("h3", {}, "Chef's rules"),
    el(
      "ul",
      { class: "tips" },
      REF.tips.map((t) => el("li", {}, t)),
    ),
  );
}

function gallonDiagram() {
  const cup = () => el("div", { class: "g-cup" }, "C");
  const pint = () => el("div", { class: "g-pint" }, el("span", {}, "P"), el("div", { class: "g-row" }, cup(), cup()));
  const quart = () =>
    el("div", { class: "g-quart" }, el("span", {}, "Q"), el("div", { class: "g-row" }, pint(), pint()));
  return el(
    "div",
    {
      class: "g-gallon",
      role: "img",
      "aria-label": "A gallon holds 4 quarts. Each quart holds 2 pints. Each pint holds 2 cups.",
    },
    el("span", {}, "G"),
    el("div", { class: "g-quarts" }, quart(), quart(), quart(), quart()),
  );
}

// ---------- Startup ----------

function wireUp() {
  for (const btn of document.querySelectorAll("[role=tab]")) {
    btn.addEventListener("click", () => showTab(btn.dataset.tab));
  }
  $("sound-btn").addEventListener("click", () => {
    sound.mode = sound.modes[(sound.modes.indexOf(sound.mode) + 1) % sound.modes.length];
    store.set("sound", sound.mode);
    renderSoundButton();
    sound.correct(); // preview the volume
  });
  $("rank-list").addEventListener("click", (e) => {
    const card = e.target.closest("[data-rank]");
    if (card) startShift(Number(card.dataset.rank));
  });
  $("ticket").addEventListener("submit", (e) => {
    e.preventDefault();
    if (S.phase === "review") nextOrder();
    else serve({});
  });
  $("hint-btn").addEventListener("click", useHint);
  $("skip-btn").addEventListener("click", skipOrder);
  $("next-btn").addEventListener("click", nextOrder);
  $("bowl-clear").addEventListener("click", () => {
    if (S.phase === "answer") {
      S.bowl = [];
      renderBowl();
      renderGlass(cur());
    }
  });
  $("quit").addEventListener("click", () => {
    if (confirm("Leave the line? This shift won't count.")) goHome();
  });
  $("home-btn").addEventListener("click", goHome);
  $("again-btn").addEventListener("click", () => startShift(S.rank.id));
  $("up-btn").addEventListener("click", () => startShift(S.rank.id + 1));
  // Keyboard: A–E or 1–5 pick a choice; Enter moves on after a correct serve.
  document.addEventListener("keydown", (e) => {
    if ($("shift").hidden || $("kitchen").hidden || S.over) return;
    if (e.target.tagName === "INPUT" || e.metaKey || e.ctrlKey || e.altKey) return;
    const o = cur();
    if (S.phase === "review" && e.key === "Enter") {
      e.preventDefault();
      nextOrder();
      return;
    }
    if (o.input !== "choice" || S.phase !== "answer") return;
    let i = "1234567890".indexOf(e.key);
    if (i < 0) i = "abcdefghij".indexOf(e.key.toLowerCase());
    if (i >= 0 && i < o.choices.length) {
      e.preventDefault();
      pickChoice(i);
    }
  });

  // Worksheet controls
  const saved = store.get("ws", {});
  $("ws-ranks").replaceChildren(
    ...RANKS.map((r) =>
      el(
        "label",
        { class: "check" },
        el("input", { type: "radio", name: "ws-rank", value: String(r.id), checked: (saved.rank || 1) === r.id }),
        ` ${r.id}. ${r.name}`,
      ),
    ),
  );
  $("ws-title").value = saved.title ?? "";
  $("ws-version").value = String(saved.version || randomVersion());
  const saveWs = () => store.set("ws", { ...wsSettings(), title: $("ws-title").value });
  $("ws-form").addEventListener("input", () => {
    saveWs();
    renderWorksheet();
  });
  $("ws-new").addEventListener("click", () => {
    $("ws-version").value = String(randomVersion());
    saveWs();
    renderWorksheet();
  });
  $("ws-print").addEventListener("click", () => printSheet("sheet"));
  $("ws-print-key").addEventListener("click", () => printSheet("key"));
  $("cheat-print").addEventListener("click", () => window.print());
}

(async function main() {
  renderSoundButton();
  try {
    await loadEngine();
    RANKS = call("ranks");
    REF = call("reference");
  } catch (err) {
    $("loading").hidden = true;
    $("load-error").hidden = false;
    $("load-error").textContent =
      `Sorry, this browser couldn't start the game (${err.message}). Try an up-to-date Chrome, Edge, Safari, or Firefox.`;
    return;
  }
  wireUp();
  renderRanks();
  renderWorksheet();
  renderCheatSheet();
  $("loading").hidden = true;
  showTab(store.get("tab", "kitchen"));
  const deep = Number(params.get("rank"));
  if (RANKS.some((r) => r.id === deep)) {
    showTab("kitchen");
    startShift(deep);
  }
})();
