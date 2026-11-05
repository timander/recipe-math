//go:build js && wasm

// Command wasm exposes the Go recipe-math engine to the web page as
// window.recipeMath. Every function takes and returns JSON strings.
package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/timander/recipe-math/problems"
	"github.com/timander/recipe-math/units"
)

type levelInfo struct {
	problems.Level
	PoolSize int `json:"poolSize"`
}

type factInfo struct {
	Text   string `json:"text"`
	Big    string `json:"big"`
	Small  string `json:"small"`
	Approx bool   `json:"approx"`
}

func main() {
	api := map[string]any{
		// levels() → [{id, name, goal, example, poolSize}]
		"levels": fn(func(args []js.Value) any {
			var out []levelInfo
			for _, l := range problems.Levels {
				out = append(out, levelInfo{l, problems.PoolSize(l.ID)})
			}
			return out
		}),
		// reference() → {units, facts, tips} for the cheat sheet
		"reference": fn(func(args []js.Value) any {
			var facts []factInfo
			for _, f := range units.Facts {
				facts = append(facts, factInfo{f.String(), f.Big, f.Small, f.Approx})
			}
			return map[string]any{"units": units.All, "facts": facts, "tips": units.Tips}
		}),
		// generate(levelsJSON, count, seed) → {problems} or {error}
		"generate": fn(func(args []js.Value) any {
			var levels []int
			if err := json.Unmarshal([]byte(args[0].String()), &levels); err != nil {
				return map[string]string{"error": err.Error()}
			}
			ps, err := problems.Generate(levels, args[1].Int(), uint64(args[2].Float()))
			if err != nil {
				return map[string]string{"error": err.Error()}
			}
			return map[string]any{"problems": ps}
		}),
		// check(problemJSON, input) → {status, message}
		"check": fn(func(args []js.Value) any {
			var p problems.Problem
			if err := json.Unmarshal([]byte(args[0].String()), &p); err != nil {
				return map[string]string{"error": err.Error()}
			}
			return problems.Check(p, args[1].String())
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
	return js.FuncOf(func(this js.Value, args []js.Value) (result any) {
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
