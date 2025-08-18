package detector

import (
	"testing"

	"github.com/montevive/go-name-detector/pkg/types"
)

// createTestDataset creates a minimal dataset for testing
func createTestDataset() *types.NameDataset {
	dataset := &types.NameDataset{
		FirstNames: make(map[string]*types.NameData),
		LastNames:  make(map[string]*types.NameData),
	}

	// Add some test first names
	dataset.FirstNames["JOSE"] = &types.NameData{
		Country: map[string]float32{"ES": 0.159, "MX": 0.203, "US": 0.098},
		Gender:  map[string]float32{"M": 0.98, "F": 0.02},
		Rank:    map[string]int32{"ES": 1, "MX": 2, "US": 15},
	}

	dataset.FirstNames["MANUEL"] = &types.NameData{
		Country: map[string]float32{"ES": 0.237, "MX": 0.156, "US": 0.045},
		Gender:  map[string]float32{"M": 0.99, "F": 0.01},
		Rank:    map[string]int32{"ES": 7, "MX": 12, "US": 89},
	}

	dataset.FirstNames["MARIA"] = &types.NameData{
		Country: map[string]float32{"ES": 0.198, "MX": 0.234, "US": 0.087},
		Gender:  map[string]float32{"M": 0.01, "F": 0.99},
		Rank:    map[string]int32{"ES": 3, "MX": 1, "US": 25},
	}

	dataset.FirstNames["JOHN"] = &types.NameData{
		Country: map[string]float32{"US": 0.456, "GB": 0.234, "CA": 0.123},
		Gender:  map[string]float32{"M": 0.99, "F": 0.01},
		Rank:    map[string]int32{"US": 8, "GB": 12, "CA": 15},
	}

	// Add some test last names
	dataset.LastNames["GARCIA"] = &types.NameData{
		Country: map[string]float32{"ES": 0.11, "MX": 0.234, "US": 0.156},
		Gender:  map[string]float32{}, // Last names don't have gender
		Rank:    map[string]int32{"ES": 1, "MX": 3, "US": 6},
	}

	dataset.LastNames["LOPEZ"] = &types.NameData{
		Country: map[string]float32{"ES": 0.074, "MX": 0.198, "US": 0.134},
		Gender:  map[string]float32{},
		Rank:    map[string]int32{"ES": 3, "MX": 5, "US": 12},
	}

	dataset.LastNames["ROBLES"] = &types.NameData{
		Country: map[string]float32{"ES": 0.06, "MX": 0.087, "US": 0.023},
		Gender:  map[string]float32{},
		Rank:    map[string]int32{"ES": 141, "MX": 67, "US": 456},
	}

	dataset.LastNames["HERMOSO"] = &types.NameData{
		Country: map[string]float32{"ES": 0.249, "MX": 0.045, "US": 0.012},
		Gender:  map[string]float32{},
		Rank:    map[string]int32{"ES": 3682, "MX": 8934, "US": 15678},
	}

	dataset.LastNames["SMITH"] = &types.NameData{
		Country: map[string]float32{"US": 0.456, "GB": 0.234, "CA": 0.123},
		Gender:  map[string]float32{},
		Rank:    map[string]int32{"US": 1, "GB": 5, "CA": 3},
	}

	return dataset
}

