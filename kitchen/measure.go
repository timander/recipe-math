// Package kitchen is the v2 recipe-math engine: real measuring tools,
// realistic recipes, and the orders (problems) a student works on a shift.
package kitchen

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/timander/recipe-math/frac"
)

// Volumes are whole numbers of eighth-teaspoons, the smallest spoon in the
// drawer. Every real tool is a whole number of them, so the math stays exact.
const (
	EighthTsp  = 1
	QuarterTsp = 2
	HalfTsp    = 4
	Tsp        = 8
	Tbsp       = 24
	FlOz       = 48
	QuarterCup = 96
	ThirdCup   = 128
	HalfCup    = 192
	Cup        = 384
	Stick      = HalfCup // one stick of butter
	Pint       = 2 * Cup
	Quart      = 4 * Cup
)

// ToolKind groups tools the way a kitchen drawer does.
type ToolKind string

const (
	Spoon  ToolKind = "spoon"
	DryCup ToolKind = "dry"
	Liquid ToolKind = "liquid" // a fill line on a glass liquid measuring cup
	Butter ToolKind = "butter" // a stick, or a tablespoon mark on the wrapper
)

// Tool is one measuring tool, or one fill line on the liquid cup.
type Tool struct {
	ID    string   `json:"id"`
	Label string   `json:"label"` // "¼ c"
	Name  string   `json:"name"`  // "¼ cup dry measure"
	Kind  ToolKind `json:"kind"`
	Size  int      `json:"size"` // eighth-teaspoons
}

var spoons = []Tool{
	{"T1", "1 T", "tablespoon", Spoon, Tbsp},
	{"t1", "1 t", "teaspoon", Spoon, Tsp},
	{"t12", "½ t", "½ teaspoon", Spoon, HalfTsp},
	{"t14", "¼ t", "¼ teaspoon", Spoon, QuarterTsp},
	{"t18", "⅛ t", "⅛ teaspoon", Spoon, EighthTsp},
}

var dryCups = []Tool{
	{"c1", "1 c", "1 cup dry measure", DryCup, Cup},
	{"c12", "½ c", "½ cup dry measure", DryCup, HalfCup},
	{"c13", "⅓ c", "⅓ cup dry measure", DryCup, ThirdCup},
	{"c14", "¼ c", "¼ cup dry measure", DryCup, QuarterCup},
}

var butterTools = []Tool{
	{"stick", "1 stick", "stick of butter (½ c)", Butter, Stick},
	{"bT", "1 T mark", "tablespoon mark on the wrapper", Butter, Tbsp},
}

// liquidLines are the marks on a 4-cup glass measuring cup: every ¼ and ⅓ cup.
var liquidLines = func() []Tool {
	seen := map[int]bool{}
	var out []Tool
	for _, d := range []int{4, 3} {
		for n := 1; n <= 4*d; n++ {
			size := Cup * n / d
			if seen[size] {
				continue
			}
			seen[size] = true
			out = append(out, Tool{
				ID: fmt.Sprintf("L%d", size), Label: CupText(size), Name: CupText(size) + " line on the liquid cup",
				Kind: Liquid, Size: size,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Size < out[j].Size })
	return out
}()

var toolByID = func() map[string]Tool {
	m := map[string]Tool{}
	for _, list := range [][]Tool{spoons, dryCups, butterTools, liquidLines} {
		for _, t := range list {
			m[t.ID] = t
		}
	}
	return m
}()

// Drawer names which tools a Measure It order offers.
type Drawer string

const (
	DrawerDry    Drawer = "dry"    // dry cups + spoons
	DrawerLiquid Drawer = "liquid" // one liquid-cup line + spoons
	DrawerButter Drawer = "butter" // sticks + wrapper marks
	DrawerSpoons Drawer = "spoons" // spoons only
)

// Tools lists the tools in a drawer, biggest first.
func (d Drawer) Tools() []Tool {
	switch d {
	case DrawerDry:
		return append(slices.Clone(dryCups), spoons...)
	case DrawerLiquid:
		return append(slices.Clone(liquidLines), spoons...)
	case DrawerButter:
		return slices.Clone(butterTools)
	}
	return slices.Clone(spoons)
}

// Solve finds the fewest scoops that measure an amount exactly with a
// drawer's tools. ok is false when the drawer can't make it exactly.
// A liquid cup counts as one scoop and can be used at most once.
func Solve(amount int, d Drawer) (best []Tool, ok bool) {
	if amount <= 0 {
		return nil, false
	}
	switch d {
	case DrawerLiquid:
		best, ok = coins(amount, spoons) // spoons only, for small amounts
		for _, line := range liquidLines {
			if line.Size > amount {
				break
			}
			rest, restOK := []Tool(nil), true
			if line.Size < amount {
				rest, restOK = coins(amount-line.Size, spoons)
			}
			if restOK && (!ok || 1+len(rest) < len(best)) {
				best, ok = append([]Tool{line}, rest...), true
			}
		}
		return best, ok
	case DrawerButter:
		return coins(amount, butterTools)
	case DrawerDry:
		return coins(amount, append(slices.Clone(dryCups), spoons...))
	}
	return coins(amount, spoons)
}

// coins is the classic fewest-coins problem, solved exactly.
func coins(amount int, tools []Tool) ([]Tool, bool) {
	const inf = 1 << 30
	count := make([]int, amount+1)
	last := make([]int, amount+1)
	smallest := make([]int, amount+1) // smallest tool used, for breaking ties
	smallest[0] = inf
	for a := 1; a <= amount; a++ {
		count[a] = inf
		for i, t := range tools {
			if t.Size > a || count[a-t.Size] >= inf {
				continue
			}
			c, sm := count[a-t.Size]+1, min(smallest[a-t.Size], t.Size)
			// Fewest scoops first; on a tie, prefer bigger spoons (¼ c + 2 T
			// over ⅓ c + 2 t), which is how cooks think.
			if c < count[a] || (c == count[a] && sm > smallest[a]) {
				count[a], last[a], smallest[a] = c, i, sm
			}
		}
	}
	if count[amount] >= inf {
		return nil, false
	}
	var out []Tool
	for a := amount; a > 0; a -= tools[last[a]].Size {
		out = append(out, tools[last[a]])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Size > out[j].Size })
	return out, true
}

// Combo writes scoops the way a cook says them: "¼ c + 2 T + 1 t".
// Repeated spoons read as "2 T"; repeated cups as "2 × ⅓ c" so it's clear
// that's two scoops.
func Combo(tools []Tool) string {
	if len(tools) == 0 {
		return "nothing"
	}
	sorted := slices.Clone(tools)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Size > sorted[j].Size })
	var parts []string
	for i := 0; i < len(sorted); {
		t, k := sorted[i], 0
		for i < len(sorted) && sorted[i].ID == t.ID {
			i, k = i+1, k+1
		}
		parts = append(parts, scoops(t, k))
	}
	return strings.Join(parts, " + ")
}

