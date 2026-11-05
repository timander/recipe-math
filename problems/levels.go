// Package problems builds unit-conversion problems at six levels, explains
// how to solve them, and checks students' answers.
package problems

// Pair is two units that a level converts between, in both directions.
type Pair struct{ A, B string }

// Level is one step in the progression, from easiest to hardest.
type Level struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Goal    string `json:"goal"`    // what students learn, in plain words
	Example string `json:"example"` // a sample problem for the level card
	Pairs   []Pair `json:"-"`
	// Dens limits which fractions may appear (1 = whole numbers only). A
	// value must also use a denominator that makes sense for its unit.
	Dens []int64 `json:"-"`
	// AnswerDens, if set, limits answers more tightly than Dens.
	AnswerDens []int64 `json:"-"`
	// Max, if set, lowers a unit's largest amount for this level.
	Max map[string]int64 `json:"-"`
}

// max returns the largest amount of a unit allowed at this level.
func (l Level) max(id string, unitMax int64) int64 {
	if m, ok := l.Max[id]; ok {
		return min(m, unitMax)
	}
	return unitMax
}

func (l Level) answerDens() []int64 {
	if l.AnswerDens != nil {
		return l.AnswerDens
	}
	return l.Dens
}

// Levels is the full progression.
var Levels = []Level{
	{
		ID: 1, Name: "Spoons", Goal: "Teaspoons and tablespoons. Whole numbers only.",
		Example: "2 Tbsp = ___ tsp",
		Pairs:   []Pair{{"Tbsp", "tsp"}},
		Dens:    []int64{1},
	},
	{
		ID: 2, Name: "Spoons to Cups", Goal: "Tablespoons, fluid ounces, and cups, with 1/4, 1/2, and 3/4 cups.",
		Example: "1/2 cup = ___ Tbsp",
		Pairs:   []Pair{{"cup", "Tbsp"}, {"floz", "Tbsp"}, {"cup", "floz"}},
		Dens:    []int64{1, 2, 4},
	},
	{
		ID: 3, Name: "Big Containers", Goal: "Cups, pints, quarts, and gallons.",
		Example: "1 gallon = ___ quarts",
		Pairs:   []Pair{{"pint", "cup"}, {"quart", "pint"}, {"quart", "cup"}, {"gallon", "quart"}},
		Dens:    []int64{1, 2, 4},
	},
	{
		ID: 4, Name: "Weight & Metric", Goal: "Ounces and pounds, grams and kilograms, milliliters and liters.",
		Example: "1 1/2 lb = ___ oz",
		Pairs:   []Pair{{"lb", "oz"}, {"kg", "g"}, {"L", "mL"}},
		Dens:    []int64{1, 2, 4, 5, 10},
	},
	{
		ID: 5, Name: "Two Steps & Thirds", Goal: "Jumps that take two steps, like teaspoons to cups, plus thirds and eighths.",
		Example: "1/3 cup = ___ Tbsp",
		Pairs: []Pair{
			{"cup", "Tbsp"}, {"cup", "tsp"}, {"floz", "tsp"}, {"pint", "Tbsp"},
			{"quart", "floz"}, {"gallon", "pint"}, {"gallon", "cup"},
		},
		Dens: []int64{1, 2, 3, 4, 8},
	},
	{
		ID: 6, Name: "US ↔ Metric", Goal: "Cups to milliliters, ounces to grams, and back. Answers are close, not exact (≈).",
		Example: "1 cup ≈ ___ mL",
		Pairs: []Pair{
			{"tsp", "mL"}, {"Tbsp", "mL"}, {"floz", "mL"}, {"cup", "mL"},
			{"oz", "g"}, {"lb", "g"}, {"kg", "lb"},
		},
		Dens:       []int64{1, 2, 4, 5, 10},
		AnswerDens: []int64{1, 2, 5, 10},
		// Everyday amounts only: a few spoons or cups, a pound or two.
		Max: map[string]int64{"tsp": 12, "Tbsp": 8, "floz": 16, "cup": 4, "mL": 1000, "oz": 16, "lb": 5, "g": 2500},
	},
}

// LevelByID returns the level with the given ID.
func LevelByID(id int) (Level, bool) {
	for _, l := range Levels {
		if l.ID == id {
			return l, true
		}
	}
	return Level{}, false
}
