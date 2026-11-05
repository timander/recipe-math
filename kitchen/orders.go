package kitchen

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/timander/recipe-math/frac"
)

// Kind is the type of order on the ticket rail.
type Kind string

const (
	KSwap    Kind = "swap"    // Unit Swap (beginner)
	KVocab   Kind = "vocab"   // Read the Recipe
	KFact    Kind = "fact"    // Walk-in Facts
	KTool    Kind = "tool"    // Grab the Tool
	KMeasure Kind = "measure" // Measure It
	KFactor  Kind = "factor"  // Conversion Factor
	KScale   Kind = "scale"   // Scale a Line
	KCard    Kind = "card"    // one line of a Recipe Card
	KCall    Kind = "call"    // Chef's Call
	KFix     Kind = "fix"     // Fix the Line Cook's Mistake
)

var stationName = map[Kind]string{
	KSwap: "Unit Swap", KVocab: "Read the Recipe", KFact: "Walk-in Facts", KTool: "Grab the Tool", KMeasure: "Measure It",
	KFactor: "Conversion Factor", KScale: "Scale a Line", KCard: "Recipe Card", KCall: "Chef's Call",
	KFix: "Fix the Mistake",
}

var seconds = map[Kind]int{
	KSwap: 25, KVocab: 20, KFact: 20, KTool: 20, KMeasure: 45, KFactor: 40, KScale: 60, KCard: 45, KCall: 30, KFix: 60,
}

// Input is how the student answers.
type Input string

const (
	InChoice Input = "choice" // tap one choice
	InBuild  Input = "build"  // tap tools into the bowl
	InNumber Input = "number" // type a number
)

// Card is a recipe card shown beside its orders.
type Card struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Makes     int        `json:"makes"`
	MakesWhat string     `json:"makesWhat"`
	Want      int        `json:"want"`
	Factor    string     `json:"factor"` // "1½"
	Lines     []CardLine `json:"lines"`
}

// CardLine is one ingredient as written, and as it should be after scaling.
type CardLine struct {
	Text   string `json:"text"`   // "¾ c sugar"
	Scaled string `json:"scaled"` // "1 c + ⅛ c"... the measured answer
	Tool   string `json:"tool"`   // "dry cups", shown on Fix cards
}

// Order is one ticket: a single question with one answer.
type Order struct {
	ID      int      `json:"id"`
	Kind    Kind     `json:"kind"`
	Station string   `json:"station"`
	Prompt  string   `json:"prompt"`
	Context string   `json:"context,omitempty"`
	Card    *Card    `json:"card,omitempty"`
	Line    int      `json:"line"` // which card line this order fills (-1: none)
	Input   Input    `json:"input"`
	Choices []string `json:"choices,omitempty"`
	Drawer  Drawer   `json:"drawer,omitempty"`
	Tools   []Tool   `json:"tools,omitempty"`
	Unit    string   `json:"unit,omitempty"` // shown after a number box
	Seconds int      `json:"seconds"`
	Hints   []string `json:"hints"`
	Explain []string `json:"explain"`
	Answer  Answer   `json:"answer"`
	Visual  *Visual  `json:"visual,omitempty"`
}

// Answer is the key. The page needs it for the printed answer key, and a
// determined student could find it, which is fine for a practice game.
type Answer struct {
	Choices    []int    `json:"choices,omitempty"`
	Vol        int      `json:"vol,omitempty"`
	Num        [2]int64 `json:"num,omitempty"`
	Text       string   `json:"text"`
	BestScoops int      `json:"bestScoops,omitempty"`
	Best       []string `json:"best,omitempty"` // tool IDs for one fewest-scoops answer
	// For catching common slips while scaling.
	Given  [2]int64 `json:"given,omitempty"` // original amount (eighths, oz, or count)
	Factor [2]int64 `json:"factor,omitempty"`
	Makes  int      `json:"makes,omitempty"`
	Want   int      `json:"want,omitempty"`
}

func pair(f frac.Frac) [2]int64   { return [2]int64{f.Num(), f.Den()} }
func unpair(p [2]int64) frac.Frac { return frac.New(p[0], max(p[1], 1)) }
func shuffled[T any](rng *rand.Rand, xs []T) []T {
	out := slices.Clone(xs)
	rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// choices shuffles right and wrong answers and records where the right
// ones landed.
func choices(rng *rand.Rand, right, wrong []string) ([]string, []int) {
	all := shuffled(rng, append(slices.Clone(right), wrong...))
	var idx []int
	for i, c := range all {
		if slices.Contains(right, c) {
			idx = append(idx, i)
		}
	}
	return all, idx
}

// ---------- Ranks ----------

// Rank is a station in the brigade. Each shift mixes its order types.
type Rank struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Goal    string `json:"goal"`
	Example string `json:"example"`
	Timed   bool   `json:"timed"` // beginners work without a clock
}