func scoops(t Tool, k int) string {
	label := t.Label
	if t.Kind == Liquid {
		label += " (liquid cup)"
	}
	if k == 1 {
		return label
	}
	switch t.ID {
	case "T1", "t1", "c1":
		return fmt.Sprintf("%d %s", k, strings.TrimPrefix(label, "1 "))
	case "stick":
		return fmt.Sprintf("%d sticks", k)
	case "bT":
		return fmt.Sprintf("%d T marks", k)
	}
	return fmt.Sprintf("%d × %s", k, label)
}

// Total adds up a list of tool IDs. It reports unknown IDs and, for the
// liquid cup, more than one fill line.
func Total(ids []string) (int, []Tool, error) {
	sum, lines := 0, 0
	var tools []Tool
	for _, id := range ids {
		t, ok := toolByID[id]
		if !ok {
			return 0, nil, fmt.Errorf("unknown tool %q", id)
		}
		if t.Kind == Liquid {
			lines++
		}
		sum += t.Size
		tools = append(tools, t)
	}
	if lines > 1 {
		return 0, nil, fmt.Errorf("a liquid cup can only be filled to one line")
	}
	return sum, tools, nil
}

// ---------- Writing amounts ----------

var vulgar = map[[2]int64]string{
	{1, 8}: "⅛", {1, 4}: "¼", {3, 8}: "⅜", {1, 2}: "½", {5, 8}: "⅝", {3, 4}: "¾", {7, 8}: "⅞",
	{1, 3}: "⅓", {2, 3}: "⅔", {1, 6}: "⅙", {5, 6}: "⅚",
}

// Mixed writes a fraction as a cook would: "2¼", "⅓", "3".
func Mixed(f frac.Frac) string {
	n, d := f.Num(), f.Den()
	whole, rem := n/d, n%d
	if rem == 0 {
		return fmt.Sprint(whole)
	}
	part, ok := vulgar[[2]int64{rem, d}]
	if !ok {
		part = fmt.Sprintf("%d/%d", rem, d)
		if whole > 0 {
			return fmt.Sprintf("%d %s", whole, part)
		}
		return part
	}
	if whole == 0 {
		return part
	}
	return fmt.Sprintf("%d%s", whole, part)
}

// CupText writes a volume in cups: "1½ c".
func CupText(e int) string { return Mixed(frac.New(int64(e), Cup)) + " c" }

// Style is how a recipe spells its units.
type Style int

const (
	Short Style = iota // T, t, c
	Abbr               // Tbsp, tsp, cup
)

// Text writes a volume the way a recipe would, using the largest unit that
// reads naturally: "2¼ c", "3 T", "1½ t", "4 t". It can return a combo
// ("1 T + 1 t") when no single unit fits, e.g. for a scaled amount.
func Text(e int, s Style) string {
	tb, ts, cu := "T", "t", "c"
	if s == Abbr {
		tb, ts, cu = "Tbsp", "tsp", "cup"
		if e > Cup {
			cu = "cups"
		}
	}
	cups := frac.New(int64(e), Cup)
	if e >= QuarterCup && slices.Contains([]int64{1, 2, 3, 4}, cups.Den()) {
		return Mixed(cups) + " " + cu
	}
	if e%Tbsp == 0 {
		return fmt.Sprintf("%d %s", e/Tbsp, tb)
	}
	t := frac.New(int64(e), Tsp)
	if t.Den() <= 8 && (e < Tbsp || (e%Tsp == 0 && e/Tsp <= 6)) {
		return Mixed(t) + " " + ts
	}
	best, _ := Solve(e, DrawerDry)
	return Combo(best)
}

// Oz writes a weight: "12 oz", "1½ lb".
func Oz(oz frac.Frac, inPounds bool) string {
	if inPounds {
		return Mixed(oz.Div(frac.Int(16))) + " lb"
	}
	return Mixed(oz) + " oz"
}
