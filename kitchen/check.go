package kitchen

import (
	"fmt"
	"slices"
	"strings"

	"github.com/timander/recipe-math/frac"
)

// Response is what the student served.
type Response struct {
	Choice int      `json:"choice"`
	Tools  []string `json:"tools"`
	Text   string   `json:"text"`
}

// Result is the verdict. A wrong answer's Message is a question that points
// at the mistake without giving the answer away.
type Result struct {
	Correct bool   `json:"correct"`
	Invalid bool   `json:"invalid"` // couldn't read it: no penalty
	Message string `json:"message"`
	// FreeHint means the message already asked the first hint question, so
	// the page can show hint 1 as unlocked.
	FreeHint   bool `json:"freeHint,omitempty"`
	Scoops     int  `json:"scoops,omitempty"`
	BestScoops int  `json:"bestScoops,omitempty"`
}

// Check judges one serve.
func Check(o Order, r Response) Result {
	switch o.Input {
	case InChoice:
		return checkChoice(o, r.Choice)
	case InBuild:
		return checkBuild(o, r.Tools)
	}
	return checkNumber(o, r.Text)
}

func checkChoice(o Order, choice int) Result {
	if choice < 0 || choice >= len(o.Choices) {
		return Result{Invalid: true, Message: "Tap one of the choices."}
	}
	if slices.Contains(o.Answer.Choices, choice) {
		return Result{Correct: true, Message: "Correct!"}
	}
	switch o.Kind {
	case KTool:
		return Result{Message: toolWhyNot(o.Answer.Choices[0], choice)}
	case KFix:
		return Result{Message: "That line checks out. Keep looking: compare each line to the original card."}
	}
	msg := "Not quite."
	if len(o.Hints) > 0 {
		msg += " " + o.Hints[0]
	}
	return Result{Message: msg, FreeHint: len(o.Hints) > 0}
}

func checkBuild(o Order, ids []string) Result {
	if len(ids) == 0 {
		return Result{Invalid: true, Message: "The bowl is empty. Tap tools to add scoops."}
	}
	sum, tools, err := Total(ids)
	if err != nil {
		return Result{Invalid: true, Message: "You can only fill the liquid cup to one line. Take one out."}
	}
	want := o.Answer.Vol
	if sum == want {
		res := Result{Correct: true, Scoops: len(tools), BestScoops: o.Answer.BestScoops}
		if len(tools) > o.Answer.BestScoops {
			res.Message = fmt.Sprintf("Correct, that measures it! A chef would do it in %d: %s.", o.Answer.BestScoops, o.Answer.Text)
		} else {
			res.Message = "Correct, and the fewest scoops possible!"
		}
		return res
	}
	got := Text(sum, Short)
	given, f := o.Answer.Given[0], unpair(o.Answer.Factor)
	if given > 0 { // a scaling order: catch the two classic slips
		switch {
		case int64(sum) == given:
			return Result{Message: fmt.Sprintf("That's the recipe's original amount. Did you multiply by %s?", Mixed(f))}
		case frac.Int(given).Div(f).Equal(frac.Int(int64(sum))):
			return Result{Message: fmt.Sprintf("You went the wrong way. With ×%s, should the amount get bigger or smaller?", Mixed(f))}
		}
		return Result{Message: fmt.Sprintf("Your scoops add up to %s. Is that really %s × %s?", got, Text(int(given), Short), Mixed(f))}
	}
	if sum > want {
		return Result{Message: fmt.Sprintf("Your scoops add up to %s. Is that more or less than the order? Which scoop could come out?", got)}
	}
	return Result{Message: fmt.Sprintf("Your scoops add up to %s. Is that more or less than the order? What's still missing?", got)}
}

func checkNumber(o Order, text string) Result {
	if strings.TrimSpace(text) == "" {
		return Result{Invalid: true, Message: "Type your answer first."}
	}
	got, err := frac.Parse(text)
	if err != nil {
		return Result{Invalid: true, Message: "Type a number like 12, 3/4, 1 1/2, or 1.5."}
	}
	want := unpair(o.Answer.Num)
	if got.Equal(want) {
		return Result{Correct: true, Message: "Correct!"}
	}
	if o.Kind == KFactor {
		mk, wt := o.Answer.Makes, o.Answer.Want
		if wt > 0 && got.Equal(frac.New(int64(mk), int64(wt))) {
			return Result{Message: fmt.Sprintf("Check it: does %d × %s make %d? You flipped it. Which number goes on top: what you want, or what the recipe makes?", mk, Mixed(got), wt)}
		}
		return Result{Message: fmt.Sprintf("Check it: what's %d × %s? Is that %d?", mk, Mixed(got), wt)}
	}
	if o.Kind == KSwap {
		given, f := unpair(o.Answer.Given), unpair(o.Answer.Factor)
		switch {
		case got.Equal(given):
			return Result{Message: "That's the number you started with. Are the two units really the same size?"}
		case got.Equal(given.Div(f)):
			return Result{Message: "You went the wrong way. Which unit is bigger? So should your number get bigger or smaller?"}
		}
		return Result{Message: "Not quite. " + o.Hints[1], FreeHint: false}
	}
	if o.Answer.Given[1] != 0 { // scaling a weight or a count
		given, f := unpair(o.Answer.Given), unpair(o.Answer.Factor)
		switch {
		case got.Equal(given):
			return Result{Message: fmt.Sprintf("That's the recipe's original amount. Did you multiply by %s?", Mixed(f))}
		case got.Equal(given.Div(f)):
			return Result{Message: fmt.Sprintf("You went the wrong way. With ×%s, should it get bigger or smaller?", Mixed(f))}
		}
		return Result{Message: fmt.Sprintf("Check your multiplication: %s × %s = ?", Mixed(given), Mixed(f))}
	}
	msg := "Not quite."
	if len(o.Hints) > 0 {
		msg += " " + o.Hints[0]
	}
	return Result{Message: msg, FreeHint: len(o.Hints) > 0}
}