var Ranks = []Rank{
	{1, "Dish Pit", "Beginner: swap one unit for another. No clock.", "6 t = ___ T", false},
	{2, "Commis", "Learn the language: T vs. t, c, oz vs. fl oz, and the core equivalents.", "“2 Tbsp.” means 2 ___", true},
	{3, "Prep Cook", "Grab the right tool for each ingredient, and measure amounts with real scoops.", "5 T = ¼ c + 1 T", true},
	{4, "Line Cook", "Find the conversion factor and scale single ingredients.", "Makes 12, need 24 → ×2", true},
	{5, "Sous Chef", "Scale a whole recipe card, mise en place style, and make the judgment calls.", "Cookies ×½, every line", true},
	{6, "Chef de Cuisine", "Run full recipe cards and catch the line cook's mistakes.", "Spot the 2 T that should be 2 t", true},
}

// Real kitchens mostly halve or double a recipe, sometimes ×3 or ×4. Repeats
// in these lists make those factors come up more often.
var (
	easyFactors = []frac.Frac{frac.Int(2), frac.Int(2), frac.New(1, 2), frac.New(1, 2), frac.Int(3), frac.Int(4), frac.New(3, 2)}
	cardFactors = []frac.Frac{frac.New(1, 2), frac.New(1, 2), frac.New(1, 2), frac.Int(2), frac.Int(2), frac.Int(2), frac.Int(3), frac.Int(4)}
)

// Shift builds one shift's ticket rail for a rank. The same seed always
// builds the same shift, so worksheets and answer keys match.
func Shift(rank int, seed uint64) ([]Order, error) {
	rng := newRNG(seed)
	g := &gen{rng: rng}
	var plan []func() []Order
	add := func(n int, f func() []Order) {
		for range n {
			plan = append(plan, f)
		}
	}
	one := func(f func() Order) func() []Order { return func() []Order { return []Order{f()} } }

	switch rank {
	case 1:
		add(10, one(g.swap))
	case 2:
		add(6, one(g.vocab))
		add(4, one(g.fact))
	case 3:
		add(2, one(g.vocab))
		add(2, one(g.fact))
		add(3, one(g.tool))
		add(3, one(func() Order { return g.measure(false) }))
	case 4:
		add(1, one(g.fact))
		add(2, one(g.tool))
		add(2, one(func() Order { return g.measure(true) }))
		add(2, one(func() Order { return g.factor(easyFactors) }))
		add(3, one(func() Order { return g.scale(easyFactors) }))
	case 5:
		add(1, one(g.tool))
		add(1, one(func() Order { return g.measure(true) }))
		add(2, one(g.call))
	case 6:
		add(1, one(func() Order { return g.measure(true) }))
		add(2, one(g.fix))
		add(2, one(g.call))
	default:
		return nil, fmt.Errorf("no rank %d", rank)
	}
	plan = shuffled(rng, plan)
	// The recipe card is the main event: it comes after the warm-up orders.
	switch rank {
	case 5:
		plan = slices.Insert(plan, 2, func() []Order { return g.card(cardFactors) })
	case 6:
		plan = slices.Insert(plan, 3, func() []Order { return g.card(cardFactors) })
	}
	var out []Order
	prompts := map[string]bool{}
	for _, f := range plan {
		var got []Order
		for range 20 { // regenerate if this exact question already came up
			if got = f(); !prompts[got[0].Prompt] {
				break
			}
		}
		prompts[got[0].Prompt] = true
		out = append(out, got...)
	}
	for i := range out {
		out[i].ID = i + 1
		out[i].Station = stationName[out[i].Kind]
		if out[i].Seconds == 0 {
			out[i].Seconds = seconds[out[i].Kind]
		}
		if out[i].Card == nil {
			out[i].Line = -1
		}
	}
	return out, nil
}

// gen remembers what it has used so a shift doesn't repeat itself.
type gen struct {
	rng   *rand.Rand
	used  map[string]bool
	cards int
}

// seen marks a key as used and reports whether it already was.
func (g *gen) seen(key string) bool {
	if g.used == nil {
		g.used = map[string]bool{}
	}
	was := g.used[key]
	g.used[key] = true
	return was
}

// fresh picks a not-yet-used index from n options, keyed by kind.
func (g *gen) fresh(kind string, n int) int {
	if g.used == nil {
		g.used = map[string]bool{}
	}
	for range 50 {
		if i := g.rng.IntN(n); !g.seen(fmt.Sprint(kind, i)) {
			return i
		}
	}
	return g.rng.IntN(n)
}

func (g *gen) vocab() Order {
	it := vocab[g.fresh("vocab", len(vocab))]
	cs, idx := choices(g.rng, []string{it.Right}, it.Wrong)
	return Order{Kind: KVocab, Prompt: it.Q, Input: InChoice, Choices: cs, Hints: it.Hints,
		Explain: []string{it.Explain}, Answer: Answer{Choices: idx, Text: it.Right}}
}

