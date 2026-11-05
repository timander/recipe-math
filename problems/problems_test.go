package problems

import (
	"slices"
	"strings"
	"testing"

	"github.com/timander/recipe-math/frac"
	"github.com/timander/recipe-math/units"
)

func TestEveryLevelMakesValidProblems(t *testing.T) {
	for _, l := range Levels {
		if PoolSize(l.ID) < 20 {
			t.Errorf("level %d can only make %d different problems", l.ID, PoolSize(l.ID))
		}
		for seed := range uint64(30) {
			ps, err := Generate([]int{l.ID}, 20, seed)
			if err != nil {
				t.Fatal(err)
			}
			if len(ps) != 20 {
				t.Fatalf("level %d: got %d problems", l.ID, len(ps))
			}
			seen := map[string]bool{}
			for _, p := range ps {
				checkProblem(t, l, p)
				if seen[p.Prompt()] {
					t.Errorf("level %d seed %d repeats %q", l.ID, seed, p.Prompt())
				}
				seen[p.Prompt()] = true
			}
		}
	}
}

func checkProblem(t *testing.T, l Level, p Problem) {
	t.Helper()
	from, to := units.MustGet(p.From), units.MustGet(p.To)
	want, approx, err := units.Convert(p.Given(), from, to)
	if err != nil || !want.Equal(p.Want()) || approx != p.Approx {
		t.Errorf("%s: answer %s is wrong", p.Prompt(), p.AnswerText)
	}
	if !slices.Contains(l.Dens, p.Given().Den()) || !slices.Contains(from.Dens(), p.Given().Den()) {
		t.Errorf("level %d: %s uses an unfriendly fraction", l.ID, p.GivenText)
	}
	if !slices.Contains(l.answerDens(), p.Want().Den()) || !slices.Contains(to.Dens(), p.Want().Den()) {
		t.Errorf("level %d: answer %s uses an unfriendly fraction", l.ID, p.AnswerText)
	}
	if p.Hint == "" || len(p.Steps) == 0 {
		t.Errorf("%s has no explanation", p.Prompt())
	}
	if last := p.Steps[len(p.Steps)-1]; !strings.Contains(strings.Join(p.Steps, " "), p.AnswerText) {
		t.Errorf("%s: steps never reach %s (last: %s)", p.Prompt(), p.AnswerText, last)
	}
	if r := Check(p, p.Answer); r.Status != Correct {
		t.Errorf("%s: the key's own answer %q was marked %s", p.Prompt(), p.Answer, r.Status)
	}
	if r := Check(p, p.AnswerText); r.Status != Correct {
		t.Errorf("%s: answer with unit %q was marked %s", p.Prompt(), p.AnswerText, r.Status)
	}
}

func TestSameSeedSameProblems(t *testing.T) {
	a, _ := Generate([]int{1, 3, 5}, 15, 4821)
	b, _ := Generate([]int{5, 1, 3, 3}, 15, 4821) // order and repeats don't matter
	c, _ := Generate([]int{1, 3, 5}, 15, 4822)
	if !slices.Equal(prompts(a), prompts(b)) {
		t.Error("same seed and levels gave different problems")
	}
	if slices.Equal(prompts(a), prompts(c)) {
		t.Error("different seeds gave the same problems")
	}
}

func TestMixedLevelsSplitEvenlyEasiestFirst(t *testing.T) {
	ps, _ := Generate([]int{4, 2}, 7, 1)
	var levels []int
	for _, p := range ps {
		levels = append(levels, p.Level)
	}
	if !slices.Equal(levels, []int{2, 2, 2, 2, 4, 4, 4}) {
		t.Errorf("levels = %v", levels)
	}
}

func TestGenerateErrors(t *testing.T) {
	if _, err := Generate(nil, 10, 1); err == nil {
		t.Error("no levels should fail")
	}
	if _, err := Generate([]int{9}, 10, 1); err == nil {
		t.Error("unknown level should fail")
	}
	if _, err := Generate([]int{1}, 0, 1); err == nil {
		t.Error("zero problems should fail")
	}
}

func TestBigWorksheetStillWorks(t *testing.T) {
	ps, err := Generate([]int{1}, 100, 3) // more than level 1's 32 problems
	if err != nil || len(ps) != 100 {
		t.Fatalf("got %d problems, %v", len(ps), err)
	}
}

