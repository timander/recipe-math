# Executive chef critique of v1 and the teacher's v2 plan

Written 2026-11-04 by a separate Claude agent playing an old-school executive chef who expects cooks to know recipe math quickly and accurately. It was handed v1's code and the teacher's play-testing feedback. This is its full report.

## 1. The Roast

1. **"1/3 cup = ___ Tbsp" → 5 1/3.** Show me the 1/3-tablespoon in my drawer. I'll wait. The real answer is **5 T + 1 t**, and v1 never asks for it. The engine allows Tbsp fractions in thirds, so it can ask for "2/3 Tbsp". That's a teaspoon pair wearing a fake mustache.
2. **Level 5 asks cup ↔ tsp, pint ↔ Tbsp, gallon ↔ cup, quart ↔ fl oz.** tsp goes up to 48 and Tbsp to 32. Nobody on my line has ever measured 48 teaspoons of anything. If you're counting spoons past 4 T, you grabbed the wrong tool. That's a judgment failure, not a math skill.
3. **Level 6 accepts 237 mL for a cup.** Sweet, but no 10th grader's liquid cup has a 237 line. The kitchen skill is "read the mL side of the Pyrex", not mental arithmetic within 3%.
4. **The fl oz vs. oz lesson lives in a cheat-sheet tip and nowhere else.** No problem ever makes a kid choose between the liquid cup and the scale. That's the mistake that wrecks cookies, and v1 tests it zero times.
5. **The hint is the answer in an apron.** "Use 1 cup = 16 Tbsp. A cup is bigger than a tablespoon… Multiply." That isn't a hint, it's the recipe card. On top of that, retries are unlimited and a hint still earns ½ point, so a kid can mash numbers until it turns green.
6. **Credit where it's due:** the "check your direction" detector and the "good start, keep going" halfway catch are the best things in this build. That's exactly what I yell across the line. Keep them, and promote them to the main feedback system. Seeded worksheets with matching answer keys are also chef's kiss. Paper never crashes.
7. **The teacher is right about the core problem:** recipes say *T*, *t*, and *c*, and cooks think in scoops, not decimals. She's naive about the dopamine, though. Thirty Chromebooks each playing breaking glass sounds like a dish pit falling down a staircase. The teacher will mute it by Wednesday.
8. **A countdown timer on kids who already freeze at fractions** teaches one thing: guess fast. I time my cooks *after* they're accurate. Speed is the reward for knowing it, not the test of whether they do.
9. **"Skips are free, but skipping the same question twice costs ½."** That's a tax code, not a game rule. If you need a footnote to explain the penalty, the 10th grader is already gaming it.
10. **"Learning requires struggle." Correct.** But struggle means *working the problem*, not *losing a life for trying*. If a wrong answer and a skip cost the same, every smart kid skips.

## 2. What a Real Kitchen Needs (ranked, day one)

| # | Skill | HS game? |
|---|---|---|
| 1 | **Scaling factor**: desired yield ÷ original yield, applied to every line. 24 cookies → 60 = ×2.5. | **YES, the spine** |
| 2 | **Converting to tool combinations**: 2.5 × ¾ c = 1⅞ c → 1½ c + ¼ c + 2 T. 2.5 × 1 t = 2½ t. | **YES** |
| 3 | **Right tool for the ingredient**: dry cups for flour/sugar, liquid cup for milk/stock, scale for chocolate in oz, spoons for leavening. fl oz ≠ oz. | **YES** |
| 4 | **Consolidating up**: 12 T → ¾ c, 6 c → 1½ qt, 24 oz → 1½ lb. Fewer scoops means fewer mistakes. | **YES** |
| 5 | **Eyeball sanity check**: "Does 8 t of salt for one pan of brownies sound right?" Estimate before you calculate. | **YES** |
| 6 | **Rounding judgment**: eggs (2.5 → round to 3, or beat and weigh about 50 g each), baking soda/powder, salt, and spices go conservative. Pan size and bake time don't scale linearly. | **YES, as "Chef's Call" questions, not formulas** |
| 7 | **As-purchased vs. edible portion**: need 2 lb peeled potatoes at 80% yield → buy 2½ lb. | **Stretch level for seniors** |
| 8 | **Costing** (cost per portion) | **Not in this game.** Separate unit. |
| — | Baker's percentages, volume↔weight by ingredient density | **No** |

