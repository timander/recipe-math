//go:build js && wasm

// Command wasm exposes the Go recipe-math engine to the web page as
// window.recipeMath. Every function takes and returns JSON strings.
package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/timander/recipe-math/kitchen"
)

func main() {
	api := map[string]any{
		// ranks() → [{id, name, goal, example}]
		"ranks": fn(func([]js.Value) any { return kitchen.Ranks }),
		// reference() → cheat-sheet content
		"reference": fn(func([]js.Value) any { return kitchen.Ref() }),
		// shift(rank, seed) → {orders} or {error}
		"shift": fn(func(args []js.Value) any {
			orders, err := kitchen.Shift(args[0].Int(), uint64(args[1].Float()))
			if err != nil {
				return map[string]string{"error": err.Error()}
			}
			return map[string]any{"orders": orders}
		}),
		// check(orderJSON, responseJSON) → {correct, invalid, message, ...}
		"check": fn(func(args []js.Value) any {
			var o kitchen.Order
			var r kitchen.Response
			if err := json.Unmarshal([]byte(args[0].String()), &o); err != nil {
				return map[string]string{"error": err.Error()}
			}
			if err := json.Unmarshal([]byte(args[1].String()), &r); err != nil {
				return map[string]string{"error": err.Error()}
			}
			return kitchen.Check(o, r)
		}),
	}
	js.Global().Set("recipeMath", js.ValueOf(api))
	if ready := js.Global().Get("recipeMathReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}
	select {} // keep the Go program alive for callbacks
}

// fn wraps a Go function so JS gets a JSON string back.
func fn(f func(args []js.Value) any) js.Func {
	return js.FuncOf(func(_ js.Value, args []js.Value) (result any) {
		// A panic would stop the Go program for good; report it instead.
		defer func() {
			if r := recover(); r != nil {
				b, _ := json.Marshal(map[string]string{"error": fmt.Sprint(r)})
				result = string(b)
			}
		}()
		b, err := json.Marshal(f(args))
		if err != nil {
			b, _ = json.Marshal(map[string]string{"error": err.Error()})
		}
		return string(b)
	})
}
