package kitchen

import "github.com/timander/recipe-math/frac"

// IngKind decides how an ingredient is measured.
type IngKind string

const (
	Dry    IngKind = "dry"    // flour, sugar: spoon into dry cups, level off
	Wet    IngKind = "liquid" // milk, oil: liquid cup at eye level
	Fat    IngKind = "butter" // sticks and wrapper marks
	Weight IngKind = "weight" // by the scale (or the package label)
	Count  IngKind = "count"  // eggs
)

// Ingredient is one thing in a recipe.
type Ingredient struct {
	Name   string  `json:"name"`
	Plural string  `json:"plural,omitempty"` // for counted things: "eggs"
	Kind   IngKind `json:"kind"`
	Pounds bool    `json:"pounds,omitempty"` // weight written in lb, not oz
	// Fancy is how lab sheets and websites write it: "butter, softened".
	Fancy string `json:"-"`
}

var ing = map[string]Ingredient{
	"flour":       {Name: "all-purpose flour", Kind: Dry},
	"sugar":       {Name: "sugar", Kind: Dry},
	"brownsugar":  {Name: "packed brown sugar", Kind: Dry, Fancy: "firmly packed brown sugar"},
	"cocoa":       {Name: "cocoa powder", Kind: Dry},
	"oats":        {Name: "rolled oats", Kind: Dry},
	"chips":       {Name: "chocolate chips", Kind: Dry},
	"blueberries": {Name: "blueberries", Kind: Dry},
	"onion":       {Name: "chopped onion", Kind: Dry},
	"soda":        {Name: "baking soda", Kind: Dry},
	"powder":      {Name: "baking powder", Kind: Dry},
	"salt":        {Name: "salt", Kind: Dry},
	"pepper":      {Name: "black pepper", Kind: Dry},
	"cinnamon":    {Name: "cinnamon", Kind: Dry},
	"chili":       {Name: "chili powder", Kind: Dry},
	"cumin":       {Name: "cumin", Kind: Dry},
	"milk":        {Name: "milk", Kind: Wet},
	"water":       {Name: "water", Kind: Wet},
	"oil":         {Name: "vegetable oil", Kind: Wet},
	"vinegar":     {Name: "vinegar", Kind: Wet},
	"vanilla":     {Name: "vanilla extract", Kind: Wet, Fancy: "pure vanilla extract"},
	"sauce":       {Name: "tomato sauce", Kind: Wet},
	"broth":       {Name: "chicken broth", Kind: Wet},
	"honey":       {Name: "honey", Kind: Wet},
	"mustard":     {Name: "Dijon mustard", Kind: Wet},
	"butter":      {Name: "butter", Kind: Fat, Fancy: "butter, softened"},
	"pasta":       {Name: "elbow macaroni", Kind: Weight},
	"cheese":      {Name: "shredded cheddar", Kind: Weight},
	"beef":        {Name: "ground beef", Kind: Weight, Pounds: true},
	"chipsoz":     {Name: "chocolate chips", Kind: Weight},
	"creamcheese": {Name: "cream cheese", Kind: Weight},
	"egg":         {Name: "egg", Plural: "eggs", Kind: Count, Fancy: "large"},
}

// Line is one ingredient line on a recipe card. Exactly one of Vol, Oz,
// or N is set, depending on the ingredient's kind.
type Line struct {
	Ing string
	Vol int       // eighth-teaspoons
	Oz  frac.Frac // ounces
	N   frac.Frac // count
}

func v(id string, e int) Line               { return Line{Ing: id, Vol: e} }
func w(id string, oz int64) Line            { return Line{Ing: id, Oz: frac.Int(oz)} }
func n(id string, count int64) Line         { return Line{Ing: id, N: frac.Int(count)} }
func c(num, den int) int                    { return Cup * num / den } // c(3,4) = ¾ cup
func tb(k int) int                          { return k * Tbsp }
func ts(num, den int) int                   { return Tsp * num / den } // ts(1,2) = ½ tsp
func recipe(r Recipe, lines ...Line) Recipe { r.Lines = lines; return r }

// Recipe is a real, classroom-sized recipe.
type Recipe struct {
	Name      string
	Makes     int
	MakesWhat string // "cookies", "servings"
	Lines     []Line
}

