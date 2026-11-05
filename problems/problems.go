package problems

import (
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"sync"

	"github.com/timander/recipe-math/frac"
	"github.com/timander/recipe-math/units"
)

// Problem is one conversion: Given From = ___ To.
type Problem struct {
	Level     int    `json:"level"`
	From      string `json:"from"`
	To        string `json:"to"`
	GivenNum  int64  `json:"givenNum"`
	GivenDen  int64  `json:"givenDen"`
	AnswerNum int64  `json:"answerNum"`
	AnswerDen int64  `json:"answerDen"`
	Approx    bool   `json:"approx"`

	// Display text, ready for the screen or the printed page.
	GivenText  string   `json:"givenText"`  // "3/4 cup"
	Equals     string   `json:"equals"`     // "=" or "≈"
	ToLabel    string   `json:"toLabel"`    // "Tbsp"
	Answer     string   `json:"answer"`     // "12"
	AnswerText string   `json:"answerText"` // "12 Tbsp"
	Hint       string   `json:"hint"`
	Steps      []string `json:"steps"`
}

// Given and Want return the problem's amounts as fractions.
func (p Problem) Given() frac.Frac { return frac.New(p.GivenNum, p.GivenDen) }
func (p Problem) Want() frac.Frac  { return frac.New(p.AnswerNum, p.AnswerDen) }

// Prompt is the problem as one line: "3/4 cup = ____ Tbsp".
func (p Problem) Prompt() string {
	return fmt.Sprintf("%s %s ____ %s", p.GivenText, p.Equals, p.ToLabel)
}

// candidate is a valid (from, to, amount) choice before it is dressed up.
type candidate struct {
	from, to units.Unit
	given    frac.Frac
}

// pools holds, per level, the valid candidates grouped by direction
// (e.g. cup→Tbsp and Tbsp→cup are separate groups), smallest amounts first.
var (
	poolMu sync.Mutex
	pools  = map[int][][]candidate{}
)

func levelPool(l Level) [][]candidate {
	poolMu.Lock()
	defer poolMu.Unlock()
	if p, ok := pools[l.ID]; ok {
		return p
	}
	var groups [][]candidate
	for _, pair := range l.Pairs {
		a, b := units.MustGet(pair.A), units.MustGet(pair.B)
		for _, dir := range [][2]units.Unit{{a, b}, {b, a}} {
			from, to := dir[0], dir[1]
			var group []candidate
			for _, given := range amounts(from, l.Dens, l.max(from.ID, from.Max())) {
				ans, _, err := units.Convert(given, from, to)
				if err != nil {
					panic(err) // a level lists a pair that can't convert
				}
				if fits(ans, to, l.answerDens(), l.max(to.ID, to.Max())) {
					group = append(group, candidate{from, to, given})
				}
			}
			if len(group) > 0 {
				groups = append(groups, group)
			}
		}
	}
	pools[l.ID] = groups
	return groups
}

// amounts lists every friendly amount of a unit for a level, smallest first.
func amounts(u units.Unit, levelDens []int64, max int64) []frac.Frac {
	var out []frac.Frac
	for _, d := range allowedDens(u, levelDens) {
		for n := int64(1); n <= max*d; n++ {
			if f := frac.New(n, d); f.Den() == d {
				out = append(out, f)
			}
		}
	}
	slices.SortFunc(out, func(a, b frac.Frac) int { return a.Cmp(b) })
	return out
}

func allowedDens(u units.Unit, levelDens []int64) []int64 {
	var out []int64
	for _, d := range u.Dens() {
		if slices.Contains(levelDens, d) {
			out = append(out, d)
		}
	}
	return out
}

func fits(amount frac.Frac, u units.Unit, levelDens []int64, max int64) bool {
	return amount.Cmp(frac.Int(0)) > 0 &&
		amount.Cmp(frac.Int(max)) <= 0 &&
		slices.Contains(allowedDens(u, levelDens), amount.Den())
}

// PoolSize is how many different problems a level can make.
func PoolSize(levelID int) int {
	l, ok := LevelByID(levelID)
	if !ok {
		return 0
	}
	n := 0
	for _, g := range levelPool(l) {
		n += len(g)
	}
	return n
}