func (g *gen) fact() Order {
	f := facts[g.fresh("fact", len(facts))]
	o := Order{Kind: KFact, Prompt: f.Q, Input: InNumber, Unit: f.Unit, Hints: slices.Clone(f.Hints),
		Explain: []string{f.Explain}, Answer: Answer{Num: pair(frac.Int(f.Answer)), Text: fmt.Sprintf("%d %s", f.Answer, f.Unit)}}
	switch pic := factPics[f.Q]; pic.pic {
	case "wall":
		return withVisual(o, wallVisual(pic.vol, pic.vol))
	case "butter":
		return withVisual(o, butterVisual(pic.vol, pic.vol))
	}
	return o
}

func (g *gen) call() Order {
	it := calls[g.fresh("call", len(calls))]
	cs, idx := choices(g.rng, it.Right, it.Wrong)
	return Order{Kind: KCall, Prompt: it.Q, Input: InChoice, Choices: cs, Hints: it.Hints,
		Explain: []string{it.Explain}, Answer: Answer{Choices: idx, Text: strings.Join(it.Right, " or ")}}
}

// ---------- Grab the Tool ----------

var toolChoices = []string{"Dry measuring cups", "Liquid measuring cup", "Measuring spoons", "Kitchen scale", "Butter wrapper marks"}

// toolFor decides the right tool for an ingredient line.
func toolFor(l Line) int {
	k := ing[l.Ing].Kind
	switch {
	case k == Weight:
		return 3
	case k == Fat:
		return 4
	case l.Vol < QuarterCup:
		return 2
	case k == Wet:
		return 1
	}
	return 0
}

var toolShort = []string{"dry cups", "liquid cup", "spoons", "scale", "butter wrapper"}

// toolLabel names the tool for a card line; eggs are just counted.
func toolLabel(l Line) string {
	if ing[l.Ing].Kind == Count {
		return "count"
	}
	return toolShort[toolFor(l)]
}

var toolHints = [][]string{
	{"Does this ingredient pour, or does it pile up?", "Which cups can you fill to the rim and level off with a straight edge?"},
	{"Does this ingredient pour, or does it pile up?", "Which cup has a spout and lines below the rim so it won't spill?"},
	{"Is this amount more or less than ¼ c, your smallest cup?", "For less than ¼ c, what tools are left in the drawer?"},
	{"Is this “oz” fluid ounces (fl oz) or plain ounces?", "Plain ounces measure weight. Which tool measures weight?"},
	{"How does butter come from the store?", "What's printed on a stick's wrapper?"},
}

// toolWhyNot is the targeted question after picking the wrong tool.
func toolWhyNot(right, chose int) string {
	switch {
	case right == 3 && chose == 1:
		return "That's the fl oz vs. oz trap! Does this say fluid ounces, or plain ounces of weight?"
	case right == 0 && chose == 1:
		return "Can you level off a dry ingredient in a cup with a spout and lines below the rim?"
	case right == 1 && chose == 0:
		return "If you fill a dry cup to the rim with liquid, what happens when you carry it?"
	case right == 2:
		return "Is that amount even as big as your smallest cup (¼ c)?"
	case chose == 2:
		return "Would you really count out that many spoonfuls? Is there a bigger tool?"
	case right == 4:
		return "Look at a stick of butter. What do the lines on the wrapper show?"
	}
	return "What kind of ingredient is it: dry, liquid, butter, or sold by weight?"
}

var toolTraps = []Line{
	w("chipsoz", 8), w("cheese", 8), w("creamcheese", 4), w("beef", 16), v("broth", c(2, 1)),
	v("water", c(3, 4)), v("honey", c(1, 2)),
}

func (g *gen) tool() Order {
	var lines []Line
	for _, r := range Recipes {
		for _, l := range r.Lines {
			if ing[l.Ing].Kind != Count {
				lines = append(lines, l)
			}
		}
	}
	lines = append(lines, toolTraps...)
	lines = append(lines, toolTraps...) // traps show up more often
	l := lines[g.fresh("tool", len(lines))]
	right := toolFor(l)
	text := lineText(l, Short)
	if l.Ing == "broth" {
		text = "16 fl oz chicken broth"
	}
	return Order{Kind: KTool, Prompt: fmt.Sprintf("The recipe says “%s.” What do you grab?", text),
		Input: InChoice, Choices: slices.Clone(toolChoices), Hints: toolHints[right],
		Explain: []string{toolExplain(l, right)}, Answer: Answer{Choices: []int{right}, Text: toolChoices[right]}}
}