// Recipes are typical high-school foods-lab recipes with standard amounts.
var Recipes = []Recipe{
	recipe(Recipe{Name: "Chocolate Chip Cookies", Makes: 24, MakesWhat: "cookies"},
		v("flour", c(9, 4)), v("soda", ts(1, 1)), v("salt", ts(1, 2)), v("butter", c(1, 1)),
		v("sugar", c(3, 4)), v("brownsugar", c(3, 4)), v("vanilla", ts(1, 1)), n("egg", 2), v("chips", c(2, 1))),
	recipe(Recipe{Name: "Pancakes", Makes: 8, MakesWhat: "pancakes"},
		v("flour", c(1, 1)), v("sugar", tb(2)), v("powder", ts(2, 1)), v("salt", ts(1, 4)),
		v("milk", c(1, 1)), n("egg", 1), v("oil", tb(2))),
	recipe(Recipe{Name: "Blueberry Muffins", Makes: 12, MakesWhat: "muffins"},
		v("flour", c(3, 2)), v("sugar", c(3, 4)), v("powder", ts(2, 1)), v("salt", ts(1, 2)),
		v("oil", c(1, 3)), n("egg", 1), v("milk", c(1, 3)), v("blueberries", c(1, 1))),
	recipe(Recipe{Name: "Mac & Cheese", Makes: 4, MakesWhat: "servings"},
		w("pasta", 8), v("butter", tb(2)), v("flour", tb(2)), v("milk", c(2, 1)),
		w("cheese", 8), v("salt", ts(1, 2)), v("pepper", ts(1, 4))),
	recipe(Recipe{Name: "Brownies", Makes: 16, MakesWhat: "brownies"},
		v("butter", c(1, 2)), v("sugar", c(1, 1)), n("egg", 2), v("vanilla", ts(1, 1)),
		v("cocoa", c(1, 3)), v("flour", c(1, 2)), v("salt", ts(1, 4)), v("powder", ts(1, 4))),
	recipe(Recipe{Name: "Chili", Makes: 6, MakesWhat: "servings"},
		w("beef", 16), v("onion", c(1, 1)), v("sauce", c(2, 1)), v("chili", tb(2)),
		v("cumin", ts(1, 1)), v("salt", ts(1, 2))),
	recipe(Recipe{Name: "Oatmeal Cookies", Makes: 36, MakesWhat: "cookies"},
		v("butter", c(1, 1)), v("brownsugar", c(1, 1)), v("sugar", c(1, 2)), n("egg", 2),
		v("vanilla", ts(1, 1)), v("flour", c(3, 2)), v("soda", ts(1, 1)), v("cinnamon", ts(1, 1)),
		v("salt", ts(1, 2)), v("oats", c(3, 1))),
	recipe(Recipe{Name: "House Vinaigrette", Makes: 6, MakesWhat: "servings"},
		v("oil", c(1, 2)), v("vinegar", c(1, 4)), v("mustard", ts(1, 1)), v("honey", tb(1)),
		v("salt", ts(1, 4)), v("pepper", ts(1, 8))),
	recipe(Recipe{Name: "Biscuits", Makes: 12, MakesWhat: "biscuits"},
		v("flour", c(2, 1)), v("powder", tb(1)), v("salt", ts(1, 1)), v("butter", tb(6)), v("milk", c(3, 4))),
}

// ---------- Read the Recipe (vocabulary) ----------

type vocabItem struct {
	Q       string
	Right   string
	Wrong   []string
	Hints   []string
	Explain string
}

