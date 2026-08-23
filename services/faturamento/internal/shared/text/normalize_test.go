package text

import "testing"

func TestNormalizeUpper(t *testing.T) {
	result := NormalizeUpper("  Maria da Silva  ")
	if result != "MARIA DA SILVA" {
		t.Fatalf("esperava texto normalizado, recebeu %q", result)
	}
}