func toolExplain(l Line, right int) string {
	name := ing[l.Ing].Name
	switch right {
	case 0:
		return fmt.Sprintf("%s is dry: spoon it into dry measuring cups and level it off.", capitalize(name))
	case 1:
		return fmt.Sprintf("%s is a liquid: liquid measuring cup on the counter, read at eye level.", capitalize(name))
	case 2:
		return "Less than ¼ c: use measuring spoons, filled level."
	case 3:
		return fmt.Sprintf("Plain oz (or lb) is weight. Weigh the %s on a scale, or use the package label. fl oz would be space.", name)
	}
	return "Butter wrappers are marked in tablespoons: 1 stick = 8 T = ½ c."
}

// ---------- Measure It ----------

type target struct {
	ing string
	vol int
}

var measureTargets = map[Drawer][]target{
	DrawerDry: {
		{"sugar", c(3, 4)}, {"flour", c(2, 3)}, {"flour", c(5, 4)}, {"oats", c(3, 2)}, {"flour", c(7, 4)},
		{"flour", c(9, 4)}, {"sugar", tb(5)}, {"brownsugar", tb(6)}, {"cocoa", tb(3)}, {"sugar", tb(10)},
		{"flour", tb(12)}, {"sugar", c(2, 3) + tb(2)}, {"salt", ts(4, 1)}, {"cinnamon", ts(5, 1)},
		{"soda", ts(3, 2)}, {"salt", ts(3, 4)}, {"powder", ts(5, 2)}, {"chili", ts(6, 1)}, {"sugar", c(1, 3) + tb(1)},
		{"cocoa", c(1, 2) + tb(2)},
	},
	DrawerLiquid: {
		{"water", tb(10)}, {"milk", 6 * FlOz}, {"milk", 12 * FlOz}, {"broth", c(5, 4) + tb(2)},
		{"oil", tb(3)}, {"water", c(1, 3) + tb(2)}, {"milk", 10 * FlOz}, {"vanilla", ts(3, 2)},
	},
	DrawerButter: {
		{"butter", tb(6)}, {"butter", c(3, 4)}, {"butter", c(1, 1)}, {"butter", tb(10)}, {"butter", c(3, 2)}, {"butter", tb(12)},
	},
}

// drawerFor picks the drawer an ingredient is measured with.
func drawerFor(id string) Drawer {
	switch ing[id].Kind {
	case Wet:
		return DrawerLiquid
	case Fat:
		return DrawerButter
	}
	return DrawerDry
}

// targetText writes a measuring target, sometimes in a smaller unit so the
// student has to consolidate (12 T → ¾ c) or in fl oz (6 fl oz → ¾ c).
func (g *gen) targetText(t target, consolidate bool) string {
	style := Short
	if g.rng.IntN(3) == 0 {
		style = Abbr
	}
	if consolidate || t.vol < QuarterCup {
		switch {
		case drawerFor(t.ing) == DrawerLiquid && t.vol%FlOz == 0 && t.vol >= QuarterCup:
			return fmt.Sprintf("%d fl oz", t.vol/FlOz)
		case t.vol%Tbsp == 0 && t.vol/Tbsp <= 12 && t.vol/Tbsp > 1:
			return fmt.Sprintf("%d %s", t.vol/Tbsp, map[Style]string{Short: "T", Abbr: "Tbsp"}[style])
		case t.vol%Tsp == 0 && t.vol/Tsp <= 6 && t.vol/Tsp > 1:
			return fmt.Sprintf("%d %s", t.vol/Tsp, map[Style]string{Short: "t", Abbr: "tsp"}[style])
		}
	}
	return Text(t.vol, style)
}

func (g *gen) measure(consolidate bool) Order {
	drawers := []Drawer{DrawerDry, DrawerDry, DrawerLiquid, DrawerButter}
	d := drawers[g.rng.IntN(len(drawers))]
	list := measureTargets[d]
	// Skip targets that are already one kind of scoop ("3 T"): no thinking needed.
	var t target
	var amount string
	var o Order
	for range 20 {
		t = list[g.fresh("measure"+string(d), len(list))]
		amount = g.targetText(t, consolidate)
		o = buildOrder(KMeasure, t.vol, d)
		if !trivial(amount, o.Answer.Text) {
			break
		}
	}
	o.Prompt = fmt.Sprintf("Measure %s %s.", amount, ing[t.ing].Name)
	o.Context = "Tap tools into the bowl. Fewest scoops earns a bonus."
	o.Hints = measureHints(t.vol, d, amount)
	o.Explain = append([]string{fmt.Sprintf("%s = %s.", amount, o.Answer.Text)}, measureFacts(o.Answer.Text)...)
	switch d {
	case DrawerDry:
		return withVisual(o, wallVisual(t.vol, t.vol))
	case DrawerButter:
		return withVisual(o, butterVisual(t.vol, t.vol))
	}
	return o
}

func toolIDs(ts []Tool) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.ID
	}
	return out
}

// trivial reports whether the answer just repeats the amount ("3 Tbsp" →
// "3 T"), which needs no thinking.
func trivial(amount, answer string) bool {
	norm := strings.NewReplacer("Tbsp", "T", "tsp", "t", "cups", "c", "cup", "c").Replace(amount)
	return norm == answer
}

