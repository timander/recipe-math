package kitchen

import (
	"slices"
	"strings"
	"testing"

	"github.com/timander/recipe-math/frac"
)

// Focused tests for branches that the broad property tests don't pin down.
// Several were found by mutation testing (make mutation).

func TestCheckChoiceBounds(t *testing.T) {
	o := Order{Kind: KVocab, Input: InChoice, Choices: []string{"a", "b"}, Hints: []string{"Hint one?"}, Answer: Answer{Choices: []int{1}}}
	for _, c := range []int{-1, 2} {
		if r := Check(o, Response{Choice: c}); !r.Invalid {
			t.Errorf("choice %d should be invalid", c)
		}
	}
	if r := Check(o, Response{Choice: 1}); !r.Correct {
		t.Error("choice 1 is right")
	}
	r := Check(o, Response{Choice: 0})
	if r.Correct || r.Invalid || r.Message != "Not quite. Hint one?" || !r.FreeHint {
		t.Errorf("wrong choice: %+v", r)
	}
	o.Hints = nil
	if r := Check(o, Response{Choice: 0}); r.Message != "Not quite." || r.FreeHint {
		t.Errorf("wrong choice without hints: %+v", r)
	}
}

func TestCheckNumberHintAndFreeHint(t *testing.T) {
	o := Order{Kind: KFact, Input: InNumber, Hints: []string{"First?"}, Answer: Answer{Num: [2]int64{4, 1}}}
	if r := Check(o, Response{Text: "5"}); r.Message != "Not quite. First?" || !r.FreeHint {
		t.Errorf("fact wrong: %+v", r)
	}
	o.Hints = nil
	if r := Check(o, Response{Text: "5"}); r.Message != "Not quite." || r.FreeHint {
		t.Errorf("fact wrong without hints: %+v", r)
	}
	if r := Check(o, Response{Text: "abc"}); !r.Invalid {
		t.Error("unreadable should be invalid")
	}
}

func TestCheckBuildMoreOrLess(t *testing.T) {
	o := buildOrder(KMeasure, tb(5), DrawerDry) // ¼ c + 1 T
	if r := Check(o, Response{Tools: []string{"c14", "T1", "T1"}}); !strings.Contains(r.Message, "Which scoop could come out?") || !strings.Contains(r.Message, "6 T") {
		t.Errorf("too much: %q", r.Message)
	}
	if r := Check(o, Response{Tools: []string{"c14"}}); !strings.Contains(r.Message, "What's still missing?") || !strings.Contains(r.Message, "¼ c") {
		t.Errorf("too little: %q", r.Message)
	}
	if r := Check(o, Response{Tools: []string{"L96", "L192"}}); !r.Invalid {
		t.Error("two liquid lines should be invalid")
	}
	if r := Check(o, Response{Tools: []string{"c14", "T1"}}); !r.Correct || r.Scoops != 2 || r.BestScoops != 2 {
		t.Errorf("best answer: %+v", r)
	}
}

func TestCheckWeightScaling(t *testing.T) {
	cheese := Line{Ing: "cheese", Oz: frac.Int(8)}
	s, _ := scaledLine(cheese, frac.Int(2))
	o := lineOrder(KScale, cheese, s, frac.Int(2), "2", Short)
	cases := map[string]string{"16": "Correct!", "8": "original amount", "4": "wrong way", "12": "Check your multiplication: 8 × 2 = ?"}
	for in, want := range cases {
		if r := Check(o, Response{Text: in}); !strings.Contains(r.Message, want) {
			t.Errorf("%s: %q, want %q", in, r.Message, want)
		}
	}
	if o.Unit != "oz" || o.Prompt != "Scale “8 oz shredded cheddar” by ×2. How much goes on the scale?" {
		t.Errorf("weight order: %q %q", o.Unit, o.Prompt)
	}
	beef := Line{Ing: "beef", Oz: frac.Int(16)}
	s, _ = scaledLine(beef, frac.New(1, 2))
	o = lineOrder(KCard, beef, s, frac.New(1, 2), "½", Short)
	if o.Unit != "lb" || o.Answer.Text != "½ lb" {
		t.Errorf("pounds: %q %q", o.Unit, o.Answer.Text)
	}
	egg := Line{Ing: "egg", N: frac.Int(2)}
	s, _ = scaledLine(egg, frac.Int(3))
	o = lineOrder(KCard, egg, s, frac.Int(3), "3", Short)
	if o.Unit != "eggs" || o.Answer.Text != "6 eggs" || o.Prompt != "Scale “2 eggs” by ×3. How many?" {
		t.Errorf("eggs: %q %q %q", o.Unit, o.Answer.Text, o.Prompt)
	}
}