var vocab = []vocabItem{
	{"A recipe says “2 T butter.” What does T mean?", "tablespoon", []string{"teaspoon", "ton", "tub"},
		[]string{"Capital T or lowercase t: which one is the bigger letter?", "Which spoon is bigger: a tablespoon or a teaspoon?"},
		"Capital T = tablespoon (the bigger spoon). Lowercase t = teaspoon."},
	{"A recipe says “1 t vanilla.” What does t mean?", "teaspoon", []string{"tablespoon", "tub", "taste"},
		[]string{"Is this t a capital or a lowercase letter?", "Lowercase letter, smaller spoon. Which spoon is smaller?"},
		"Lowercase t = teaspoon. Mixing up t and T puts in 3 times too much (or too little)."},
	{"Which abbreviation means tablespoon?", "Tbsp", []string{"tsp", "t", "Tsp"},
		[]string{"Which choice has the letter b for “table”?", "tsp is the teaspoon. Which one is left with a b in it?"},
		"Tablespoon: T, Tbsp, Tbs, or tbsp. Teaspoon: t or tsp."},
	{"Which abbreviation means teaspoon?", "tsp", []string{"Tbsp", "T", "tbs"},
		[]string{"Does “teaspoon” have a b in it?", "Which choice has no b and starts with a lowercase t?"},
		"Teaspoon: t or tsp. Tablespoon: T, Tbsp, Tbs, or tbsp."},
	{"A recipe says “1½ c flour.” What does c mean?", "cup", []string{"can", "container", "carton"},
		[]string{"What tool do you scoop flour with?", "Which choice is a measuring tool?"},
		"c (or C) = cup."},
	{"“8 oz cheddar” is measured by…", "weight, on a scale", []string{"space, in a liquid cup", "counting pieces", "the size of the bag"},
		[]string{"Does “oz” say fluid ounces or plain ounces?", "Plain ounces measure how heavy something is. What tool measures heavy?"},
		"oz = ounces of weight (use a scale). fl oz = fluid ounces of space (use a liquid cup)."},
	{"“8 fl oz milk” is measured by…", "space, in a liquid cup", []string{"weight, on a scale", "counting cartons", "a dry measuring cup"},
		[]string{"What does the “fl” in fl oz stand for?", "Fluid ounces measure how much space a liquid fills. Which tool shows that?"},
		"fl oz = fluid ounces: space. 8 fl oz = 1 cup in a liquid measuring cup."},
	{"What does lb mean?", "pound", []string{"liter", "large bag", "level"},
		[]string{"Is lb a weight or a volume?", "16 oz make one of these. Which one?"},
		"lb = pound (from the Latin libra). 1 lb = 16 oz."},
	{"A recipe says “a pinch of salt.” About how much is that?", "less than ⅛ t", []string{"1 t", "1 T", "¼ c"},
		[]string{"How much salt can you pick up between your thumb and one finger?", "Is that more or less than the smallest spoon in the drawer?"},
		"A pinch is what you can pinch between two fingers: less than ⅛ t."},
	{"“1 c packed brown sugar” means…", "press it into the cup until it holds its shape", []string{"pour it in loosely", "use a liquid cup", "weigh it in its package"},
		[]string{"What does it mean to pack a suitcase full?", "If you dumped the cup out, should it hold the cup's shape?"},
		"Packed = press it firmly into a dry cup. It should hold the cup's shape when you tip it out."},
	{"How do you measure 1 c of flour?", "spoon it into a dry cup, then level it with a straight edge", []string{"scoop the cup into the bag and shake it", "pour it into a liquid cup", "pack it down hard"},
		[]string{"What happens to flour if you push a cup into the bag?", "Scooping packs flour down. How could you fill the cup without packing it?"},
		"Spoon flour in and level it off. Scooping packs it down and can add 25% more flour: dry, tough cookies."},
	{"How do you read a liquid measuring cup?", "set it on the counter and read the line at eye level", []string{"hold it up high and look from below", "look down from above", "fill it to the very top"},
		[]string{"Where should the cup sit so it isn't tilted?", "Where should your eyes be compared to the line?"},
		"Flat counter, eyes level with the line. Looking from above makes you under-fill."},
	{"What does gal mean?", "gallon", []string{"gram", "galley", "glass"},
		[]string{"What size jug does milk come in at the store?", "4 quarts make one of these. Which one?"},
		"gal = gallon = 4 quarts = 16 cups."},
	{"What does qt mean?", "quart", []string{"quarter cup", "quantity", "quick"},
		[]string{"Is it a tiny amount, or bigger than a cup?", "4 cups make one of these. Which one?"},
		"qt = quart = 4 cups (a quarter of a gallon)."},
	{"What does pt mean?", "pint", []string{"part", "pinch", "portion"},
		[]string{"What size container does a small tub of ice cream come in?", "2 cups make one of these. Which one?"},
		"pt = pint = 2 cups."},
	{"What does g mean on a nutrition label?", "gram", []string{"gallon", "glass", "grain"},
		[]string{"Is it metric?", "Is it a tiny weight or a big volume?"},
		"g = gram, a small metric weight. 1 oz ≈ 28 g."},
	{"What does mL mean on a measuring cup?", "milliliter", []string{"mega liter", "milk level", "mixing line"},
		[]string{"Is it metric?", "1 t ≈ 5 of these. Big unit or tiny one?"},
		"mL = milliliter, a small metric volume. 1 c ≈ 240 mL."},
}