// buildOrder sets up a tool-building order for an exact amount.
func buildOrder(k Kind, vol int, d Drawer) Order {
	best, ok := Solve(vol, d)
	if !ok {
		panic(fmt.Sprintf("kitchen: %d can't be measured with the %s drawer", vol, d))
	}
	return Order{Kind: k, Input: InBuild, Drawer: d, Tools: d.Tools(),
		Answer: Answer{Vol: vol, Text: Combo(best), BestScoops: len(best), Best: toolIDs(best)}}
}

func measureHints(vol int, d Drawer, amount string) []string {
	best, _ := Solve(vol, d)
	first := best[0]
	h := []string{fmt.Sprintf("What's the biggest tool that fits in %s without going over?", amount)}
	if d == DrawerLiquid {
		h[0] = fmt.Sprintf("Which line on the liquid cup is the highest one that doesn't go over %s?", amount)
	}
	if len(best) > 1 {
		h = append(h, fmt.Sprintf("Start with %s. How much is still missing, and which tool holds that?", first.Label))
	} else {
		h = append(h, "Could one single tool hold all of it? Remember 4 T = ¼ c and 3 t = 1 T.")
	}
	return h
}

// measureFacts adds the equivalents behind a combo, as a reminder.
func measureFacts(combo string) []string {
	var out []string
	if strings.Contains(combo, "c") && strings.Contains(combo, "T") {
		out = append(out, "Remember: 4 T = ¼ c, and 16 T = 1 c.")
	}
	if strings.Contains(combo, "stick") {
		out = append(out, "1 stick = ½ c = 8 T.")
	}
	return out
}

// ---------- Conversion Factor ----------

func (g *gen) pickRecipeFactor(factors []frac.Frac, ok func(Recipe, frac.Frac) bool) (Recipe, frac.Frac) {
	for range 500 {
		r := Recipes[g.rng.IntN(len(Recipes))]
		f := factors[g.rng.IntN(len(factors))]
		want := frac.Int(int64(r.Makes)).Mul(f)
		if want.IsInt() && (ok == nil || ok(r, f)) {
			return r, f
		}
	}
	panic("kitchen: no recipe fits these factors")
}

func factorHints(makes, want int) []string {
	return []string{
		"Is your batch bigger or smaller than the recipe? So is the factor more than 1 or less than 1?",
		fmt.Sprintf("The factor answers this: %d × ? = %d.", makes, want),
		fmt.Sprintf("Divide what you want by what the recipe makes. What's %d ÷ %d?", want, makes),
	}
}

func (g *gen) factor(factors []frac.Frac) Order {
	r, f := g.pickRecipeFactor(factors, nil)
	want := int(frac.Int(int64(r.Makes)).Mul(f).Num())
	return Order{Kind: KFactor, Input: InNumber, Unit: "×",
		Prompt:  fmt.Sprintf("%s makes %d %s. You need %d. What's the conversion factor?", r.Name, r.Makes, r.MakesWhat, want),
		Context: "Type it as a whole number, fraction, or decimal: 2, 1/2, 1 1/2, or 1.5",
		Hints:   factorHints(r.Makes, want),
		Explain: []string{fmt.Sprintf("Want ÷ makes: %d ÷ %d = %s. Multiply every line by %s.", want, r.Makes, Mixed(f), Mixed(f))},
		Answer:  Answer{Num: pair(f), Text: "×" + Mixed(f), Makes: r.Makes, Want: want},
	}
}

// ---------- Scaling lines ----------

// scaledLine works out a line ×f, if it can be measured with real tools.
func scaledLine(l Line, f frac.Frac) (Line, bool) {
	out := Line{Ing: l.Ing}
	switch ing[l.Ing].Kind {
	case Weight:
		out.Oz = l.Oz.Mul(f)
		return out, slices.Contains([]int64{1, 2, 4}, out.Oz.Den())
	case Count:
		out.N = l.N.Mul(f)
		return out, slices.Contains([]int64{1, 2}, out.N.Den())
	}
	vol := frac.Int(int64(l.Vol)).Mul(f)
	if !vol.IsInt() {
		return out, false
	}
	out.Vol = int(vol.Num())
	if out.Vol > 8*Cup {
		return out, false
	}
	_, ok := Solve(out.Vol, drawerFor(l.Ing))
	return out, ok
}