## 3. Typical Quantities (what's really in the drawer)

**Measuring spoons:** ⅛ t, ¼ t, ½ t, 1 t, 1 T. Some sets add ½ T and 1½ t. "Pinch" and "dash" mean "less than ⅛ t": no conversion, just judgment.
- Recipe amounts: ⅛–2 t (salt, baking soda/powder, vanilla, spices), 1–3 T (butter, oil, sugar, flour for a roux). Past 4 T, recipes switch to ¼ c.

**Dry cups:** ¼, ⅓, ½, 1 c. Some sets add ⅔ and ¾.
- Amounts: ¼, ⅓, ½, ⅔, ¾, 1, 1¼, 1½, 2, 2¼, 3 c (flour, sugar, oats).

**Liquid cups:** 1 c, 2 c, 4 c (1 qt) glass, marked in c, oz, and mL.
- Amounts: ¼ c to 1 qt (milk, water, stock, oil).

**Weight:**
- oz: 4 oz and 8 oz (cheese, cream cheese, chocolate)
- lb: 1, 1½, 2, 5 lb (meat, potatoes)
- Cans: 14.5 oz, 15 oz, 28 oz

**The one butter conversion everybody needs:** 1 stick = ½ c = 8 T = 4 oz = 113 g.

**Conversions cooks actually do:**
- T ↔ t (3 t = 1 T)
- T ↔ cup (4 T = ¼ c, 5 T + 1 t = ⅓ c, 8 T = ½ c, 12 T = ¾ c, 16 T = 1 c)
- cups ↔ pint/quart/gallon when scaling stock, milk, or soup
- oz ↔ lb
- sticks ↔ T/cups/oz
- fl oz ↔ cups on the liquid cup
- reading grams/mL straight off the scale or cup

**Textbook trivia: kill it.** gallon → tsp, pint → Tbsp, cup → tsp, kg ↔ lb mental math, mL ↔ cup to 3%. ⅓ T is not a thing.

## 4. Verdict on Each Mechanic

**Countdown timer: KILL it in learning mode, CHANGE it to an optional Rush mode.** A per-problem countdown punishes the slow-but-careful kid and rewards guessing. In learning mode, run a visible ticket clock that *counts up* and report "ticket time" at the end, so kids can beat their personal best. Unlock **Rush** once a student scores ≥80% on that station. Even then the timer gives bonus points; it never takes lives.

**Life bar: KEEP it, but rebalance it so it doesn't punish trying.** Starting numbers:
- **4 plates** per ticket of 8 orders.
- **First wrong answer on a problem:** targeted feedback (the v1 direction and halfway detectors), no penalty.
- **Second wrong answer:** accident, −1 plate.
- **Hint:** −½ plate, but the **first rung of each problem's hint ladder is free**.
- **Correct on the first try:** heals ½ plate, up to 4.
- **Zero plates:** "Chef pulls you off the line." Show the worked steps for everything missed, offer a replay, and keep the student's progress.

**Accident sound: CHANGE it.** Use a soft plate clatter, not breaking glass. Breaking glass startles kids, and kids with sensory issues will hate it.

**Hints cost ½: KEEP, with the free first rung above.**

**Socratic hints: KEEP, and do them properly.** Each rung asks one question. The last rung shows the setup but never the final number. Example ladders:
- **¾ c = ___ T**
  1. "How many T fill a ¼ cup?"
  2. "How many ¼ cups are in ¾ cup?"
  3. "So that's 3 scoops of 4 T. What's 3 × 4?"
- **Measure 5 T with real tools**
  1. "What's the biggest tool that fits in 5 T without going over?"
  2. "¼ c = 4 T. How much is left?"
  3. "Which spoon holds that?"
- **Recipe serves 4, you need 10. 2 T butter = ?**
  1. "Is your batch bigger or smaller? So more or less than 2 T?"
  2. "How many times bigger is 10 than 4?"
  3. "2 × 2½ = ? Now turn that into tools."
