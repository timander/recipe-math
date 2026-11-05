package kitchen

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata/")

// TestGoldenShifts pins every order students see (prompts, choices, hints,
// explanations, answers, pictures) for a few seeded shifts per station. Any
// change to wording or math shows up as a diff; if it's intended, run
//
//	go test ./kitchen -run TestGoldenShifts -update
//
// and review the golden files in the commit.
func TestGoldenShifts(t *testing.T) {
	for _, r := range Ranks {
		for _, seed := range []uint64{1, 2, 3, 4} {
			name := fmt.Sprintf("rank%d_seed%d.json", r.ID, seed)
			t.Run(name, func(t *testing.T) {
				orders, err := Shift(r.ID, seed)
				if err != nil {
					t.Fatal(err)
				}
				got, err := json.MarshalIndent(orders, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				got = append(got, '\n')
				path := filepath.Join("testdata", "golden", name)
				if *update {
					if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
						t.Fatal(err)
					}
					if err = os.WriteFile(path, got, 0o644); err != nil {
						t.Fatal(err)
					}
					return
				}
				want, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("missing golden file (run with -update): %v", err)
				}
				if !bytes.Equal(got, want) {
					t.Errorf("%s changed; if intended, rerun with -update and review the diff", path)
				}
			})
		}
	}
}
