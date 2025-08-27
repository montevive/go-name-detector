package detector

import (
	"testing"
)

func TestNormalizeAccents(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"José", "Jose"},
		{"García", "Garcia"},
		{"François", "Francois"},
		{"Müller", "Muller"},
		{"Andrés", "Andres"},
		{"María", "Maria"},
		{"González", "Gonzalez"},
		{"Hernández", "Hernandez"},
		{"López", "Lopez"},
		{"Martínez", "Martinez"},
		{"Rodríguez", "Rodriguez"},
		{"Sánchez", "Sanchez"},
		{"Pérez", "Perez"},
		{"Ramón", "Ramon"},
		{"Ángel", "Angel"},
		{"José Manuel", "Jose Manuel"},
		{"María García", "Maria Garcia"},
		{"John Smith", "John Smith"}, // No accents should remain unchanged
		{"", ""}, // Empty string
	}

	for _, tt := range tests {
		result := normalizeAccents(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeAccents(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeForLookup(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"José", "JOSE"},
		{"García", "GARCIA"},
		{"  José  ", "JOSE"}, // Test trimming
		{"josé", "JOSE"}, // Test case conversion
		{"JOSÉ", "JOSE"},
		{"María García", "MARIA GARCIA"},
		{"John Smith", "JOHN SMITH"},
	}

	for _, tt := range tests {
		result := normalizeForLookup(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeForLookup(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// Benchmark the normalization function
func BenchmarkNormalizeAccents(b *testing.B) {
	testNames := []string{"José", "García", "François", "Müller", "María García López"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, name := range testNames {
			normalizeAccents(name)
		}
	}
}