// lineText writes a line as the recipe card prints it, in a recipe style:
// "¾ c sugar", "¾ C. sugar", "3/4 cup sugar".
func lineText(l Line, s Style) string {
	in := ing[l.Ing]
	name := in.Name
	if s != Short && in.Fancy != "" && in.Kind != Count {
		name = in.Fancy
	}
	switch in.Kind {
	case Weight:
		return OzStyled(l.Oz, in.Pounds, s) + " " + name
	case Count:
		size := ""
		if s != Short && in.Fancy != "" {
			size = in.Fancy + " "
		}
		if l.N.Equal(frac.Int(1)) {
			return "1 " + size + in.Name
		}
		return spaced(l.N, s) + " " + size + in.Plural
	case Fat:
		if l.Vol%Stick == 0 {
			n := l.Vol / Stick
			return fmt.Sprintf("%s (%d stick%s) %s", Text(l.Vol, s), n, map[bool]string{true: "s"}[n > 1], name)
		}
	}
	return Text(l.Vol, s) + " " + name
}

// amountText is just the amount of a line, in a style: "2 T", "3 eggs".
func amountText(l Line, s Style) string {
	in := ing[l.Ing]
	switch in.Kind {
	case Weight:
		return OzStyled(l.Oz, in.Pounds, s)
	case Count:
		if l.N.Cmp(frac.Int(1)) <= 0 {
			return spaced(l.N, s) + " " + in.Name
		}
		return spaced(l.N, s) + " " + in.Plural
	}
	return Text(l.Vol, s)
}

// measuredText is how a scaled line gets measured: tools, weight, or count.
func measuredText(l Line) string {
	in := ing[l.Ing]
	switch in.Kind {
	case Weight:
		return Oz(l.Oz, in.Pounds)
	case Count:
		return Mixed(l.N) + " " + map[bool]string{true: in.Plural, false: in.Name}[!l.N.Equal(frac.Int(1))]
	}
	best, _ := Solve(l.Vol, drawerFor(l.Ing))
	return Combo(best)
}

// smallUnit writes a volume in T or t (whichever is whole) for working.
func smallUnit(e int) string {
	if e%Tbsp == 0 {
		return fmt.Sprintf("%d T", e/Tbsp)
	}
	return Mixed(frac.New(int64(e), Tsp)) + " t"
}

// lineOrder builds the order that measures one scaled line.
func lineOrder(k Kind, orig, scaled Line, f frac.Frac, fs string, st Style) Order {
	in := ing[orig.Ing]
	switch in.Kind {
	case Weight:
		unit, amount, given := "oz", scaled.Oz, orig.Oz
		if in.Pounds {
			unit, amount, given = "lb", scaled.Oz.Div(frac.Int(16)), orig.Oz.Div(frac.Int(16))
		}
		return Order{Kind: k, Input: InNumber, Unit: unit,
			Prompt: fmt.Sprintf("Scale “%s” by ×%s. How much goes on the scale?", lineText(orig, st), fs),
			Hints: []string{
				fmt.Sprintf("With ×%s, should you weigh out more or less than the recipe says?", fs),
				fmt.Sprintf("What's %s × %s?", Mixed(given), fs),
			},
			Explain: []string{fmt.Sprintf("%s %s × %s = %s %s. Weigh it, or read the package labels.", Mixed(given), unit, fs, Mixed(amount), unit)},
			Answer:  Answer{Num: pair(amount), Text: Mixed(amount) + " " + unit, Given: pair(given), Factor: pair(f)},
		}
	case Count:
		return Order{Kind: k, Input: InNumber, Unit: in.Plural,
			Prompt:  fmt.Sprintf("Scale “%s” by ×%s. How many?", lineText(orig, st), fs),
			Hints:   []string{fmt.Sprintf("With ×%s, more or fewer than %s?", fs, Mixed(orig.N)), fmt.Sprintf("What's %s × %s?", Mixed(orig.N), fs)},
			Explain: []string{fmt.Sprintf("%s × %s = %s.", Mixed(orig.N), fs, Mixed(scaled.N))},
			Answer:  Answer{Num: pair(scaled.N), Text: measuredText(scaled), Given: pair(orig.N), Factor: pair(f)},
		}
	}
	d := drawerFor(orig.Ing)
	o := buildOrder(k, scaled.Vol, d)
	o.Prompt = fmt.Sprintf("Scale “%s” by ×%s, then measure it.", lineText(orig, st), fs)
	o.Answer.Given, o.Answer.Factor = [2]int64{int64(orig.Vol), 1}, pair(f)
	// Socratic ladder: direction, convert to a friendlier unit, multiply.
	o.Hints = []string{fmt.Sprintf("With ×%s, should the amount get bigger or smaller?", fs)}
	small := smallUnit(orig.Vol)
	if orig.Vol >= QuarterCup && orig.Vol%Tbsp == 0 && f.Den() != 1 {
		o.Hints = append(o.Hints, fmt.Sprintf("Fractions of cups get messy. How many T is %s? (There are 16 T in every cup.)", Text(orig.Vol, Short)))
		o.Hints = append(o.Hints, fmt.Sprintf("What's %s × %s? Which tools add up to that?", small, fs))
	} else {
		o.Hints = append(o.Hints, fmt.Sprintf("What's %s × %s? Then which tools add up to it?", Text(orig.Vol, Short), fs))
	}
	// Work in one unit (T or t), then say what that is in cups if it helps.
	unit, size := "t", Tsp
	if orig.Vol%Tbsp == 0 {
		unit, size = "T", Tbsp
	}
	product := Mixed(frac.New(int64(scaled.Vol), int64(size))) + " " + unit
	work := fmt.Sprintf("%s × %s = %s", small, fs, product)
	if other := Text(scaled.Vol, Short); other != product {
		work += " = " + other
	}
	o.Explain = []string{work + ".", fmt.Sprintf("Measure it as %s.", o.Answer.Text)}
	if d == DrawerButter {
		return withVisual(o, butterVisual(orig.Vol, scaled.Vol))
	}
	return withVisual(o, wallVisual(orig.Vol, scaled.Vol))
}