// ---------- Walk-in Facts (core equivalents to memorize) ----------

type fact struct {
	Q       string
	Answer  int64
	Unit    string
	Hints   []string
	Explain string
}

// factPics gives some facts a picture: the amount they're about and which
// drawing to use.
var factPics = map[string]struct {
	vol int
	pic string
}{
	"How many T are in ¼ c?":               {QuarterCup, "wall"},
	"How many T are in ½ c?":               {HalfCup, "wall"},
	"How many T are in 1 c?":               {Cup, "wall"},
	"How many T are in ¾ c?":               {3 * QuarterCup, "wall"},
	"How many T are in 1 stick of butter?": {Stick, "butter"},
	"How many sticks of butter make 1 c?":  {Cup, "butter"},
}

var facts = []fact{
	{"How many t are in 1 T?", 3, "t", []string{"Is a T bigger or smaller than a t?", "How many t scoops would fill one T spoon?"}, "3 t = 1 T"},
	{"How many T are in ¼ c?", 4, "T", []string{"How many T are in a whole cup?", "A ¼ cup is 1 of 4 equal parts of a cup. How do you split 16 T into 4 equal parts?"}, "4 T = ¼ c"},
	{"How many T are in ½ c?", 8, "T", []string{"How many T are in ¼ c?", "½ c is two ¼ cups. What's two of those?"}, "8 T = ½ c"},
	{"How many T are in 1 c?", 16, "T", []string{"How many T are in ¼ c?", "1 c is four ¼ cups. What's four of those?"}, "16 T = 1 c"},
	{"How many T are in ¾ c?", 12, "T", []string{"How many T are in ¼ c?", "¾ c is three ¼ cups. What's three of those?"}, "12 T = ¾ c"},
	{"How many fl oz are in 1 c?", 8, "fl oz", []string{"Look at the side of a liquid cup. What's printed at the 1 c line?", "1 fl oz = 2 T, and 1 c = 16 T. How many 2s fit in 16?"}, "8 fl oz = 1 c"},
	{"How many T are in 1 fl oz?", 2, "T", []string{"How many fl oz are in a cup, and how many T?", "8 fl oz and 16 T are the same cup. How many T for each fl oz?"}, "2 T = 1 fl oz"},
	{"How many c are in 1 pt?", 2, "c", []string{"A pint of ice cream: is it more or less than 1 c?", "Think “a pint's a pound the world around”: 16 fl oz. How many 8 fl oz cups is that?"}, "2 c = 1 pt"},
	{"How many c are in 1 qt?", 4, "c", []string{"“Quart” sounds like “quarter.” A quarter of what?", "A quart is a quarter of a gallon (16 c). What's 16 split into 4?"}, "4 c = 1 qt"},
	{"How many qt are in 1 gal?", 4, "qt", []string{"What does “quart” sound like?", "A quart is a quarter of a gallon. How many quarters make a whole?"}, "4 qt = 1 gal"},
	{"How many c are in 1 gal?", 16, "c", []string{"How many qt are in a gallon?", "Each quart is 4 c. How many cups in 4 quarts?"}, "16 c = 1 gal"},
	{"How many oz are in 1 lb?", 16, "oz", []string{"Look at a 1 lb block of cheese. What does the oz part of the label say?", "A ½ lb package is 8 oz. What's double that?"}, "16 oz = 1 lb"},
	{"How many T are in 1 stick of butter?", 8, "T", []string{"Look at a butter wrapper. How many lines are marked on it?", "1 stick = ½ c. How many T are in ½ c?"}, "1 stick = 8 T = ½ c = 4 oz"},
	{"How many sticks of butter make 1 c?", 2, "sticks", []string{"How many cups is 1 stick?", "1 stick = ½ c. How many halves make a whole?"}, "2 sticks = 1 c"},
	{"How many oz does 1 stick of butter weigh?", 4, "oz", []string{"A box of butter has 4 sticks and weighs 1 lb. How many oz is 1 lb?", "16 oz shared by 4 sticks: how much for each?"}, "1 stick = 4 oz (a 1 lb box has 4 sticks)"},
}

