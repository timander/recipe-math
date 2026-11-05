package kitchen

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/timander/recipe-math/frac"
)

func TestSolve(t *testing.T) {
	cases := []struct {
		amount int
		d      Drawer
		want   string
	}{
		{tb(5), DrawerDry, "¼ c + 1 T"},
		{ts(4, 1), DrawerDry, "1 T + 1 t"}, // the teacher's example: 4 t = 1 T + 1 t
		{c(1, 3), DrawerDry, "⅓ c"},
		{tb(12), DrawerDry, "½ c + ¼ c"},
		{c(2, 3), DrawerDry, "2 × ⅓ c"},
		{c(9, 4), DrawerDry, "2 c + ¼ c"},
		{ts(3, 4), DrawerDry, "½ t + ¼ t"},
		{tb(10), DrawerLiquid, "½ c (liquid cup) + 2 T"},
		{12 * FlOz, DrawerLiquid, "1½ c (liquid cup)"},
		{tb(3), DrawerLiquid, "3 T"},
		{c(3, 4), DrawerButter, "1 stick + 4 T marks"},
		{tb(6), DrawerButter, "6 T marks"},
	}
	for _, c := range cases {
		best, ok := Solve(c.amount, c.d)
		if !ok || Combo(best) != c.want {
			t.Errorf("Solve(%d, %s) = %q, want %q", c.amount, c.d, Combo(best), c.want)
		}
	}
	if _, ok := Solve(5, DrawerButter); ok {
		t.Error("butter wrapper can't measure 5/8 t")
	}
	if _, ok := Solve(0, DrawerDry); ok {
		t.Error("zero should not be measurable")
	}
}

func TestTotal(t *testing.T) {
	sum, _, err := Total([]string{"c14", "T1", "t1"})
	if err != nil || sum != QuarterCup+Tbsp+Tsp {
		t.Errorf("Total = %d, %v", sum, err)
	}
	if _, _, err := Total([]string{"L192", "L96"}); err == nil {
		t.Error("two liquid lines should fail")
	}
	if _, _, err := Total([]string{"spatula"}); err == nil {
		t.Error("unknown tool should fail")
	}
}

func TestText(t *testing.T) {
	cases := map[int]string{
		c(9, 4): "2¼ c", c(1, 3): "⅓ c", tb(3): "3 T", ts(1, 2): "½ t", ts(4, 1): "4 t", ts(3, 2): "1½ t",
		tb(5): "5 T", c(1, 1): "1 c", tb(2) + ts(2, 1): "2 T + 2 t", ts(1, 8): "⅛ t",
	}
	for e, want := range cases {
		if got := Text(e, Short); got != want {
			t.Errorf("Text(%d) = %q, want %q", e, got, want)
		}
	}
	if got := Text(c(3, 2), Abbr); got != "1½ cups" {
		t.Errorf("Abbr = %q", got)
	}
	if got := Oz(frac.Int(24), true); got != "1½ lb" {
		t.Errorf("Oz = %q", got)
	}
}

func TestRecipesAreRealistic(t *testing.T) {
	for _, r := range Recipes {
		for _, l := range r.Lines {
			in, ok := ing[l.Ing]
			if !ok {
				t.Fatalf("%s: unknown ingredient %s", r.Name, l.Ing)
			}
			if in.Kind != Weight && in.Kind != Count {
				if _, ok := Solve(l.Vol, drawerFor(l.Ing)); !ok {
					t.Errorf("%s: %s can't be measured", r.Name, lineText(l, Short))
				}
				if l.Vol > 4*Cup {
					t.Errorf("%s: %s is too big", r.Name, lineText(l, Short))
				}
			}
		}
	}
}

// correctResponse builds the right answer for any order.
func correctResponse(o Order) Response {
	switch o.Input {
	case InChoice:
		return Response{Choice: o.Answer.Choices[0]}
	case InBuild:
		best, _ := Solve(o.Answer.Vol, o.Drawer)
		var ids []string
		for _, t := range best {
			ids = append(ids, t.ID)
		}
		return Response{Tools: ids}
	}
	return Response{Text: Mixed(unpair(o.Answer.Num))}
}

func TestEveryRankEveryOrder(t *testing.T) {
	for _, rank := range Ranks {
		for seed := range uint64(60) {
			orders, err := Shift(rank.ID, seed)
			if err != nil {
				t.Fatal(err)
			}
			if len(orders) < 8 {
				t.Errorf("rank %d seed %d: only %d orders", rank.ID, seed, len(orders))
			}
			seen := map[string]bool{}
			for _, o := range orders {
				checkOrder(t, rank.ID, o)
				key := o.Prompt + fmt.Sprint(o.Line)
				if seen[key] {
					t.Errorf("rank %d seed %d repeats %q", rank.ID, seed, o.Prompt)
				}
				seen[key] = true
			}
		}
	}
}

