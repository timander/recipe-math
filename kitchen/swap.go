package kitchen

import (
	"fmt"

	"github.com/timander/recipe-math/frac"
)

// Unit Swap is the beginner station: one unit to another, nothing else.
// Only conversions cooks really make, with everyday amounts.

// swapUnit is a unit as each recipe style spells it.
type swapUnit struct {
	spell [4]string // Short, Abbr, Lab, Words (singular)
	word  string    // for hints: "tablespoon"
	words string    // plural, for hints: "tablespoons"
}

var swapUnits = map[string]swapUnit{
	"T":     {[4]string{"T", "Tbsp", "Tbsp.", "tablespoon"}, "tablespoon", "tablespoons"},
	"t":     {[4]string{"t", "tsp", "tsp.", "teaspoon"}, "teaspoon", "teaspoons"},
	"c":     {[4]string{"c", "cup", "C.", "cup"}, "cup", "cups"},
	"floz":  {[4]string{"fl oz", "fl oz", "fl. oz.", "fluid ounce"}, "fluid ounce", "fluid ounces"},
	"pt":    {[4]string{"pt", "pint", "pt.", "pint"}, "pint", "pints"},
	"qt":    {[4]string{"qt", "quart", "qt.", "quart"}, "quart", "quarts"},
	"gal":   {[4]string{"gal", "gallon", "gal.", "gallon"}, "gallon", "gallons"},
	"oz":    {[4]string{"oz", "oz", "oz.", "ounce"}, "ounce", "ounces"},
	"lb":    {[4]string{"lb", "lb", "lb.", "pound"}, "pound", "pounds"},
	"stick": {[4]string{"stick", "stick", "stick", "stick"}, "stick of butter", "sticks of butter"},
}

// label spells a unit for an amount, adding an s where a word needs it.
func (u swapUnit) label(s Style, amount frac.Frac) string {
	name := u.spell[s]
	plural := amount.Cmp(frac.Int(1)) > 0
	if plural && (s == Words || name == "stick" || name == "cup" || name == "pint" || name == "quart" || name == "gallon") {
		name += "s"
	}
	return name
}

type swapPair struct {
	big, small string
	factor     int64       // 1 big = factor small
	amounts    []frac.Frac // everyday amounts of the big unit
}

func fr(n, d int64) frac.Frac { return frac.New(n, d) }

var swapPairs = []swapPair{
	{"T", "t", 3, []frac.Frac{fr(1, 1), fr(2, 1), fr(3, 1), fr(4, 1)}},
	{"c", "T", 16, []frac.Frac{fr(1, 4), fr(1, 2), fr(3, 4), fr(1, 1)}},
	{"c", "floz", 8, []frac.Frac{fr(1, 2), fr(1, 1), fr(3, 2), fr(2, 1)}},
	{"pt", "c", 2, []frac.Frac{fr(1, 1), fr(2, 1), fr(3, 1)}},
	{"qt", "c", 4, []frac.Frac{fr(1, 2), fr(1, 1), fr(2, 1)}},
	{"qt", "pt", 2, []frac.Frac{fr(1, 1), fr(2, 1)}},
	{"gal", "qt", 4, []frac.Frac{fr(1, 2), fr(1, 1), fr(2, 1)}},
	{"lb", "oz", 16, []frac.Frac{fr(1, 4), fr(1, 2), fr(1, 1), fr(2, 1)}},
	{"stick", "T", 8, []frac.Frac{fr(1, 2), fr(1, 1), fr(2, 1)}},
	{"c", "stick", 2, []frac.Frac{fr(1, 2), fr(1, 1), fr(2, 1)}},
}

func (g *gen) swap() Order {
	p := swapPairs[g.fresh("swappair", len(swapPairs))]
	a := p.amounts[g.rng.IntN(len(p.amounts))]
	inSmall := a.Mul(frac.Int(p.factor))
	st := Styles[g.rng.IntN(len(Styles))]
	big, small := swapUnits[p.big], swapUnits[p.small]

	// Half the time go big → small (multiply), half small → big (divide).
	from, to, given, want := big, small, a, inSmall
	down := g.rng.IntN(2) == 0
	if down {
		from, to, given, want = small, big, inSmall, a
	}
	toLabel := to.label(st, want)
	hints := []string{
		fmt.Sprintf("Which is bigger: one %s or one %s? So should your number come out bigger or smaller than %s?", from.word, to.word, spaced(given, st)),
		fmt.Sprintf("How many %s fit in one %s?", small.words, big.word),
	}
	var explain string
	if down {
		hints = append(hints, fmt.Sprintf("How many groups of that size fit into %s?", spaced(given, st)))
		explain = fmt.Sprintf("%d %s = 1 %s, so %s ÷ %d = %s.", p.factor, small.label(st, frac.Int(p.factor)), big.spell[st], Mixed(given), p.factor, Mixed(want))
	} else {
		hints = append(hints, fmt.Sprintf("So you have %s groups of that many. How do you find the total?", spaced(given, st)))
		explain = fmt.Sprintf("1 %s = %d %s, so %s × %d = %s.", big.spell[st], p.factor, small.label(st, frac.Int(p.factor)), Mixed(given), p.factor, Mixed(want))
	}
	o := Order{Kind: KSwap, Input: InNumber, Unit: toLabel,
		Prompt:  fmt.Sprintf("%s %s = ___ %s", spaced(given, st), from.label(st, given), toLabel),
		Context: "Type a whole number or a fraction, like 4 or 1/2.",
		Hints:   hints,
		Explain: []string{explain},
		Answer: Answer{Num: pair(want), Text: spaced(want, st) + " " + toLabel,
			Given: pair(given), Factor: pair(want.Div(given))},
	}
	// Cups, tablespoons, and butter get a picture (the cup wall or sticks).
	vol := swapVolume(p.big, a)
	switch {
	case p.big == "c" && p.small == "T":
		return withVisual(o, wallVisual(vol, vol))
	case p.big == "stick" || p.small == "stick":
		return withVisual(o, butterVisual(vol, vol))
	}
	return o
}

// swapVolume is an amount of a big unit in eighth-teaspoons (0 if not a volume we draw).
func swapVolume(unit string, amount frac.Frac) int {
	size := map[string]int64{"c": Cup, "stick": Stick, "T": Tbsp}[unit]
	v := amount.Mul(frac.Int(size))
	if size == 0 || !v.IsInt() {
		return 0
	}
	return int(v.Num())
}
