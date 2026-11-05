# 🛎️ Recipe Math: The Line

A kitchen game for high school (grades 10–12) Family & Consumer Sciences foods classes. Students work shifts on a ticket rail and climb the brigade from Dish Pit to Chef de Cuisine. Along the way they learn to read recipe notation, grab the right tool, measure with real scoops, and scale whole recipes, mise en place style.

The app is **one file**: `dist/recipe-math.html`. Open it in any browser on a Chromebook, phone, or laptop. It works offline. Every shift also prints as a worksheet with an answer key.

## How a shift works

- **🍽 Plates are your life** (4 by default). A wrong serve is an accident: a plate clatters and you lose one. The feedback asks a question that points at the mistake without giving away the answer.
- **💡 Hints are questions** (Socratic), and each costs ½ plate. The last hint sets up the math but never gives the number.
- **↪ Skip** sends a ticket to the back of the rail for free. Skipping the same ticket again costs ½ plate.
- **✅ Correct** wins back ½ plate and earns tips (more on the first try, plus a bonus for the fewest scoops).
- **⏱ The shift clock** counts down. Finishing early earns tips. Running late never costs plates; the time bonus just stops.
- **Sounds:** an order-up bell, a soft tick for correct answers, a plate clatter for accidents, a chime for a finished recipe card. All are synthesized in the browser. The 🔉 button cycles soft / full / off.
- Lose every plate and **Chef pulls you off the line**: a review of what tripped you up, then a retry.
- **🔥 Hot line:** 3+ clean answers in a row earn a streak bonus. An accident or a hint quietly resets it.
- **Clean answers move on by themselves** after a moment. After an accident or a hint, the student taps Next, so they read the explanation.
- **Chef gear:** tips add up across shifts and unlock cosmetic gear (Paper cap → Apron → Side towel → Chef's knife → Toque → Copper pan → Golden whisk). Each station also tracks a personal best.
- **On phones**, the mixing bowl becomes a sticky tray at the bottom with its own Serve button.

## Stations

| # | Station | What's on the rail |
|---|---|---|
| 1 | Dish Pit (beginner, no clock) | Unit Swap only: one unit to another (6 t = 2 T, ½ c = 8 T, 2 sticks = 1 c, 2 lb = 32 oz) |
| 2 | Commis | Read the Recipe (T vs. t, C., oz vs. fl oz, pinch, packed) · Walk-in Facts |
| 3 | Prep Cook | + Grab the Tool (dry cups / liquid cup / spoons / scale / butter wrapper) · Measure It |
| 4 | Line Cook | + Conversion Factor · Scale a Line |
| 5 | Sous Chef | + **Recipe Card**: factor → every line → mise en place · Chef's Call (eggs, leavening, salt, pans) |
| 6 | Chef de Cuisine | + Fix the Line Cook's Mistake: a side-by-side ledger of original → what the cook wrote |

Recipe cards are mostly ×½ or ×2, sometimes ×3 or ×4. They're written in the styles real recipes use: lab handouts (`1 ½ C.`, `3 Tbsp. butter, softened`), recipe sites (`1 1/2 cups`, `2 large eggs`), and classic cookbooks (`1½ c`, `2 T`, `1 t`). See [`docs/recipe-notation.md`](docs/recipe-notation.md).

Choices work tap-to-pick, then Serve (or tap the same choice again), so a slip of the finger doesn't cost a plate.

Every shift mixes order types, so it never drills one conversion over and over.

**Measure It** uses the real drawer: spoons ⅛ t–1 T, dry cups ¼–1 c, a glass liquid cup with ¼ and ⅓ lines, and butter sticks with tablespoon marks. Any exact combination counts. The bowl lists the scoops but never adds them up; that's the student's job.

## Teacher settings (add to the URL)

| Setting | Effect |
|---|---|
| `?plates=6` | More (or fewer) plates per shift, from 1 to 10 |
| `?timer=off` | Hide the clock |
| `?sound=off` | Start with sound off (students can still change it) |
| `?rank=1` | Open straight into a station's shift (1–6) |
| `?auto=off` | Always wait for Next instead of moving on after clean answers |

Combine them: `recipe-math.html?plates=6&timer=off`.

## Background

- [`docs/research.md`](docs/research.md): what students and home cooks struggle with, plus the learning science behind the design.
- [`docs/chef-critique.md`](docs/chef-critique.md): an executive chef's roast of v1 and of the v2 plan.
- [`docs/v2-design.md`](docs/v2-design.md): the design, and how the teacher's rules and the chef's advice were combined.
- [`docs/recipe-notation.md`](docs/recipe-notation.md): how real recipes write amounts and units.
- [`docs/ux-critique.md`](docs/ux-critique.md): a UX expert's roast of v2.

## How it's built

All the math is Go, compiled to WebAssembly and bundled into the HTML file:

- `frac/`: exact fractions; reads answers like `2.5`, `2 1/2`, `5/2`, or `½`.
- `kitchen/`: tools and the fewest-scoops solver, ingredients, real recipes, order generators with Socratic hint ladders, shifts (seeded, so worksheets can be recreated), and answer checking with targeted feedback.
- `cmd/wasm/`: exposes the engine to the page as `window.recipeMath`.
- `cmd/bundle/`: inlines everything into `dist/recipe-math.html`.
- `web/`: the page (HTML, CSS, plain JavaScript, sounds, print layouts).

Requires Go 1.24+.

```sh
make test    # Go tests
make build   # → dist/recipe-math.html
make open    # build and open in your browser
```
