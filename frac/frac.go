// Package frac is a small exact-fraction type for kitchen math.
// Kitchen amounts are small, so int64 numerators and denominators are plenty,
// and exact fractions mean 1/3 × 3 is exactly 1, never 0.999.
package frac

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Frac is a reduced fraction n/d with d > 0. The zero value is 0.
type Frac struct {
	n, d int64
}

// New returns n/d in lowest terms. It panics if d is 0.
func New(n, d int64) Frac {
	if d == 0 {
		panic("frac: zero denominator")
	}
	if d < 0 {
		n, d = -n, -d
	}
	g := gcd(abs(n), d)
	return Frac{n / g, d / g}
}

// Int returns the whole number n as a fraction.
func Int(n int64) Frac { return Frac{n, 1} }

// Num returns the numerator.
func (a Frac) Num() int64 { return a.n }

// Den returns the denominator (always ≥ 1).
func (a Frac) Den() int64 {
	if a.d == 0 {
		return 1
	}
	return a.d
}

func (a Frac) Add(b Frac) Frac { return New(a.n*b.Den()+b.n*a.Den(), a.Den()*b.Den()) }
func (a Frac) Sub(b Frac) Frac { return New(a.n*b.Den()-b.n*a.Den(), a.Den()*b.Den()) }

func (a Frac) Mul(b Frac) Frac {
	// Cross-reduce first to keep the numbers small.
	g1 := gcd(abs(a.n), b.Den())
	g2 := gcd(abs(b.n), a.Den())
	return New((a.n/g1)*(b.n/g2), (a.Den()/g2)*(b.Den()/g1))
}

// Div returns a ÷ b. It panics if b is 0.
func (a Frac) Div(b Frac) Frac {
	if b.n == 0 {
		panic("frac: division by zero")
	}
	return a.Mul(New(b.Den(), b.n))
}

// Cmp returns -1, 0, or +1.
func (a Frac) Cmp(b Frac) int {
	l, r := a.n*b.Den(), b.n*a.Den()
	switch {
	case l < r:
		return -1
	case l > r:
		return 1
	}
	return 0
}

func (a Frac) Equal(b Frac) bool { return a.Cmp(b) == 0 }
func (a Frac) IsInt() bool       { return a.Den() == 1 }
func (a Frac) Float64() float64  { return float64(a.n) / float64(a.Den()) }

// String formats as a whole or mixed number: "12", "3/4", "1 1/2".
func (a Frac) String() string {
	n, d := a.n, a.Den()
	sign := ""
	if n < 0 {
		sign, n = "-", -n
	}
	whole, rem := n/d, n%d
	switch {
	case rem == 0:
		return sign + strconv.FormatInt(whole, 10)
	case whole == 0:
		return fmt.Sprintf("%s%d/%d", sign, rem, d)
	}
	return fmt.Sprintf("%s%d %d/%d", sign, whole, rem, d)
}

// Decimal formats as an exact decimal ("0.75", "2.5") and reports whether
// that was possible (only when the denominator has no factors but 2 and 5).
func (a Frac) Decimal() (string, bool) {
	d := a.Den()
	places := 0
	for d%10 == 0 || d%2 == 0 || d%5 == 0 {
		switch {
		case d%10 == 0:
			d /= 10
		case d%2 == 0:
			d /= 2
		default:
			d /= 5
		}
		places++
	}
	if d != 1 {
		return "", false
	}
	// Enough places for any 2^a·5^b denominator; trailing zeros trimmed.
	s := strconv.FormatFloat(a.Float64(), 'f', places, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s, true
}

// Unicode fraction characters students might paste or type on a phone.
var vulgar = strings.NewReplacer(
	"½", " 1/2", "⅓", " 1/3", "⅔", " 2/3", "¼", " 1/4", "¾", " 3/4",
	"⅛", " 1/8", "⅜", " 3/8", "⅝", " 5/8", "⅞", " 7/8", "⁄", "/",
)

// ErrFormat is returned when the text isn't a number Parse understands.
var ErrFormat = errors.New(`type a number like 12, 3/4, 1 1/2, or 1.5`)

// Parse reads a whole number, fraction, mixed number, or decimal:
// "12", "3/4", "1 1/2", "1-1/2", "1½", "1.5", ".75". Trailing words such as
// a unit ("12 Tbsp") are ignored. Negative amounts are rejected.
func Parse(s string) (Frac, error) {
	s = vulgar.Replace(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ",", "")
	// Keep only the leading numeric part, so "12 Tbsp" reads as "12".
	end := strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune("0123456789./- ", r)
	})
	if end >= 0 {
		s = s[:end]
	}
	s = strings.TrimSpace(s)
	// "1-1/2" is a mixed number, not subtraction.
	if i := strings.Index(s, "-"); i > 0 && strings.Contains(s[i:], "/") {
		s = s[:i] + " " + s[i+1:]
	}
	fields := strings.Fields(s)
	switch len(fields) {
	case 1:
		if strings.Contains(fields[0], "/") {
			return parseFraction(fields[0])
		}
		return parseDecimal(fields[0])
	case 2:
		whole, err := parseWhole(fields[0])
		if err != nil {
			return Frac{}, err
		}
		part, err := parseFraction(fields[1])
		if err != nil {
			return Frac{}, err
		}
		return whole.Add(part), nil
	}
	return Frac{}, ErrFormat
}

func parseWhole(s string) (Frac, error) {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil || n < 0 {
		return Frac{}, ErrFormat
	}
	return Int(n), nil
}

func parseFraction(s string) (Frac, error) {
	top, bottom, ok := strings.Cut(s, "/")
	if !ok {
		return Frac{}, ErrFormat
	}
	n, err1 := strconv.ParseInt(top, 10, 32)
	d, err2 := strconv.ParseInt(bottom, 10, 32)
	if err1 != nil || err2 != nil || n < 0 || d <= 0 {
		return Frac{}, ErrFormat
	}
	return New(n, d), nil
}

func parseDecimal(s string) (Frac, error) {
	whole, dec, hasDot := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	if !hasDot {
		return parseWhole(whole)
	}
	if len(dec) == 0 || len(dec) > 6 || strings.ContainsAny(dec, "-./") {
		if len(dec) == 0 {
			return parseWhole(whole)
		}
		return Frac{}, ErrFormat
	}
	w, err := parseWhole(whole)
	if err != nil {
		return Frac{}, err
	}
	n, err := strconv.ParseInt(dec, 10, 32)
	if err != nil {
		return Frac{}, ErrFormat
	}
	scale := int64(1)
	for range dec {
		scale *= 10
	}
	return w.Add(New(n, scale)), nil
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	if a == 0 {
		return 1
	}
	return a
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
