// Package units defines the kitchen units students learn and the
// equivalents ("facts") that connect them.
//
// Inside one system (US or metric) every conversion is exact. Between
// systems the facts are the rounded kitchen values used on US nutrition
// labels and in most foods textbooks, so those conversions are approximate.
package units

import (
	"fmt"

	"github.com/timander/recipe-math/frac"
)

type Dimension string

const (
	Volume Dimension = "volume" // how much space: spoons, cups, mL
	Weight Dimension = "weight" // how heavy: ounces, pounds, grams
)

type System string

const (
	US     System = "US"
	Metric System = "metric"
)

// Unit is one kitchen unit.
type Unit struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`   // "tablespoon"
	Plural    string    `json:"plural"` // "tablespoons"
	Abbr      string    `json:"abbr"`   // shown in problems: "Tbsp"
	AbbrPl    string    `json:"abbrPlural"`
	OtherAbbr []string  `json:"otherAbbr"` // other spellings seen in recipes
	Dimension Dimension `json:"dimension"`
	System    System    `json:"system"`
	Tool      string    `json:"tool"` // how you'd measure it

	// perBase is how many base units fit in one of this unit. The bases are
	// teaspoon (US volume), ounce (US weight), milliliter, and gram.
	perBase frac.Frac
	// dens are the denominators that make sense for this unit in a problem
	// (nobody measures 1/3 of a gallon or 7/10 of a teaspoon).
	dens []int64
	// max is the largest amount of this unit that appears in a problem.
	max int64
}

// Dens returns the denominators that read naturally for this unit.
func (u Unit) Dens() []int64 { return u.dens }

// Max returns the largest amount of this unit used in problems.
func (u Unit) Max() int64 { return u.max }

// Label returns the unit as written after an amount: "1 cup", "2 cups".
func (u Unit) Label(amount frac.Frac) string {
	if amount.Cmp(frac.Int(1)) > 0 {
		return u.AbbrPl
	}
	return u.Abbr
}

// All lists every unit, smallest to largest within each group.
var All = []Unit{
	// US volume, base = teaspoon
	{ID: "tsp", Name: "teaspoon", Plural: "teaspoons", Abbr: "tsp", AbbrPl: "tsp", OtherAbbr: []string{"t"},
		Dimension: Volume, System: US, Tool: "measuring spoon", perBase: frac.Int(1), dens: []int64{1, 2, 4, 8}, max: 48},
	{ID: "Tbsp", Name: "tablespoon", Plural: "tablespoons", Abbr: "Tbsp", AbbrPl: "Tbsp", OtherAbbr: []string{"T", "tbsp", "tbs"},
		Dimension: Volume, System: US, Tool: "measuring spoon", perBase: frac.Int(3), dens: []int64{1, 2, 3}, max: 32},
	{ID: "floz", Name: "fluid ounce", Plural: "fluid ounces", Abbr: "fl oz", AbbrPl: "fl oz",
		Dimension: Volume, System: US, Tool: "liquid measuring cup", perBase: frac.Int(6), dens: []int64{1}, max: 32},
	{ID: "cup", Name: "cup", Plural: "cups", Abbr: "cup", AbbrPl: "cups", OtherAbbr: []string{"c", "C"},
		Dimension: Volume, System: US, Tool: "measuring cup", perBase: frac.Int(48), dens: []int64{1, 2, 3, 4}, max: 16},
	{ID: "pint", Name: "pint", Plural: "pints", Abbr: "pint", AbbrPl: "pints", OtherAbbr: []string{"pt"},
		Dimension: Volume, System: US, Tool: "liquid measuring cup", perBase: frac.Int(96), dens: []int64{1, 2}, max: 8},
	{ID: "quart", Name: "quart", Plural: "quarts", Abbr: "quart", AbbrPl: "quarts", OtherAbbr: []string{"qt"},
		Dimension: Volume, System: US, Tool: "large liquid measure", perBase: frac.Int(192), dens: []int64{1, 2, 4}, max: 8},
	{ID: "gallon", Name: "gallon", Plural: "gallons", Abbr: "gallon", AbbrPl: "gallons", OtherAbbr: []string{"gal"},
		Dimension: Volume, System: US, Tool: "jug or stockpot", perBase: frac.Int(768), dens: []int64{1, 2, 4}, max: 4},

	// US weight, base = ounce
	{ID: "oz", Name: "ounce", Plural: "ounces", Abbr: "oz", AbbrPl: "oz",
		Dimension: Weight, System: US, Tool: "kitchen scale", perBase: frac.Int(1), dens: []int64{1, 2, 4}, max: 64},
	{ID: "lb", Name: "pound", Plural: "pounds", Abbr: "lb", AbbrPl: "lb", OtherAbbr: []string{"lbs", "#"},
		Dimension: Weight, System: US, Tool: "kitchen scale", perBase: frac.Int(16), dens: []int64{1, 2, 4, 5, 10}, max: 12},

	// Metric volume, base = milliliter
	{ID: "mL", Name: "milliliter", Plural: "milliliters", Abbr: "mL", AbbrPl: "mL", OtherAbbr: []string{"ml"},
		Dimension: Volume, System: Metric, Tool: "metric measuring cup or spoon", perBase: frac.Int(1), dens: []int64{1}, max: 4000},
	{ID: "L", Name: "liter", Plural: "liters", Abbr: "L", AbbrPl: "L", OtherAbbr: []string{"l"},
		Dimension: Volume, System: Metric, Tool: "liter pitcher", perBase: frac.Int(1000), dens: []int64{1, 2, 4, 5, 10}, max: 4},

	// Metric weight, base = gram
	{ID: "g", Name: "gram", Plural: "grams", Abbr: "g", AbbrPl: "g",
		Dimension: Weight, System: Metric, Tool: "kitchen scale", perBase: frac.Int(1), dens: []int64{1}, max: 5000},
	{ID: "kg", Name: "kilogram", Plural: "kilograms", Abbr: "kg", AbbrPl: "kg",
		Dimension: Weight, System: Metric, Tool: "kitchen scale", perBase: frac.Int(1000), dens: []int64{1, 2, 4, 5, 10}, max: 5},
}

var byID = func() map[string]Unit {
	m := make(map[string]Unit, len(All))
	for _, u := range All {
		m[u.ID] = u
	}
	return m
}()

// Get returns the unit with the given ID.
func Get(id string) (Unit, bool) {
	u, ok := byID[id]
	return u, ok
}

// MustGet is Get for IDs known at compile time.
func MustGet(id string) Unit {
	u, ok := byID[id]
	if !ok {
		panic("units: unknown unit " + id)
	}
	return u
}

// Fact is one equivalent students learn: 1 Big = Factor Small.
type Fact struct {
	Big    string    `json:"big"`
	Small  string    `json:"small"`
	Factor frac.Frac `json:"-"`
	Approx bool      `json:"approx"` // true for US ↔ metric
}

// String reads like the cheat sheet: "1 cup = 16 Tbsp", "1 cup ≈ 240 mL".
func (f Fact) String() string {
	eq := "="
	if f.Approx {
		eq = "≈"
	}
	big, small := MustGet(f.Big), MustGet(f.Small)
	return fmt.Sprintf("1 %s %s %s", big.Abbr, eq, Format(f.Factor, small))
}

// Facts are the equivalents on the cheat sheet. The order matters: when two
// routes between units are equally short, the earlier fact is used to
// explain the answer.
var Facts = []Fact{
	{Big: "Tbsp", Small: "tsp", Factor: frac.Int(3)},
	{Big: "cup", Small: "Tbsp", Factor: frac.Int(16)},
	{Big: "floz", Small: "Tbsp", Factor: frac.Int(2)},
	{Big: "cup", Small: "floz", Factor: frac.Int(8)},
	{Big: "pint", Small: "cup", Factor: frac.Int(2)},
	{Big: "quart", Small: "pint", Factor: frac.Int(2)},
	{Big: "quart", Small: "cup", Factor: frac.Int(4)},
	{Big: "gallon", Small: "quart", Factor: frac.Int(4)},
	{Big: "lb", Small: "oz", Factor: frac.Int(16)},
	{Big: "L", Small: "mL", Factor: frac.Int(1000)},
	{Big: "kg", Small: "g", Factor: frac.Int(1000)},

	// US ↔ metric, rounded kitchen values
	{Big: "tsp", Small: "mL", Factor: frac.Int(5), Approx: true},
	{Big: "Tbsp", Small: "mL", Factor: frac.Int(15), Approx: true},
	{Big: "floz", Small: "mL", Factor: frac.Int(30), Approx: true},
	{Big: "cup", Small: "mL", Factor: frac.Int(240), Approx: true},
	{Big: "oz", Small: "g", Factor: frac.Int(28), Approx: true},
	{Big: "lb", Small: "g", Factor: frac.Int(454), Approx: true},
	{Big: "kg", Small: "lb", Factor: frac.New(11, 5), Approx: true},
}

// Tips appear on the cheat sheet and in hints.
var Tips = []string{
	"Bigger unit → smaller unit: the number gets bigger, so multiply.",
	"Smaller unit → bigger unit: the number gets smaller, so divide.",
	"Capital T is tablespoon (Tbsp). Lowercase t is teaspoon (tsp). A tablespoon is 3 times bigger.",
	"4 Tbsp = 1/4 cup. Count by 4s: 8 Tbsp = 1/2 cup, 12 Tbsp = 3/4 cup, 16 Tbsp = 1 cup.",
	"Fluid ounces (fl oz) measure space. Ounces (oz) measure weight. 1 cup of water weighs about 8 oz, but 1 cup of flour weighs only about 4 1/2 oz.",
	"US ↔ metric values are rounded (≈). A cup is really about 237 mL, but recipes and food labels use 240 mL.",
}

// Format writes an amount the way it reads best for its unit: decimals for
// metric ("0.75 L") and tenths ("4.4 lb"), fractions for US ("1 1/2 cups").
func Format(amount frac.Frac, u Unit) string {
	return FormatNumber(amount, u) + " " + u.Label(amount)
}

// FormatNumber is Format without the unit.
func FormatNumber(amount frac.Frac, u Unit) string {
	if u.System == Metric || amount.Den()%5 == 0 {
		if s, ok := amount.Decimal(); ok {
			return s
		}
	}
	return amount.String()
}

// Convert changes an amount from one unit to another. Within a system it is
// exact. Between systems it uses the one approximate fact that links the two
// units directly, and reports approx = true.
func Convert(amount frac.Frac, from, to Unit) (result frac.Frac, approx bool, err error) {
	if from.Dimension != to.Dimension {
		return frac.Frac{}, false, fmt.Errorf("can't convert %s (%s) to %s (%s)", from.Name, from.Dimension, to.Name, to.Dimension)
	}
	if from.System == to.System {
		return amount.Mul(from.perBase).Div(to.perBase), false, nil
	}
	for _, f := range Facts {
		switch {
		case f.Approx && f.Big == from.ID && f.Small == to.ID:
			return amount.Mul(f.Factor), true, nil
		case f.Approx && f.Small == from.ID && f.Big == to.ID:
			return amount.Div(f.Factor), true, nil
		}
	}
	return frac.Frac{}, false, fmt.Errorf("no kitchen equivalent links %s and %s", from.Name, to.Name)
}

// Step is one hop in a worked conversion.
type Step struct {
	Fact     Fact
	From, To Unit
	Multiply bool // true when going from the big unit to the small one
}

// Route finds the fewest cheat-sheet facts that connect two units, so the
// worked answer uses equivalents students already know. Same-system routes
// use exact facts only; cross-system routes use one approximate fact.
func Route(from, to Unit) ([]Step, error) {
	if from.ID == to.ID {
		return nil, nil
	}
	approx := from.System != to.System
	type node struct {
		id   string
		path []Step
	}
	seen := map[string]bool{from.ID: true}
	queue := []node{{id: from.ID}}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, f := range Facts {
			if f.Approx != approx {
				continue
			}
			var next string
			var multiply bool
			switch cur.id {
			case f.Big:
				next, multiply = f.Small, true
			case f.Small:
				next, multiply = f.Big, false
			default:
				continue
			}
			if seen[next] {
				continue
			}
			seen[next] = true
			path := append(append([]Step(nil), cur.path...), Step{
				Fact: f, From: MustGet(cur.id), To: MustGet(next), Multiply: multiply,
			})
			if next == to.ID {
				return path, nil
			}
			if !approx { // cross-system routes are a single fact
				queue = append(queue, node{id: next, path: path})
			}
		}
	}
	return nil, fmt.Errorf("no route from %s to %s", from.Name, to.Name)
}