func TestFactorFlipNeedsWant(t *testing.T) {
	o := Order{Kind: KFactor, Input: InNumber, Answer: Answer{Num: [2]int64{2, 1}, Makes: 12, Want: 24}}
	if r := Check(o, Response{Text: "1/2"}); !strings.Contains(r.Message, "You flipped it") {
		t.Errorf("flip: %q", r.Message)
	}
	if r := Check(o, Response{Text: "3"}); r.Message != "Check it: what's 12 × 3? Is that 24?" {
		t.Errorf("other: %q", r.Message)
	}
}

func TestToolWhyNot(t *testing.T) {
	cases := []struct {
		right, chose int
		want         string
	}{
		{3, 1, "fl oz vs. oz trap"},
		{0, 1, "level off a dry ingredient"},
		{1, 0, "fill a dry cup to the rim with liquid"},
		{2, 0, "smallest cup"},
		{0, 2, "that many spoonfuls"},
		{4, 0, "lines on the wrapper"},
		{0, 3, "dry, liquid, butter, or sold by weight"},
	}
	for _, c := range cases {
		if got := toolWhyNot(c.right, c.chose); !strings.Contains(got, c.want) {
			t.Errorf("toolWhyNot(%d,%d) = %q, want %q", c.right, c.chose, got, c.want)
		}
	}
}

func TestToolFor(t *testing.T) {
	cases := map[int]Line{
		0: v("flour", c(1, 1)), 1: v("milk", c(1, 1)), 2: v("milk", tb(2)), 3: w("cheese", 8), 4: v("butter", tb(2)),
	}
	for want, l := range cases {
		if got := toolFor(l); got != want {
			t.Errorf("toolFor(%s) = %d, want %d", l.Ing, got, want)
		}
	}
	if toolLabel(n("egg", 2)) != "count" || toolLabel(v("flour", c(1, 1))) != "dry cups" {
		t.Error("toolLabel")
	}
}

func TestLiquidLines(t *testing.T) {
	if len(liquidLines) != 24 {
		t.Fatalf("a 4-cup liquid measure has 24 lines (¼s and ⅓s), got %d", len(liquidLines))
	}
	if liquidLines[0].Size != QuarterCup || liquidLines[len(liquidLines)-1].Size != 4*Cup {
		t.Errorf("lines should run from ¼ c to 4 c: %v … %v", liquidLines[0], liquidLines[len(liquidLines)-1])
	}
	for i := 1; i < len(liquidLines); i++ {
		if liquidLines[i].Size <= liquidLines[i-1].Size {
			t.Errorf("lines out of order or repeated at %d", i)
		}
	}
	best, ok := Solve(c(1, 1)+tb(1), DrawerLiquid)
	if !ok || Combo(best) != "1 c (liquid cup) + 1 T" {
		t.Errorf("1 c + 1 T liquid = %q", Combo(best))
	}
	best, _ = Solve(tb(2), DrawerLiquid)
	if Combo(best) != "2 T" {
		t.Errorf("small liquid amounts use spoons only: %q", Combo(best))
	}
}

func TestComboStyled(t *testing.T) {
	cases := []struct {
		e    int
		s    Style
		want string
	}{
		{tb(2) + ts(2, 1), Lab, "2 Tbsp. + 2 tsp."},
		{tb(1) + ts(1, 1), Words, "1 tablespoon + 1 teaspoon"},
		{tb(2) + ts(3, 2), Words, "2 tablespoons + 1 teaspoon + ½ teaspoon"},
		{tb(2) + ts(2, 1), Short, "2 T + 2 t"},
		{tb(1) + ts(1, 2), Abbr, "1 Tbsp + ½ tsp"},
	}
	for _, c := range cases {
		best, _ := Solve(c.e, DrawerDry)
		if got := comboStyled(best, c.s); got != c.want {
			t.Errorf("comboStyled(%d, %d) = %q, want %q", c.e, c.s, got, c.want)
		}
	}
	best, _ := Solve(c(1, 1)+c(1, 4), DrawerDry)
	if got := comboStyled(best, Lab); got != "1 C. + ¼ C." {
		t.Errorf("cups in lab style = %q", got)
	}
}