func (g *gen) scale(factors []frac.Frac) Order {
	for range 500 {
		r, f := g.pickRecipeFactor(factors, nil)
		l := r.Lines[g.rng.IntN(len(r.Lines))]
		if ing[l.Ing].Kind == Count {
			continue
		}
		s, ok := scaledLine(l, f)
		key := fmt.Sprint("scale", r.Name, l.Ing)
		if !ok || g.seen(key) {
			continue
		}
		o := lineOrder(KScale, l, s, f, Mixed(f), Styles[g.rng.IntN(len(Styles))])
		want := int(frac.Int(int64(r.Makes)).Mul(f).Num())
		o.Context = fmt.Sprintf("%s: makes %d %s, you need %d (×%s).", r.Name, r.Makes, r.MakesWhat, want, Mixed(f))
		return o
	}
	panic("kitchen: no scalable line found")
}

// ---------- Recipe Card ----------

// cardFits reports whether every line of a recipe scales to something a
// cook can measure (eggs may come out in halves; that's a Chef's Call).
func cardFits(r Recipe, f frac.Frac) bool {
	for _, l := range r.Lines {
		if _, ok := scaledLine(l, f); !ok {
			return false
		}
	}
	return true
}

func (g *gen) card(factors []frac.Frac) []Order {
	r, f := g.pickRecipeFactor(factors, cardFits)
	g.cards++
	fs := Mixed(f)
	want := int(frac.Int(int64(r.Makes)).Mul(f).Num())
	card := &Card{ID: g.cards, Name: r.Name, Makes: r.Makes, MakesWhat: r.MakesWhat, Want: want, Factor: fs}
	st := Styles[g.rng.IntN(len(Styles))] // each card is written in one recipe style
	for _, l := range r.Lines {
		s, _ := scaledLine(l, f)
		card.Lines = append(card.Lines, CardLine{Text: lineText(l, st), Scaled: measuredText(s), Tool: toolLabel(l)})
	}

	fo := g.factorFor(r, f, want)
	fo.Card = card
	out := []Order{fo}
	for i, l := range r.Lines {
		s, _ := scaledLine(l, f)
		var o Order
		if ing[l.Ing].Kind == Count && !s.N.IsInt() {
			o = eggCall(g.rng, l, s, fs, st)
		} else {
			o = lineOrder(KCard, l, s, f, fs, st)
		}
		o.Card, o.Line = card, i
		out = append(out, o)
	}
	return out
}

func (g *gen) factorFor(r Recipe, f frac.Frac, want int) Order {
	return Order{Kind: KFactor, Input: InNumber, Unit: "×",
		Prompt:  fmt.Sprintf("Mise en place: %s makes %d %s, and the order is for %d. What's your conversion factor?", r.Name, r.Makes, r.MakesWhat, want),
		Context: "Type 2, 1/2, 1 1/2, or 1.5",
		Hints:   factorHints(r.Makes, want),
		Explain: []string{fmt.Sprintf("%d ÷ %d = %s. Every line gets ×%s.", want, r.Makes, Mixed(f), Mixed(f))},
		Answer:  Answer{Num: pair(f), Text: "×" + Mixed(f), Makes: r.Makes, Want: want},
	}
}