// Generate makes count problems spread evenly over the given levels, easiest
// level first. The same seed always gives the same problems, so a printed
// worksheet's version number recreates its answer key.
func Generate(levelIDs []int, count int, seed uint64) ([]Problem, error) {
	if len(levelIDs) == 0 {
		return nil, fmt.Errorf("pick at least one level")
	}
	if count < 1 || count > 100 {
		return nil, fmt.Errorf("count must be 1 to 100, got %d", count)
	}
	ids := slices.Clone(levelIDs)
	slices.Sort(ids)
	ids = slices.Compact(ids)
	rng := rand.New(rand.NewPCG(seed, 0x5eed_c0de))

	var out []Problem
	for i, id := range ids {
		l, ok := LevelByID(id)
		if !ok {
			return nil, fmt.Errorf("no level %d", id)
		}
		n := count / len(ids)
		if i < count%len(ids) {
			n++
		}
		for _, c := range pick(levelPool(l), n, rng) {
			p, err := build(l.ID, c)
			if err != nil {
				return nil, err
			}
			out = append(out, p)
		}
	}
	return out, nil
}

// pick chooses n candidates, rotating through the directions so a set mixes
// e.g. cups→Tbsp with Tbsp→cups, and leaning toward smaller amounts. It
// avoids repeats until the level runs out of different problems.
func pick(groups [][]candidate, n int, rng *rand.Rand) []candidate {
	used := map[string]bool{}
	key := func(c candidate) string { return c.from.ID + ">" + c.to.ID + ":" + c.given.String() }
	total := 0
	for _, g := range groups {
		total += len(g)
	}
	order := rng.Perm(len(groups))
	var out []candidate
	for i := 0; len(out) < n; i++ {
		if len(used) == total {
			clear(used) // every problem used once; allow repeats
		}
		g := groups[order[i%len(order)]]
		// Lean small: take the lower of two random positions.
		for range 20 {
			j := min(rng.IntN(len(g)), rng.IntN(len(g)))
			if !used[key(g[j])] {
				used[key(g[j])] = true
				out = append(out, g[j])
				break
			}
		}
		if i > n*50 { // a tiny level is exhausted; fill from anywhere
			g := groups[rng.IntN(len(groups))]
			out = append(out, g[rng.IntN(len(g))])
		}
	}
	rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// New builds one problem directly, e.g. New(2, "cup", "Tbsp", frac.New(3, 4)).
func New(level int, from, to string, given frac.Frac) (Problem, error) {
	f, ok1 := units.Get(from)
	t, ok2 := units.Get(to)
	if !ok1 || !ok2 {
		return Problem{}, fmt.Errorf("unknown unit %q or %q", from, to)
	}
	return build(level, candidate{f, t, given})
}

func build(level int, c candidate) (Problem, error) {
	ans, approx, err := units.Convert(c.given, c.from, c.to)
	if err != nil {
		return Problem{}, err
	}
	route, err := units.Route(c.from, c.to)
	if err != nil {
		return Problem{}, err
	}
	eq := "="
	if approx {
		eq = "≈"
	}
	p := Problem{
		Level: level, From: c.from.ID, To: c.to.ID,
		GivenNum: c.given.Num(), GivenDen: c.given.Den(),
		AnswerNum: ans.Num(), AnswerDen: ans.Den(),
		Approx:     approx,
		GivenText:  units.Format(c.given, c.from),
		Equals:     eq,
		ToLabel:    c.to.Label(ans),
		Answer:     units.FormatNumber(ans, c.to),
		AnswerText: units.Format(ans, c.to),
	}
	p.Hint, p.Steps = explain(c, ans, route, eq)
	return p, nil
}

// explain writes the hint (which facts to use) and the worked steps.
func explain(c candidate, ans frac.Frac, route []units.Step, eq string) (string, []string) {
	var facts []string
	allMul, allDiv := true, true
	for _, s := range route {
		facts = append(facts, s.Fact.String())
		allMul = allMul && s.Multiply
		allDiv = allDiv && !s.Multiply
	}

	var rule string
	switch {
	case allMul:
		rule = directionRule(c.from, c.to, true)
	case allDiv:
		rule = directionRule(c.from, c.to, false)
	}

	var hint string
	if len(facts) == 1 {
		hint = "Use " + facts[0] + "."
	} else {
		hint = fmt.Sprintf("This takes %d steps: %s.", len(facts), strings.Join(facts, ", then "))
	}
	if rule != "" {
		hint += " " + rule
	}

	steps := []string{}
	if rule != "" {
		steps = append(steps, rule)
	}
	v := c.given
	for _, s := range route {
		f := s.Fact
		big, small := units.MustGet(f.Big), units.MustGet(f.Small)
		factor := units.FormatNumber(f.Factor, small)
		var next frac.Frac
		if s.Multiply {
			next = v.Mul(f.Factor)
			steps = append(steps, fmt.Sprintf("%s, so %s × %s %s %s.",
				f.String(), units.FormatNumber(v, s.From), factor, eq, units.Format(next, s.To)))
		} else {
			next = v.Div(f.Factor)
			steps = append(steps, fmt.Sprintf("%s %s 1 %s, so %s ÷ %s %s %s.",
				units.Format(f.Factor, small), eq, big.Abbr, units.FormatNumber(v, s.From), factor, eq, units.Format(next, s.To)))
		}
		v = next
	}
	if len(route) > 1 && (allMul || allDiv) {
		bigger, smaller := c.from, c.to
		if allDiv {
			bigger, smaller = c.to, c.from
		}
		direct, _, _ := units.Convert(frac.Int(1), bigger, smaller)
		steps = append(steps, fmt.Sprintf("Shortcut: 1 %s = %s.", bigger.Abbr, units.Format(direct, smaller)))
	}
	if c.to.ID == "Tbsp" && ans.Den() == 3 {
		whole := ans.Num() / 3
		tsp := ans.Num() % 3
		steps = append(steps, fmt.Sprintf("To measure it: %s.", measureThirds(whole, tsp)))
	}
	return hint, steps
}

func measureThirds(wholeTbsp, tsp int64) string {
	t := "1 tsp"
	if tsp == 2 {
		t = "2 tsp"
	}
	if wholeTbsp == 0 {
		return t
	}
	return fmt.Sprintf("%d Tbsp + %s", wholeTbsp, t)
}

// Status is how an answer was judged.
type Status string

const (
	Correct Status = "correct"
	Wrong   Status = "wrong"
	Invalid Status = "invalid" // not a number we could read
)

// Result is the verdict on one answer. Message never gives the answer away
// on a wrong try, so the student can try again.
type Result struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
}

