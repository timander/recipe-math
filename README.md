# 🥄 Recipe Math: Units & Conversions

Practice kitchen unit conversions on a Chromebook or phone, or print worksheets with answer keys.
Built for high school (grades 10–12) Family & Consumer Sciences foods classes.

The finished app is **one file**: `dist/recipe-math.html`. Open it in any browser. It works offline, so you can email it or put it on a USB stick.

## What's in it

- **Practice:** 10-problem rounds at six levels. A wrong answer gets a nudge, such as "check your direction" or "good start, keep going," and the student can try again. **Hint** shows which equivalents to use. After each problem, the steps are shown. Right on the first try = 1 point; with help = ½. Up to 3 stars per level.
- **Worksheet:** Pick levels and how many problems, then print a student copy and an answer key (with the work shown). The **version number** picks the problems, so the same version always prints the same worksheet and key.
- **Cheat Sheet:** A one-page reference, including the "Gallon Kingdom" picture (1 gallon = 4 quarts = 8 pints = 16 cups).

## Units

| Unit | Written as | Measures |
|---|---|---|
| teaspoon | tsp, t | volume |
| tablespoon | Tbsp, T | volume |
| fluid ounce | fl oz | volume |
| cup | cup, c | volume |
| pint | pint, pt | volume |
| quart | quart, qt | volume |
| gallon | gallon, gal | volume |
| ounce | oz | weight |
| pound | lb | weight |
| milliliter | mL | volume |
| liter | L | volume |
| gram | g | weight |
| kilogram | kg | weight |

**Exact:** 1 Tbsp = 3 tsp · 1 fl oz = 2 Tbsp · 1 cup = 16 Tbsp = 8 fl oz · 1 pint = 2 cups · 1 quart = 2 pints = 4 cups · 1 gallon = 4 quarts · 1 lb = 16 oz · 1 L = 1000 mL · 1 kg = 1000 g

**US ↔ metric (rounded, as on US food labels):** 1 tsp ≈ 5 mL · 1 Tbsp ≈ 15 mL · 1 fl oz ≈ 30 mL · 1 cup ≈ 240 mL · 1 oz ≈ 28 g · 1 lb ≈ 454 g · 1 kg ≈ 2.2 lb. Answers within 3% count, so the true values (1 cup = 237 mL) are accepted too.

## Levels

| # | Level | Example |
|---|---|---|
| 1 | Spoons: tsp ↔ Tbsp, whole numbers | 2 Tbsp = ___ tsp |
| 2 | Spoons to Cups: Tbsp, fl oz, cups (¼, ½, ¾) | 1/2 cup = ___ Tbsp |
| 3 | Big Containers: cups, pints, quarts, gallons | 1 gallon = ___ quarts |
| 4 | Weight & Metric: oz ↔ lb, g ↔ kg, mL ↔ L | 1 1/2 lb = ___ oz |
| 5 | Two Steps & Thirds: tsp ↔ cups, etc. | 1/3 cup = ___ Tbsp |
| 6 | US ↔ Metric (≈) | 1 cup ≈ ___ mL |

Levels are defined in `problems/levels.go`. Each one lists its unit pairs and which fractions are allowed, so a level never asks for something like 1/3 gallon or 0.7 tsp.

## How it's built

All the math is Go, compiled to WebAssembly and bundled into the HTML file:

- `frac/`: exact fractions, so 1/3 × 3 = 1 exactly. Reads answers like `12`, `3/4`, `1 1/2`, `1-1/2`, `½`, `1.5`, or `12 Tbsp`.
- `units/`: the units, the equivalents, conversions, and the shortest chain of known equivalents used to explain an answer.
- `problems/`: the levels, the problem generator (seeded, so worksheets can be recreated), the worked steps, and answer checking with common-mistake feedback.
- `cmd/wasm/`: exposes the engine to the page as `window.recipeMath`.
- `cmd/bundle/`: inlines everything into `dist/recipe-math.html`.
- `web/`: the page (HTML, CSS, plain JavaScript; screen and print layouts).

Requires Go 1.24+.

```sh
make test    # Go tests
make build   # → dist/recipe-math.html
make open    # build and open in your browser
```
