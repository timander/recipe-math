package kitchen

// Visual asks the page to draw a picture for an order: the cup wall (one
// cup split into ½, ⅓, ¼, 2 T, and 1 T bars) or butter sticks with their
// tablespoon marks. Amounts are eighth-teaspoons.
type Visual struct {
	Kind   string `json:"kind"`   // "wall" or "butter"
	Given  int    `json:"given"`  // shaded in the last hint: the starting amount
	Answer int    `json:"answer"` // shaded after answering (0 = no picture)
}

// pictureHint is the last hint rung on orders that have a picture.
const pictureHint = "Picture it: how many blocks fit under the shaded part?"

// wallVisual returns a cup-wall picture when both amounts fit in one cup and
// tablespoons are involved; otherwise nil.
func wallVisual(given, answer int) *Visual {
	if given <= 0 || given > Cup || given%Tbsp != 0 {
		return nil
	}
	if answer > Cup || answer%Tbsp != 0 {
		answer = 0
	}
	return &Visual{Kind: "wall", Given: given, Answer: answer}
}

// butterVisual returns butter sticks (up to 4) with tablespoon marks.
func butterVisual(given, answer int) *Visual {
	if given <= 0 || given > 4*Stick || given%Tbsp != 0 {
		return nil
	}
	if answer > 4*Stick || answer%Tbsp != 0 {
		answer = 0
	}
	return &Visual{Kind: "butter", Given: given, Answer: answer}
}

// withVisual attaches a picture and adds the picture hint as the last rung.
func withVisual(o Order, v *Visual) Order {
	if v == nil {
		return o
	}
	o.Visual = v
	o.Hints = append(o.Hints, pictureHint)
	return o
}
