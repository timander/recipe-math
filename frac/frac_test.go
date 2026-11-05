package frac

import "testing"

func TestArithmetic(t *testing.T) {
	third := New(1, 3)
	if got := third.Mul(Int(3)); !got.Equal(Int(1)) {
		t.Errorf("1/3 × 3 = %v, want 1", got)
	}
	if got := New(3, 4).Mul(Int(16)); !got.Equal(Int(12)) {
		t.Errorf("3/4 × 16 = %v, want 12", got)
	}
	if got := Int(12).Div(Int(16)); got.String() != "3/4" {
		t.Errorf("12 ÷ 16 = %v, want 3/4", got)
	}
	if got := New(1, 2).Add(New(1, 3)); got.String() != "5/6" {
		t.Errorf("1/2 + 1/3 = %v", got)
	}
	if got := New(1, 2).Sub(New(3, 4)); got.String() != "-1/4" {
		t.Errorf("1/2 - 3/4 = %v", got)
	}
	if New(2, -4).String() != "-1/2" {
		t.Error("negative denominator not normalized")
	}
	var zero Frac
	if zero.Den() != 1 || zero.String() != "0" || !zero.Add(Int(2)).Equal(Int(2)) {
		t.Error("zero value should act like 0")
	}
}

func TestString(t *testing.T) {
	for f, want := range map[Frac]string{
		Int(12): "12", New(3, 4): "3/4", New(3, 2): "1 1/2", New(16, 3): "5 1/3", New(-7, 4): "-1 3/4",
	} {
		if got := f.String(); got != want {
			t.Errorf("String(%d/%d) = %q, want %q", f.n, f.d, got, want)
		}
	}
}

func TestDecimal(t *testing.T) {
	cases := []struct {
		f    Frac
		want string
		ok   bool
	}{
		{New(3, 4), "0.75", true}, {New(11, 5), "2.2", true}, {Int(1600), "1600", true},
		{New(1, 8), "0.125", true}, {New(1, 3), "", false}, {New(9, 10), "0.9", true},
	}
	for _, c := range cases {
		got, ok := c.f.Decimal()
		if got != c.want || ok != c.ok {
			t.Errorf("Decimal(%v) = %q,%v want %q,%v", c.f, got, ok, c.want, c.ok)
		}
	}
}

func TestParse(t *testing.T) {
	good := map[string]Frac{
		"12": Int(12), " 3/4 ": New(3, 4), "1 1/2": New(3, 2), "1-1/2": New(3, 2),
		"1½": New(3, 2), "½": New(1, 2), "1.5": New(3, 2), ".75": New(3, 4), "0.333": New(333, 1000),
		"12 Tbsp": Int(12), "1 1/2 cups": New(3, 2), "1,600": Int(1600), "4/2": Int(2), "5.": Int(5),
		"2 ⅔": New(8, 3), "0": Int(0), "0/4": Int(0), "1.123456": New(1123456, 1000000),
	}
	for in, want := range good {
		got, err := Parse(in)
		if err != nil || !got.Equal(want) {
			t.Errorf("Parse(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, in := range []string{"", "abc", "1/0", "-3", "1 2 3", "1/2/3", "1.2.3", "3/", "/4", "1 1.5", "Tbsp", ".", ". T", "-1/2", "1.1234567"} {
		if got, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) = %v, want an error", in, got)
		}
	}
}

func TestCmpAndIsInt(t *testing.T) {
	if New(1, 3).Cmp(New(1, 2)) != -1 || New(1, 2).Cmp(New(1, 3)) != 1 || New(2, 4).Cmp(New(1, 2)) != 0 {
		t.Error("Cmp")
	}
	if !Int(3).IsInt() || New(3, 2).IsInt() {
		t.Error("IsInt")
	}
	for f, want := range map[Frac]string{New(1, 16): "0.0625", New(1, 20): "0.05", New(3, 40): "0.075", New(7, 1000): "0.007"} {
		if got, ok := f.Decimal(); !ok || got != want {
			t.Errorf("Decimal(%v) = %q, want %q", f, got, want)
		}
	}
}

func TestNumAndDivByZero(t *testing.T) {
	if New(6, 8).Num() != 3 {
		t.Error("Num should be reduced")
	}
	defer func() {
		if recover() == nil {
			t.Error("dividing by zero should panic")
		}
	}()
	Int(1).Div(Int(0))
}

func TestNewZeroDenominatorPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a zero denominator should panic")
		}
	}()
	New(1, 0)
}