func TestDetectPII_SpanishNames(t *testing.T) {
	dataset := createTestDataset()
	detector := New(dataset)

	tests := []struct {
		name           string
		words          []string
		expectedResult bool
		minConfidence  float64
		expectedFirst  []string
		expectedLast   []string
		threshold      float64
	}{
		{
			name:           "Spanish full name",
			words:          []string{"Jose", "Manuel", "Garcia", "Lopez"},
			expectedResult: true,
			minConfidence:  0.7,
			expectedFirst:  []string{"Jose", "Manuel"},
			expectedLast:   []string{"Garcia", "Lopez"},
			threshold:      0.7,
		},
		{
			name:           "Spanish single first, double last",
			words:          []string{"Jose", "Garcia", "Lopez"},
			expectedResult: true,
			minConfidence:  0.6,
			expectedFirst:  []string{"Jose"},
			expectedLast:   []string{"Garcia", "Lopez"},
			threshold:      0.6,
		},
		{
			name:           "Spanish double first, single last",
			words:          []string{"Jose", "Manuel", "Garcia"},
			expectedResult: true,
			minConfidence:  0.6,
			expectedFirst:  []string{"Jose", "Manuel"},
			expectedLast:   []string{"Garcia"},
			threshold:      0.6,
		},
		{
			name:           "English name",
			words:          []string{"John", "Smith"},
			expectedResult: true,
			minConfidence:  0.5,
			expectedFirst:  []string{"John"},
			expectedLast:   []string{"Smith"},
			threshold:      0.6, // Lower threshold for this test
		},
		{
			name:           "Non-name words",
			words:          []string{"The", "Quick", "Brown", "Fox"},
			expectedResult: false,
			minConfidence:  0.0,
			threshold:      0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectPIIWithThreshold(tt.words, tt.threshold)

			if result.IsLikelyName != tt.expectedResult {
				t.Errorf("Expected IsLikelyName=%v, got %v (confidence: %v)", tt.expectedResult, result.IsLikelyName, result.Confidence)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %v, got %v", tt.minConfidence, result.Confidence)
			}

			if tt.expectedResult {
				if !equalStringSlices(result.Details.FirstNames, tt.expectedFirst) {
					t.Errorf("Expected first names %v, got %v", tt.expectedFirst, result.Details.FirstNames)
				}

				if !equalStringSlices(result.Details.Surnames, tt.expectedLast) {
					t.Errorf("Expected surnames %v, got %v", tt.expectedLast, result.Details.Surnames)
				}
			}
		})
	}
}

func TestDetectPII_EdgeCases(t *testing.T) {
	dataset := createTestDataset()
	detector := New(dataset)

	tests := []struct {
		name  string
		words []string
	}{
		{"Empty input", []string{}},
		{"Single word", []string{"Jose"}},
		{"Too many words", []string{"A", "B", "C", "D", "E", "F", "G"}},
		{"Empty strings", []string{"", "Jose", "", "Garcia", ""}},
		{"Numbers", []string{"John", "123", "Smith"}},
		{"Special characters", []string{"Jose@", "Garcia!"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectPII(tt.words)
			
			// These should all return false or low confidence
			if result.IsLikelyName && result.Confidence > 0.5 {
				t.Errorf("Expected low confidence for edge case, got %v", result.Confidence)
			}
		})
	}
}

func TestDetectPII_Thresholds(t *testing.T) {
	dataset := createTestDataset()
	detector := New(dataset)

	words := []string{"Jose", "Garcia"}

	// Test different thresholds
	thresholds := []float64{0.3, 0.5, 0.7, 0.9}

	for _, threshold := range thresholds {
		result := detector.DetectPIIWithThreshold(words, threshold)
		
		if result.Confidence >= threshold && !result.IsLikelyName {
			t.Errorf("Confidence %v >= threshold %v but IsLikelyName is false", 
				result.Confidence, threshold)
		}
		
		if result.Confidence < threshold && result.IsLikelyName {
			t.Errorf("Confidence %v < threshold %v but IsLikelyName is true", 
				result.Confidence, threshold)
		}
	}
}

// Helper function to compare string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Benchmark tests
func BenchmarkDetectPII_Spanish(b *testing.B) {
	dataset := createTestDataset()
	detector := New(dataset)
	words := []string{"Jose", "Manuel", "Garcia", "Lopez"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectPII(words)
	}
}

func BenchmarkDetectPII_English(b *testing.B) {
	dataset := createTestDataset()
	detector := New(dataset)
	words := []string{"John", "Smith"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectPII(words)
	}
}

func BenchmarkDetectPII_NonName(b *testing.B) {
	dataset := createTestDataset()
	detector := New(dataset)
	words := []string{"The", "Quick", "Brown", "Fox"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectPII(words)
	}
}