func TestTextEdges(t *testing.T) {
	cases := []struct {
		e    int
		s    Style
		want string
	}{
		{c(1, 1), Words, "1 cup"},
		{c(3, 2), Words, "1 ½ cups"},
		{tb(1), Words, "1 tablespoon"},
		{tb(2), Words, "2 tablespoons"},
		{ts(1, 1), Words, "1 teaspoon"},
		{ts(3, 2), Words, "1 ½ teaspoons"},
		{ts(6, 1), Short, "2 T"},
		{ts(7, 1), Short, "2 T + 1 t"},
		{ts(5, 1), Short, "5 t"},
		{tb(4), Short, "¼ c"},
		{tb(3), Short, "3 T"},
		{ts(1, 4), Lab, "¼ tsp."},
	}
	for _, c := range cases {
		if got := Text(c.e, c.s); got != c.want {
			t.Errorf("Text(%d, %d) = %q, want %q", c.e, c.s, got, c.want)
		}
	}
	if got := Mixed(frac.New(7, 5)); got != "1 2/5" {
		t.Errorf("Mixed(7/5) = %q", got)
	}
	if got := Mixed(frac.New(2, 5)); got != "2/5" {
		t.Errorf("Mixed(2/5) = %q", got)
	}
	if got := OzStyled(frac.Int(32), true, Words); got != "2 pounds" {
		t.Errorf("OzStyled = %q", got)
	}
	if got := OzStyled(frac.Int(16), true, Words); got != "1 pound" {
		t.Errorf("OzStyled one pound = %q", got)
	}
	if got := OzStyled(frac.Int(8), false, Lab); got != "8 oz." {
		t.Errorf("OzStyled lab = %q", got)
	}
}

func TestSwapLabels(t *testing.T) {
	cases := []struct {
		unit   string
		s      Style
		amount frac.Frac
		want   string
	}{
		{"c", Abbr, frac.Int(2), "cups"},
		{"c", Abbr, frac.Int(1), "cup"},
		{"c", Short, frac.Int(2), "c"},
		{"pt", Abbr, frac.Int(2), "pints"},
		{"qt", Abbr, frac.Int(3), "quarts"},
		{"gal", Abbr, frac.Int(2), "gallons"},
		{"stick", Lab, frac.Int(2), "sticks"},
		{"T", Words, frac.Int(2), "tablespoons"},
		{"T", Lab, frac.Int(2), "Tbsp."},
		{"oz", Words, frac.New(1, 2), "ounce"},
	}
	for _, c := range cases {
		if got := swapUnits[c.unit].label(c.s, c.amount); got != c.want {
			t.Errorf("label(%s, %d, %v) = %q, want %q", c.unit, c.s, c.amount, got, c.want)
		}
	}
	if swapVolume("c", frac.New(1, 2)) != HalfCup || swapVolume("stick", frac.Int(1)) != Stick || swapVolume("lb", frac.Int(1)) != 0 {
		t.Error("swapVolume")
	}
	if swapVolume("T", frac.New(1, 48)) != 0 {
		t.Error("swapVolume should reject amounts that aren't whole eighth-teaspoons")
	}
}

func TestVisualLimits(t *testing.T) {
	cases := []struct {
		got  *Visual
		want *Visual
	}{
		{wallVisual(Cup, Cup), &Visual{"wall", Cup, Cup}},
		{wallVisual(Cup+Tbsp, 0), nil},
		{wallVisual(0, 0), nil},
		{wallVisual(Tbsp, Cup+Tbsp), &Visual{"wall", Tbsp, 0}},
		{wallVisual(Tbsp, Tsp), &Visual{"wall", Tbsp, 0}},
		{butterVisual(4*Stick, 4*Stick), &Visual{"butter", 4 * Stick, 4 * Stick}},
		{butterVisual(4*Stick+Tbsp, 0), nil},
		{butterVisual(Tsp, 0), nil},
		{butterVisual(Stick, 5*Stick), &Visual{"butter", Stick, 0}},
		{butterVisual(Stick, Tsp), &Visual{"butter", Stick, 0}},
	}
	for i, c := range cases {
		if (c.got == nil) != (c.want == nil) || (c.got != nil && *c.got != *c.want) {
			t.Errorf("case %d: got %+v, want %+v", i, c.got, c.want)
		}
	}
	o := withVisual(Order{Hints: []string{"a?"}}, nil)
	if o.Visual != nil || len(o.Hints) != 1 {
		t.Error("nil visual should leave the order alone")
	}
}

func TestEggCall(t *testing.T) {
	rng := newRNG(1)
	o := eggCall(rng, n("egg", 3), Line{Ing: "egg", N: frac.New(3, 2)}, "½", Short)
	for _, want := range []string{"Round up to 2 eggs", "Use 1 egg, then beat 1 more and add half of it (about 1½ T)"} {
		if !slices.Contains(o.Choices, want) {
			t.Errorf("missing right choice %q in %q", want, o.Choices)
		}
	}
	if !slices.Contains(o.Choices, "Use 3 eggs, same as the recipe") || !slices.Contains(o.Choices, "Crack 1½ eggs straight into the bowl") {
		t.Errorf("wrong choices: %q", o.Choices)
	}
	o = eggCall(rng, n("egg", 1), Line{Ing: "egg", N: frac.New(1, 2)}, "½", Short)
	if !slices.Contains(o.Choices, "Beat 1 egg and add half of it (about 1½ T)") || !slices.Contains(o.Choices, "Round up to 1 egg") {
		t.Errorf("half an egg: %q", o.Choices)
	}
	if slices.Contains(o.Choices, "Use 1 egg, same as the recipe") {
		t.Error("don't offer the recipe amount as wrong when it's also the rounded answer")
	}
}

