package detector

import (
	"strings"
	"unicode"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// normalizeAccents removes accents and diacritical marks from a string
// Example: "José García" -> "Jose Garcia"
func normalizeAccents(s string) string {
	// Create a transformer that removes accents by decomposing Unicode characters
	// and then removing the combining diacritical marks
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	
	// Apply the transformation
	result, _, err := transform.String(t, s)
	if err != nil {
		// If transformation fails, return original string
		return s
	}
	
	return result
}

// normalizeForLookup normalizes a name for database lookup
// This applies both accent normalization and case normalization
func normalizeForLookup(name string) string {
	// First normalize accents, then trim and convert to uppercase
	normalized := normalizeAccents(name)
	return strings.ToUpper(strings.TrimSpace(normalized))
}