func checkOrder(t *testing.T, rank int, o Order) {
	t.Helper()
	where := fmt.Sprintf("rank %d %s %q", rank, o.Kind, o.Prompt)
	if o.Prompt == "" || o.Station == "" || o.Seconds <= 0 || len(o.Explain) == 0 || o.Answer.Text == "" {
		t.Errorf("%s: missing a field", where)
	}
	if len(o.Hints) < 2 {
		t.Errorf("%s: needs a hint ladder, has %d", where, len(o.Hints))
	}
	for _, h := range o.Hints {
		if !strings.Contains(h, "?") {
			t.Errorf("%s: hint is not a question: %q", where, h)
		}
		if o.Input != InChoice && strings.Contains(h, o.Answer.Text) {
			t.Errorf("%s: hint gives away %q: %q", where, o.Answer.Text, h)
		}
	}
	if r := Check(o, correctResponse(o)); !r.Correct {
		t.Errorf("%s: the right answer was marked wrong: %s", where, r.Message)
	}
	if o.Input == InBuild {
		if o.Answer.Vol > 8*Cup || o.Answer.Vol <= 0 {
			t.Errorf("%s: unrealistic amount %d", where, o.Answer.Vol)
		}
		if len(o.Tools) == 0 {
			t.Errorf("%s: empty drawer", where)
		}
	}
	// Wrong answers never reveal the answer.
	for _, bad := range wrongResponses(o) {
		r := Check(o, bad)
		if r.Correct {
			continue
		}
		if r.Invalid {
			t.Errorf("%s: wrong answer %+v was unreadable", where, bad)
		}
		if strings.Contains(r.Message, o.Answer.Text) {
			t.Errorf("%s: wrong-answer message gives away %q: %q", where, o.Answer.Text, r.Message)
		}
	}
}

func wrongResponses(o Order) []Response {
	switch o.Input {
	case InChoice:
		var out []Response
		for i := range o.Choices {
			if !slices.Contains(o.Answer.Choices, i) {
				out = append(out, Response{Choice: i})
			}
		}
		return out
	case InBuild:
		return []Response{{Tools: []string{"t18"}}, {Tools: []string{"c1", "c1", "c1", "c1", "c1", "c1", "c1", "c1", "c1"}}}
	}
	return []Response{{Text: "999"}, {Text: "1/64"}}
}

func TestSameSeedSameShift(t *testing.T) {
	for _, rank := range Ranks {
		a, _ := Shift(rank.ID, 77)
		b, _ := Shift(rank.ID, 77)
		c, _ := Shift(rank.ID, 78)
		if prompts(a) != prompts(b) {
			t.Errorf("rank %d: same seed, different shift", rank.ID)
		}
		if prompts(a) == prompts(c) {
			t.Errorf("rank %d: different seeds, same shift", rank.ID)
		}
	}
	if _, err := Shift(9, 1); err == nil {
		t.Error("rank 9 should fail")
	}
}

func prompts(os []Order) string {
	var b strings.Builder
	for _, o := range os {
		b.WriteString(o.Prompt + "|")
	}
	return b.String()
}

func TestRecipeCardShift(t *testing.T) {
	orders, _ := Shift(4, 5)
	var card []Order
	for _, o := range orders {
		if o.Card != nil {
			card = append(card, o)
		}
	}
	if len(card) < 6 || card[0].Kind != KFactor {
		t.Fatalf("card should start with the factor and cover every line, got %d orders", len(card))
	}
	if len(card) != len(card[0].Card.Lines)+1 {
		t.Errorf("card has %d lines but %d orders", len(card[0].Card.Lines), len(card))
	}
}

func TestCheckFeedback(t *testing.T) {
	g := &gen{rng: newRNG(1)}
	r := Recipes[0] // cookies, makes 24
	f := frac.New(5, 2)
	fo := g.factorFor(r, f, 60)
	cases := []struct {
		o    Order
		resp Response
		want string
	}{
		{fo, Response{Text: "2.5"}, "Correct!"},
		{fo, Response{Text: "2 1/2"}, "Correct!"},
		{fo, Response{Text: "2/5"}, "You flipped it"},
		{fo, Response{Text: "3"}, "what's 24 × 3? Is that 60?"},
		{fo, Response{Text: ""}, "Type your answer"},
	}
	sugar := Line{Ing: "sugar", Vol: c(3, 4)}
	s, _ := scaledLine(sugar, frac.New(1, 2))
	lo := lineOrder(KScale, sugar, s, frac.New(1, 2), "½")
	cases = append(cases,
		struct {
			o    Order
			resp Response
			want string
		}{lo, Response{Tools: []string{"c14", "T1", "T1"}}, "Correct, and the fewest"},
		struct {
			o    Order
			resp Response
			want string
		}{lo, Response{Tools: []string{"T1", "T1", "T1", "T1", "T1", "T1"}}, "A chef would do it in 3"},
		struct {
			o    Order
			resp Response
			want string
		}{lo, Response{Tools: []string{"c12", "c14"}}, "original amount"},
		struct {
			o    Order
			resp Response
			want string
		}{lo, Response{Tools: []string{"c1", "c12"}}, "wrong way"},
		struct {
			o    Order
			resp Response
			want string
		}{lo, Response{Tools: []string{"c14"}}, "Is that really ¾ c × ½?"},
		struct {
			o    Order
			resp Response
			want string
		}{lo, Response{}, "bowl is empty"},
	)
	for _, c := range cases {
		got := Check(c.o, c.resp)
		if !strings.Contains(got.Message, c.want) {
			t.Errorf("%s with %+v: %q, want %q", c.o.Prompt, c.resp, got.Message, c.want)
		}
	}
	// 6 T = ¼ c + 2 T is 3 scoops; ⅓ c + ... can't beat it.
	if lo.Answer.BestScoops != 3 || lo.Answer.Text != "¼ c + 2 T" {
		t.Errorf("¾ c × ½ best = %q (%d)", lo.Answer.Text, lo.Answer.BestScoops)
	}
}
