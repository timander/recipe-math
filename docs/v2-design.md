# v2 design: "The Line"

The student works shifts in a kitchen and climbs the brigade. Each shift is a rail of tickets (orders) that mixes problem types, so it never drills one conversion over and over. The end goal is the last station: scaling a whole recipe card, mise en place style.

## Ranks (stations)

| Rank | Station | Order types in the shift |
|---|---|---|
| 1 | Dish Pit | Read the Recipe (T vs. t, c, oz vs. fl oz, pinch, packed) · Walk-in Facts (the core equivalents) |
| 2 | Prep Cook | + Grab the Tool (dry cups / liquid cup / spoons / scale / butter wrapper) · Measure It (build an amount from tool tokens) |
| 3 | Line Cook | + Conversion Factor (want ÷ makes) · Scale a Line (scale one ingredient, then measure it) |
| 4 | Sous Chef | + Recipe Card (whole recipe: factor → every line → mise en place) · Chef's Call (eggs, leavening, salt, pans, bake time) |
| 5 | Chef de Cuisine | + Fix the Line Cook's Mistake · harder factors (×1½, ×2½, ×¾, ×⅔) |

Every shift also mixes in some orders from earlier ranks (spaced review). All ranks are unlocked, so a teacher can send a class straight to any station.

## Answers are real kitchen actions
- **Measure It** uses a drawer of real tools: spoons ⅛ t, ¼ t, ½ t, 1 t, 1 T; dry cups ¼, ⅓, ½, 1 c. Any exact combination is correct. Matching the fewest possible scoops earns a bonus. The bowl lists the scoops but **never shows the running total**, so students add it up themselves.
- **Liquids** use a glass liquid cup: tap a fill line, then add spoons for anything extra.
- **Butter** uses sticks and the tablespoon marks on the wrapper.
- **Weight** is typed in oz or lb, as read off the scale.
- **Recipe notation:** T, t, c, oz, lb, with Tbsp/tsp/cup spelled out sometimes so students see both.
- **Quantities are realistic:** what recipes actually call for (⅛ t to 3 T, ¼ c to 4 c). No gallon → teaspoon trivia.

## Your rules vs. the chef's: how I settled them

| Mechanic | Your plan | Chef | v2 |
|---|---|---|---|
| Life bar | tolerates n accidents | about 4 plates | **4 plates** by default (teacher can set `?plates=n`) |
| Mistake | accident + dropped plate | first miss free, soft clatter | **Every wrong serve is an accident** (−1 plate, plate-clatter sound), **but** the feedback is a targeted Socratic question, so the next try is informed |
| Hint | costs ½ | ½, first rung free | **½ plate per rung**. Rungs are questions. The last rung is the setup, never the number |
| Skip | free; same question twice = ½ | goes to back of the rail | **Both:** a skip sends the ticket to the back of the rail for free; skipping it a second time costs ½ plate |
| Correct | heals | heals ½ | **+½ plate** (up to full) and a subtle "tick" |
| Countdown | yes | not in learning mode | **A shift clock counts down.** Beating it earns tips (bonus points). Running out never costs plates; the clock turns red ("tickets are late") and tips stop. Teachers can hide it (`?timer=off`) |
| Bell | order-up bell to start | keep | **Order-up bell** at the start of each shift; a short chime when a ticket is cleared |
| Sound volume | dopamine | soft or off by default | **Soft by default**, with a 3-way toggle (soft / full / off) that remembers its setting. `?sound=off` for whole-class settings |

Losing all plates means "Chef pulled you off the line": a review screen with every missed order worked out, then a retry. Clearing a shift earns 1–3 stars, based on plates left and on-time service.

## Paper parity
The Worksheet tab prints any station's shift as a paper ticket packet. Choices print as circles, Measure It as "write the tools", and recipe cards as a scaling table (ingredient | recipe | × factor | measure with). The version number recreates the matching answer key.