// approxTolerance accepts US ↔ metric answers within 3%, so both the
// rounded kitchen value (240 mL per cup) and the true one (237 mL) count.
const approxTolerance = 0.03

// Check judges a student's typed answer.
func Check(p Problem, input string) Result {
	if strings.TrimSpace(input) == "" {
		return Result{Invalid, "Type your answer first."}
	}
	got, err := frac.Parse(input)
	if err != nil {
		return Result{Invalid, "Hmm, I can't read that. Type a number like 12, 3/4, 1 1/2, or 1.5."}
	}
	want := p.Want()
	to := units.MustGet(p.To)
	from := units.MustGet(p.From)

	if got.Equal(want) {
		return Result{Correct, "Correct!"}
	}
	if p.Approx && near(got, want, approxTolerance) {
		return Result{Correct, fmt.Sprintf("Close enough! Using the kitchen value, it's %s.", p.AnswerText)}
	}
	isDecimal := strings.Contains(input, ".")
	if !p.Approx && isDecimal && math.Abs(got.Float64()-want.Float64()) < 0.01 {
		return Result{Correct, fmt.Sprintf("Correct! As a fraction, that's %s.", p.AnswerText)}
	}

	// Did they multiply when they should divide, or the other way round?
	given := p.Given()
	flipped := given.Mul(given).Div(want)
	if near(got, flipped, 0.01) {
		multiply := want.Cmp(given) > 0
		fix := "Divide instead of multiplying."
		if multiply {
			fix = "Multiply instead of dividing."
		}
		return Result{Wrong, "Check your direction. " + directionRule(from, to, multiply) + " " + fix}
	}

	// Did they stop halfway through a two-step problem?
	if route, err := units.Route(from, to); err == nil && len(route) > 1 {
		v := given
		for _, s := range route[:len(route)-1] {
			v, _, _ = units.Convert(v, s.From, s.To)
			if got.Equal(v) {
				return Result{Wrong, fmt.Sprintf("Good start: that's %s. Keep going until you get to %s.",
					units.Format(v, s.To), to.Plural)}
			}
		}
	}
	return Result{Wrong, "Not quite. Try again, or tap Hint."}
}

func near(got, want frac.Frac, tol float64) bool {
	w := want.Float64()
	return math.Abs(got.Float64()-w) <= tol*math.Abs(w)
}

// directionRule explains multiply vs divide in words:
// "A cup is bigger than a tablespoon, so you need more tablespoons."
func directionRule(from, to units.Unit, multiply bool) string {
	if multiply {
		return fmt.Sprintf("%s is bigger than %s, so you need more %s. Multiply.",
			capitalize(article(from.Name)), article(to.Name), to.Plural)
	}
	return fmt.Sprintf("%s is smaller than %s, so you need fewer %s. Divide.",
		capitalize(article(from.Name)), article(to.Name), to.Plural)
}

func article(word string) string {
	if strings.ContainsRune("aeiou", rune(word[0])) {
		return "an " + word
	}
	return "a " + word
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
