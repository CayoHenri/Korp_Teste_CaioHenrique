package text

import "testing"

func TestNormalizeUpper(t *testing.T) {
	result := NormalizeUpper("  Teclado Mecânico  ")
	if result != "TECLADO MECÂNICO" {
		t.Fatalf("esperava texto normalizado, recebeu %q", result)
	}
}
