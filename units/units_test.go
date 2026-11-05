package units

import (
	"testing"

	"github.com/timander/recipe-math/frac"
)

func TestIDsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, u := range All {
		if seen[u.ID] {
			t.Errorf("duplicate unit %s", u.ID)
		}
		seen[u.ID] = true
		if u.Name == "" || u.Plural == "" || u.Abbr == "" || u.AbbrPl == "" || u.Tool == "" || len(u.dens) == 0 || u.max < 1 {
			t.Errorf("unit %s is missing a field", u.ID)
		}
	}
}

// Every exact fact must agree with the unit sizes, so the cheat sheet and
// the conversions can never disagree.
func TestFactsMatchUnitSizes(t *testing.T) {
	for _, f := range Facts {
		big, ok1 := Get(f.Big)
		small, ok2 := Get(f.Small)
		if !ok1 || !ok2 {
			t.Fatalf("fact %v uses an unknown unit", f)
		}
		if big.Dimension != small.Dimension {
			t.Errorf("%v mixes volume and weight", f)
		}
		if f.Approx != (big.System != small.System) {
			t.Errorf("%v: approx should be true exactly when it crosses systems", f)
		}
		got, approx, err := Convert(frac.Int(1), big, small)
		if err != nil || !got.Equal(f.Factor) || approx != f.Approx {
			t.Errorf("Convert(1 %s → %s) = %v, %v; fact says %v", f.Big, f.Small, got, err, f.Factor)
		}
	}
}

func TestClassicEquivalents(t *testing.T) {
	cases := []struct {
		amount   frac.Frac
		from, to string
		want     frac.Frac
	}{
		{frac.Int(1), "cup", "tsp", frac.Int(48)},
		{frac.New(1, 4), "cup", "Tbsp", frac.Int(4)},
		{frac.New(1, 3), "cup", "Tbsp", frac.New(16, 3)},
		{frac.Int(1), "gallon", "cup", frac.Int(16)},
		{frac.Int(1), "gallon", "floz", frac.Int(128)},
		{frac.Int(1), "quart", "floz", frac.Int(32)},
		{frac.Int(2), "pint", "quart", frac.Int(1)},
		{frac.Int(8), "oz", "lb", frac.New(1, 2)},
		{frac.Int(750), "mL", "L", frac.New(3, 4)},
		{frac.Int(2), "kg", "lb", frac.New(22, 5)},
		{frac.New(1, 2), "cup", "mL", frac.Int(120)},
	}
	for _, c := range cases {
		got, _, err := Convert(c.amount, MustGet(c.from), MustGet(c.to))
		if err != nil || !got.Equal(c.want) {
			t.Errorf("%v %s → %s = %v (%v), want %v", c.amount, c.from, c.to, got, err, c.want)
		}
	}
}

func TestConvertErrors(t *testing.T) {
	if _, _, err := Convert(frac.Int(1), MustGet("cup"), MustGet("oz")); err == nil {
		t.Error("cups → oz (volume → weight) should fail")
	}
	if _, _, err := Convert(frac.Int(1), MustGet("gallon"), MustGet("mL")); err == nil {
		t.Error("gallon → mL has no single kitchen fact and should fail")
	}
}

func TestRoute(t *testing.T) {
	cases := map[[2]string]int{
		{"cup", "tsp"}: 2, {"tsp", "cup"}: 2, {"cup", "gallon"}: 2, {"Tbsp", "tsp"}: 1,
		{"floz", "tsp"}: 2, {"pint", "Tbsp"}: 2, {"cup", "mL"}: 1, {"kg", "lb"}: 1, {"cup", "cup"}: 0,
	}
	for pair, want := range cases {
		from, to := MustGet(pair[0]), MustGet(pair[1])
		route, err := Route(from, to)
		if err != nil || len(route) != want {
			t.Errorf("Route(%s→%s) = %d steps (%v), want %d", pair[0], pair[1], len(route), err, want)
			continue
		}
		// Following the route must give the same answer as converting directly.
		v := frac.Int(1)
		for _, s := range route {
			v, _, _ = Convert(v, s.From, s.To)
		}
		direct, _, _ := Convert(frac.Int(1), from, to)
		if !v.Equal(direct) {
			t.Errorf("Route(%s→%s) gives %v, direct gives %v", pair[0], pair[1], v, direct)
		}
	}
	// cup → tsp should go through Tbsp (the facts students know), not fl oz.
	route, _ := Route(MustGet("cup"), MustGet("tsp"))
	if route[0].To.ID != "Tbsp" {
		t.Errorf("cup → tsp routed via %s, want Tbsp", route[0].To.ID)
	}
	if _, err := Route(MustGet("gallon"), MustGet("mL")); err == nil {
		t.Error("gallon → mL should have no route")
	}
}

func TestFormat(t *testing.T) {
	cases := []struct {
		amount frac.Frac
		unit   string
		want   string
	}{
		{frac.Int(1), "cup", "1 cup"}, {frac.New(3, 4), "cup", "3/4 cup"}, {frac.New(3, 2), "cup", "1 1/2 cups"},
		{frac.New(3, 4), "L", "0.75 L"}, {frac.New(22, 5), "lb", "4.4 lb"}, {frac.Int(12), "Tbsp", "12 Tbsp"},
		{frac.New(1, 2), "lb", "1/2 lb"}, {frac.Int(1600), "mL", "1600 mL"}, {frac.Int(1), "floz", "1 fl oz"},
	}
	for _, c := range cases {
		if got := Format(c.amount, MustGet(c.unit)); got != c.want {
			t.Errorf("Format(%v %s) = %q, want %q", c.amount, c.unit, got, c.want)
		}
	}
	f := Fact{Big: "cup", Small: "mL", Factor: frac.Int(240), Approx: true}
	if f.String() != "1 cup ≈ 240 mL" {
		t.Errorf("fact reads %q", f.String())
	}
}