- **"8 oz chocolate chips": use the liquid cup's 8 oz line?**
  1. "Is that 8 ounces of *space* or 8 ounces of *weight*?"
  2. "Which tool tells you weight?"
- **Recipe has 3 eggs, scaled to ×½**
  1. "Can you crack 1½ eggs?"
  2. "What would a chef do: round, or beat and measure?"

**Skip rule: CHANGE it.** A skip is free, and the ticket goes to the back of the rail ("86'd, comes back later"). An order still unfinished when the ticket closes counts as a miss. One rule, no footnotes.

**Order-up bell: KEEP it.** It plays once per round, so it's cheap and fun.

**Correct-answer sound: KEEP it, but quiet.** Make it a soft "tick" or pan sizzle under 300 ms. Save a bigger sound for completing a full ticket.

**Sound fatigue fix:**
- Sound is **off by default on Chromebooks**, with a teacher setting (URL flag) of off / soft / full.
- A mute toggle that remembers its setting.
- Headphones encouraged.
- No sound on keystrokes, ever.

## 5. Gameplay to Build

**A. The Ticket Rail (scaling)**
- How it plays: a ticket drops: "Choc chip cookies, yields 24, need 60." The student enters the factor (2.5). Then each ingredient line flips up in turn, and the student builds the answer by tapping tool tokens from a drawer (¼ c, 1 T, ½ t…). Multiple valid combos are accepted, and fewer scoops scores higher.
- Trains: skills 1, 2, 4.
- Why it's not repetitive: every recipe brings different yields, factors (×½, ×1½, ×3, ×2.5), units, and ingredients.

**B. Mise en Place Station**
- How it plays: the student gets a scaled recipe and drags the *right tool* onto each ingredient's prep bowl: liquid cup for milk, dry cups for flour, scale for 8 oz chocolate, spoons for soda. Then they set the amount or the count of scoops. Wrong tool = accident.
- Trains: liquid vs. dry vs. weight, and tool combos.
- Why it's not repetitive: the decisions are physical and varied, not "type a number".

**C. Fix the Line Cook's Mistake**
- How it plays: a scaled recipe card has 1–3 planted errors (salt scaled ×8 to 16 t, "2⅝ eggs", flour measured in the liquid cup, "1 c = 8 oz chocolate chips", ⅓ c written as "5.33 T"). The student taps each error, fixes it, and picks a reason.
- Trains: sanity checks, rounding judgment, critical reading.
- Why it's not repetitive: it's detective work, and the errors rotate.

**D. Banquet Prep (12th-grade boss level)**
- How it plays: scale a soup or cookie recipe to 120 servings. Consolidate units (48 T butter → 3 c → 6 sticks → 1½ lb), then write the buy list using AP/EP yields.
- Trains: big scaling, consolidating up, purchasing.
- Why it's not repetitive: it's one big project per session, not a drill.

**Keep v1's drill as a 60-second "Walk-in Warm-up"** with realistic amounts only. It's for memorizing the 8 core facts, nothing more.

## 6. Non-Negotiables (miss one and you're fired)

1. **Answers are tool combinations** in real recipe notation (T, t, c, oz, lb), with every valid combo accepted. "5⅓ T" is never the final answer.
2. **Realistic quantities only.** Cap every unit to what recipes actually call for. Delete gallon → tsp, pint → Tbsp, cup → tsp, and thirds of a Tbsp.
3. **The ingredient decides the tool.** The liquid/dry/weight choice and the fl oz vs. oz trap appear in every session, not only on a cheat sheet.
4. **Whole-recipe scaling** (desired ÷ original → every line → measurable tools) is the core loop and the end goal, not a bonus level.
5. **Hints ask, they don't tell.** Laddered Socratic questions; the last rung shows the setup, never the number.
6. **A wrong answer earns targeted feedback before any penalty.** Keep v1's direction and halfway detectors. No countdown in learning mode.
7. **Chef's Call judgment items** (eggs, leavening, salt, spices) appear regularly, accepting a range or a reason, not one number.
8. **Classroom-safe sound and paper parity.** Sound is off or soft by default and teacher-controlled. Every mode prints as a recipe-card worksheet with a seeded answer key.

*Now get back on the line. Tickets are piling up.*