func TestScaledLineLimits(t *testing.T) {
	if _, ok := scaledLine(v("flour", c(3, 1)), frac.Int(4)); ok {
		t.Error("12 c of flour is more than a recipe card should ask for")
	}
	if _, ok := scaledLine(v("flour", c(2, 1)), frac.Int(4)); !ok {
		t.Error("8 c is the limit and should be allowed")
	}
	if _, ok := scaledLine(n("egg", 1), frac.New(1, 3)); ok {
		t.Error("a third of an egg isn't a card amount")
	}
	if got := amountText(Line{Ing: "egg", N: frac.New(1, 2)}, Short); got != "½ egg" {
		t.Errorf("half egg = %q", got)
	}
	if got := amountText(Line{Ing: "egg", N: frac.Int(1)}, Short); got != "1 egg" {
		t.Errorf("one egg = %q", got)
	}
	if got := lineText(Line{Ing: "egg", N: frac.Int(1)}, Lab); got != "1 large egg" {
		t.Errorf("one large egg = %q", got)
	}
	if got := lineText(Line{Ing: "butter", Vol: tb(6)}, Short); got != "6 T butter" {
		t.Errorf("butter without whole sticks = %q", got)
	}
	if got := lineText(Line{Ing: "butter", Vol: c(1, 1)}, Short); got != "1 c (2 sticks) butter" {
		t.Errorf("butter sticks = %q", got)
	}
	if smallUnit(tb(2)) != "2 T" || smallUnit(ts(3, 2)) != "1½ t" {
		t.Error("smallUnit")
	}
}

func TestShiftIDsAndLines(t *testing.T) {
	orders, _ := Shift(5, 9)
	for i, o := range orders {
		if o.ID != i+1 {
			t.Errorf("order %d has ID %d", i, o.ID)
		}
		if o.Card == nil && o.Line != -1 {
			t.Errorf("order %d without a card should have Line -1", o.ID)
		}
	}
	if capitalize("") != "" || capitalize("abc") != "Abc" {
		t.Error("capitalize")
	}
}

// Found by /code-review: no order may need more scoops than the bowl holds,
// and a recipe card's factor order isn't tied to an ingredient line.
func TestEveryAnswerFitsTheBowl(t *testing.T) {
	for _, r := range Ranks {
		for seed := range uint64(400) {
			orders, _ := Shift(r.ID, seed)
			for _, o := range orders {
				if o.Input == InBuild && o.Answer.BestScoops > MaxScoops {
					t.Fatalf("rank %d seed %d: %q needs %d scoops", r.ID, seed, o.Prompt, o.Answer.BestScoops)
				}
				if o.Kind == KFactor && o.Line != -1 {
					t.Fatalf("rank %d seed %d: factor order has Line %d", r.ID, seed, o.Line)
				}
			}
		}
	}
	if _, ok := scaledLine(v("milk", c(2, 1)), frac.Int(4)); ok {
		t.Error("8 c of milk can't be measured in a reasonable number of scoops")
	}
}

func TestSwapHintPlurals(t *testing.T) {
	g := &gen{rng: newRNG(5)}
	for range 300 {
		o := g.swap()
		for _, h := range o.Hints {
			if strings.Contains(h, "stick of butters") || strings.Contains(h, "ouncess") {
				t.Errorf("bad plural: %q", h)
			}
		}
	}
}

func TestReference(t *testing.T) {
	r := Ref()
	if len(r.Facts) != len(facts) || len(r.Abbrevs) < 10 || len(r.Tools) != 5 || len(r.Tips) < 5 {
		t.Errorf("reference is missing content: %d facts, %d abbrevs, %d tools, %d tips", len(r.Facts), len(r.Abbrevs), len(r.Tools), len(r.Tips))
	}
	if len(r.Drawer) != len(dryCups)+len(spoons)+len(butterTools) {
		t.Errorf("drawer has %d tools", len(r.Drawer))
	}
	if r.Facts[0] != "3 t = 1 T" {
		t.Errorf("first fact = %q", r.Facts[0])
	}
}

func TestBuildOrderPanicsOnUnmeasurable(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("an amount the drawer can't make should panic (it's a data bug)")
		}
	}()
	buildOrder(KMeasure, 5, DrawerButter)
}