// eggCall turns half an egg into a judgment question.
func eggCall(rng *rand.Rand, orig, scaled Line, fs string, st Style) Order {
	x := scaled.N
	lo := x.Num() / x.Den()
	hi := lo + 1
	eggs := func(k int64) string {
		if k == 1 {
			return "1 egg"
		}
		return fmt.Sprintf("%d eggs", k)
	}
	right := []string{fmt.Sprintf("Round up to %s", eggs(hi))}
	if lo > 0 {
		right = append(right, fmt.Sprintf("Use %s, then beat 1 more and add half of it (about 1½ T)", eggs(lo)))
	} else {
		right = append(right, "Beat 1 egg and add half of it (about 1½ T)")
	}
	wrong := []string{fmt.Sprintf("Crack %s eggs straight into the bowl", Mixed(x)), "Leave the eggs out"}
	if !orig.N.Equal(frac.Int(hi)) {
		wrong = append(wrong, fmt.Sprintf("Use %s, same as the recipe", eggs(orig.N.Num())))
	}
	cs, idx := choices(rng, right, wrong)
	return Order{Kind: KCall, Input: InChoice, Choices: cs,
		Prompt:  fmt.Sprintf("Chef's Call: “%s” ×%s = %s eggs. What do you do?", lineText(orig, st), fs, Mixed(x)),
		Hints:   []string{"Can you crack half an egg cleanly?", "What could you do to an egg so you can measure part of it?"},
		Explain: []string{fmt.Sprintf("%s eggs: round up, or beat one egg and use half (a large egg is about 3 T).", Mixed(x))},
		Answer:  Answer{Choices: idx, Text: strings.Join(right, " or ")},
	}
}

// ---------- Fix the Mistake ----------

// fix plants one mistake in a scaled recipe. The card shows each line as
// original → what the cook wrote, side by side, so students compare rows
// instead of hunting around the screen.
func (g *gen) fix() Order {
	for range 500 {
		r, f := g.pickRecipeFactor(cardFactors, cardFits)
		fs := Mixed(f)
		st := Styles[g.rng.IntN(len(Styles))]
		var scaled []Line
		var cook []string
		for _, l := range r.Lines {
			sl, _ := scaledLine(l, f)
			scaled = append(scaled, sl)
			cook = append(cook, amountText(sl, st))
		}
		bad := g.rng.IntN(len(r.Lines))
		l, s := r.Lines[bad], scaled[bad]
		in := ing[l.Ing]
		right := cook[bad]
		var why, fixHint string
		switch g.rng.IntN(3) {
		case 0: // T/t mix-up on a small spoon amount
			if in.Kind == Count || in.Kind == Weight || s.Vol >= Tbsp || s.Vol%Tsp != 0 {
				continue
			}
			k := s.Vol / Tsp
			cook[bad] = fmt.Sprintf("%d %s", k, unitName("T", st, k > 1))
			why = fmt.Sprintf("The cook wrote %s, but %s × %s = %s. A tablespoon is 3 times bigger than a teaspoon!", cook[bad], amountText(l, st), fs, right)
			fixHint = "Is every tablespoon (T) really a tablespoon, or should one of them be a teaspoon (t)?"
		case 1: // forgot to scale
			if in.Kind == Count || amountText(l, st) == right {
				continue
			}
			cook[bad] = amountText(l, st)
			why = fmt.Sprintf("The cook forgot to scale the %s: %s × %s = %s.", in.Name, cook[bad], fs, right)
			fixHint = "Did every line actually change? Does one still match the original?"
		default: // went the wrong way: divided instead of multiplied (or the reverse)
			wrong, ok := scaledLine(l, frac.Int(1).Div(f))
			if !ok || in.Kind == Count {
				continue
			}
			cook[bad] = amountText(wrong, st)
			dir := "bigger"
			if f.Cmp(frac.Int(1)) < 0 {
				dir = "smaller"
			}
			why = fmt.Sprintf("The cook went the wrong way. With ×%s every amount gets %s: %s × %s = %s, not %s.", fs, dir, amountText(l, st), fs, right, cook[bad])
			fixHint = fmt.Sprintf("With ×%s, should every amount get bigger or smaller? Does one row go the other way?", fs)
		}
		want := int(frac.Int(int64(r.Makes)).Mul(f).Num())
		card := &Card{Name: r.Name, Makes: r.Makes, MakesWhat: r.MakesWhat, Want: want, Factor: fs}
		var cs []string
		for i, l := range r.Lines {
			name := ing[l.Ing].Name
			card.Lines = append(card.Lines, CardLine{Text: amountText(l, st), Scaled: cook[i], Tool: name})
			cs = append(cs, fmt.Sprintf("%s: %s → %s", name, amountText(l, st), cook[i]))
		}
		return Order{Kind: KFix, Input: InChoice, Choices: cs,
			Prompt:  fmt.Sprintf("The line cook scaled %s ×%s (%d → %d %s). One row is wrong. Which one?", r.Name, fs, r.Makes, want, r.MakesWhat),
			Context: fmt.Sprintf("Each row shows the original amount and what the cook wrote after ×%s. Tap the wrong row.", fs),
			Card:    card, Line: -1,
			Hints:   []string{fmt.Sprintf("Go row by row. Is each new amount really the original × %s?", fs), fixHint},
			Explain: []string{strings.ReplaceAll(why, "..", ".")}, // "1 C." at a sentence end
			Answer:  Answer{Choices: []int{bad}, Text: cs[bad]},
		}
	}
	panic("kitchen: couldn't plant a mistake")
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func newRNG(seed uint64) *rand.Rand { return rand.New(rand.NewPCG(seed, 0x6b69_7463_6865_6e)) }