func mustNew(t *testing.T, level int, from, to string, given frac.Frac) Problem {
	t.Helper()
	p, err := New(level, from, to, given)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestExplanation(t *testing.T) {
	p := mustNew(t, 5, "cup", "tsp", frac.New(3, 4))
	want := []string{
		"A cup is bigger than a teaspoon, so you need more teaspoons. Multiply.",
		"1 cup = 16 Tbsp, so 3/4 × 16 = 12 Tbsp.",
		"1 Tbsp = 3 tsp, so 12 × 3 = 36 tsp.",
		"Shortcut: 1 cup = 48 tsp.",
	}
	if !slices.Equal(p.Steps, want) {
		t.Errorf("steps =\n%s", strings.Join(p.Steps, "\n"))
	}
	if p.Hint != "This takes 2 steps: 1 cup = 16 Tbsp, then 1 Tbsp = 3 tsp. "+want[0] {
		t.Errorf("hint = %q", p.Hint)
	}

	p = mustNew(t, 2, "Tbsp", "cup", frac.Int(12))
	if p.Steps[1] != "16 Tbsp = 1 cup, so 12 ÷ 16 = 3/4 cup." || p.Prompt() != "12 Tbsp = ____ cup" {
		t.Errorf("divide steps = %q, prompt %q", p.Steps, p.Prompt())
	}

	p = mustNew(t, 5, "cup", "Tbsp", frac.New(1, 3))
	if p.AnswerText != "5 1/3 Tbsp" || p.Steps[len(p.Steps)-1] != "To measure it: 5 Tbsp + 1 tsp." {
		t.Errorf("thirds: %s %q", p.AnswerText, p.Steps)
	}

	p = mustNew(t, 6, "cup", "mL", frac.New(1, 2))
	if p.Prompt() != "1/2 cup ≈ ____ mL" || p.Steps[1] != "1 cup ≈ 240 mL, so 1/2 × 240 ≈ 120 mL." {
		t.Errorf("metric: %q %q", p.Prompt(), p.Steps)
	}
}

func TestCheck(t *testing.T) {
	cupTbsp := mustNew(t, 2, "cup", "Tbsp", frac.New(3, 4)) // 12
	tbspCup := mustNew(t, 2, "Tbsp", "cup", frac.Int(12))   // 3/4
	third := mustNew(t, 5, "cup", "Tbsp", frac.New(1, 3))   // 5 1/3
	cupTsp := mustNew(t, 5, "cup", "tsp", frac.New(3, 4))   // 36
	metric := mustNew(t, 6, "cup", "mL", frac.Int(2))       // 480 (really 473)

	cases := []struct {
		p       Problem
		input   string
		status  Status
		message string // substring
	}{
		{cupTbsp, "12", Correct, "Correct!"},
		{cupTbsp, "12 Tbsp", Correct, "Correct!"},
		{cupTbsp, "", Invalid, "Type your answer"},
		{cupTbsp, "twelve", Invalid, "can't read"},
		{cupTbsp, "3/64", Wrong, "Multiply instead of dividing"},
		{cupTbsp, "11", Wrong, "Not quite"},
		{tbspCup, "3/4", Correct, "Correct!"},
		{tbspCup, "0.75", Correct, "Correct!"},
		{tbspCup, ".75", Correct, "Correct!"},
		{tbspCup, "¾", Correct, "Correct!"},
		{tbspCup, "192", Wrong, "Divide instead of multiplying"},
		{tbspCup, "0.7", Wrong, "Not quite"},
		{third, "5 1/3", Correct, "Correct!"},
		{third, "16/3", Correct, "Correct!"},
		{third, "5.33", Correct, "As a fraction, that's 5 1/3 Tbsp"},
		{third, "5.3", Wrong, "Not quite"},
		{cupTsp, "12", Wrong, "Good start: that's 12 Tbsp. Keep going until you get to teaspoons."},
		{metric, "480", Correct, "Correct!"},
		{metric, "473", Correct, "Close enough"},
		{metric, "400", Wrong, "Not quite"},
	}
	for _, c := range cases {
		r := Check(c.p, c.input)
		if r.Status != c.status || !strings.Contains(r.Message, c.message) {
			t.Errorf("Check(%s, %q) = %s %q; want %s containing %q", c.p.Prompt(), c.input, r.Status, r.Message, c.status, c.message)
		}
	}
}

func TestWrongAnswersNeverRevealTheAnswer(t *testing.T) {
	for _, l := range Levels {
		ps, _ := Generate([]int{l.ID}, 20, 9)
		for _, p := range ps {
			for _, bad := range []string{"999", "1/64", p.Given().String()} {
				r := Check(p, bad)
				if r.Status == Wrong && strings.Contains(r.Message, p.AnswerText) {
					t.Errorf("%s: wrong-answer message gives away %s: %q", p.Prompt(), p.AnswerText, r.Message)
				}
			}
		}
	}
}

func prompts(ps []Problem) []string {
	var out []string
	for _, p := range ps {
		out = append(out, p.Prompt())
	}
	return out
}
