package ds

import (
	"testing"
)

func TestSource(t *testing.T) {

	srcnames := map[string]string{
		"":        "*_*_*_*",
		"_":       "*_*_*_*",
		"__":      "*_*_*_*",
		"___":     "*_*_*_*",
		"A":       "A_*_*_*",
		"A_B":     "A_B_*_*",
		"A_B_C":   "A_B_*_C",
		"A_B_C_D": "A_B_C_D",
		"A_B__C":  "A_B_*_C",
	}

	for k, v := range srcnames {
		t.Run("check "+k, func(t *testing.T) {
			t.Log(k)
			if s := NewSource(k); s.String() != v {
				t.Errorf("source mismatch for %s, expected %s got %s", k, v, s.String())
			}
		})
	}
}
