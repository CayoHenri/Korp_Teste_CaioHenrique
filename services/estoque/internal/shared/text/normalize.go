package text

import "strings"

// NormalizeUpper remove espacos externos e padroniza o texto em letras maiusculas.
func NormalizeUpper(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
