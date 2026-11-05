package kitchen

// Reference is everything the cheat sheet prints.
type Reference struct {
	Facts   []string    `json:"facts"`
	Abbrevs [][2]string `json:"abbrevs"` // {"T, Tbsp, tbsp", "tablespoon"}
	Tools   [][2]string `json:"tools"`   // {"Dry measuring cups", "flour, sugar…"}
	Drawer  []Tool      `json:"drawer"`
	Tips    []string    `json:"tips"`
}

// Ref builds the cheat-sheet content from the same data the game uses.
func Ref() Reference {
	r := Reference{
		Abbrevs: [][2]string{
			{"t, tsp", "teaspoon (small spoon)"}, {"T, Tbsp, Tbs, tbsp", "tablespoon (big spoon) = 3 t"},
			{"c, C", "cup"}, {"fl oz", "fluid ounce: space, read on a liquid cup"}, {"oz", "ounce: weight, on a scale"},
			{"lb, lbs, #", "pound = 16 oz"}, {"pt", "pint = 2 c"}, {"qt", "quart = 4 c"}, {"gal", "gallon = 4 qt"},
			{"g", "gram (metric weight)"}, {"mL", "milliliter (metric volume)"}, {"pinch, dash", "less than ⅛ t"},
		},
		Tools: [][2]string{
			{"Dry measuring cups", "flour, sugar, oats, chips: spoon in, level off. Brown sugar: pack it."},
			{"Liquid measuring cup", "milk, water, oil, broth: on the counter, read at eye level."},
			{"Measuring spoons", "anything less than ¼ c: salt, baking soda, spices, vanilla."},
			{"Kitchen scale", "anything in oz or lb (not fl oz): cheese, meat, pasta."},
			{"Butter wrapper", "1 stick = ½ c = 8 T = 4 oz. Cut at the T marks."},
		},
		Drawer: append(append(dryCupsCopy(), spoons...), butterTools...),
		Tips: []string{
			"Capital T is the big spoon. Mixing up T and t means 3× too much or too little.",
			"Conversion factor = what you want ÷ what the recipe makes. Multiply every line by it.",
			"Fractions of cups get messy: change to T first (16 T in a cup), multiply, then pick tools.",
			"Fewest scoops = fewest chances to spill. 12 T is just ½ c + ¼ c.",
			"Eggs can't be split cleanly: round, or beat one and use half (about 1½ T).",
			"Bake time depends on the size of each piece, not the size of the batch.",
			"Salt and spices: when scaling up, add a little less, taste, then adjust.",
		},
	}
	for _, f := range facts {
		r.Facts = append(r.Facts, f.Explain)
	}
	return r
}

func dryCupsCopy() []Tool { return append([]Tool(nil), dryCups...) }