// ---------- Chef's Call (judgment) ----------

type call struct {
	Q       string
	Right   []string // more than one choice can be a good call
	Wrong   []string
	Hints   []string
	Explain string
}

var calls = []call{
	{"You're making half a recipe that calls for 3 eggs. What's the best call?",
		[]string{"Use 1 egg, then beat a 2nd egg and add half of it (about 1½ T)", "Round to 2 eggs"},
		[]string{"Use 3 eggs anyway", "Skip the eggs", "Crack 1½ eggs straight into the bowl"},
		[]string{"Can you crack half an egg cleanly?", "What could you do to an egg so you can measure part of it?"},
		"Half of 3 = 1½ eggs. Beat one egg and use half (about 1½ T), or round to 2 for most recipes."},
	{"You accidentally used 1 T of baking soda instead of 1 t. What happens?",
		[]string{"The cookies taste bitter and soapy: that's 3 times too much"},
		[]string{"Nothing, T and t are the same", "They rise a little higher and taste better", "They just bake faster"},
		[]string{"How many t are in 1 T?", "What happens when you put 3 times too much of a strong ingredient in?"},
		"1 T = 3 t, so that's 3× the baking soda: bitter, soapy, and it can rise then collapse."},
	{"You're doubling a cookie recipe. How long do you bake each pan?",
		[]string{"The same time: each cookie is still the same size"},
		[]string{"Twice as long", "Half as long", "It depends on how many cookies are in the oven, so add 5 min per dozen"},
		[]string{"Does doubling the dough make each cookie bigger?", "What decides how long one cookie takes to bake?"},
		"Bake time depends on the size of each piece, not the size of the batch. Same cookies, same time."},
	{"You're doubling brownies that bake in an 8×8 pan. What's the best call?",
		[]string{"Use two 8×8 pans (or one 9×13) at the same temperature, and check for doneness"},
		[]string{"Use the same 8×8 pan and bake twice as long", "Raise the oven temperature by double", "Use the 8×8 pan and bake the same time"},
		[]string{"If you pour twice the batter in the same pan, how thick does it get?", "What happens to the middle of a very thick brownie?"},
		"Double batter in the same pan is twice as thick: burnt edges, raw middle. Use twice the pan area."},
	{"A soup for 4 uses ½ t salt. You're making it for 16. What's the best call?",
		[]string{"Add about 1½ t, taste, then add more if it needs it"},
		[]string{"Add 2 T right away", "Skip the salt; people can add their own", "Add ½ t, the same as the recipe"},
		[]string{"What's the conversion factor from 4 to 16?", "Salt is easy to add but impossible to take out. Should you add it all at once?"},
		"×4 says 2 t, but cooks add salt a little under, taste, and adjust. You can't take salt out."},
	{"Your recipe calls for 2 c milk, but you only have a dry measuring cup. What's the best call?",
		[]string{"Fill it carefully to the very rim and use it twice"},
		[]string{"Fill it 3 times to be safe", "Guess with a mixing bowl", "Use 2 c of cream instead"},
		[]string{"Do dry and liquid cups hold different amounts, or are they just shaped differently?", "Why is it harder to fill a dry cup with liquid?"},
		"Both hold the same volume. A liquid cup just has a spout and room below the rim so it doesn't spill."},
	{"A recipe calls for 1 c butter. You have sticks. How many?",
		[]string{"2 sticks"},
		[]string{"1 stick", "4 sticks", "8 sticks"},
		[]string{"How many cups is 1 stick?", "How many halves make a whole?"},
		"1 stick = ½ c, so 1 c = 2 sticks."},
	{"Scaling a recipe ×3 gives 4½ eggs. What's the best call?",
		[]string{"Use 4 eggs, then beat a 5th and add half of it", "Round to 5 eggs"},
		[]string{"Round down to 3 eggs", "Use 9 eggs", "Leave the eggs out"},
		[]string{"Can you crack half an egg?", "Is 4½ closer to 4 or 5? Will one extra half-egg ruin it?"},
		"4½ eggs: use 4 plus half a beaten egg, or round to 5. Eggs are forgiving by a half."},
}
