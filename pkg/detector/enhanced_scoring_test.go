package detector

import (
	"testing"

	"github.com/montevive/go-name-detector/pkg/types"
)

func TestDetectPII_EnhancedScoring(t *testing.T) {
	dataset := createTestDataset()
	detector := New(dataset)

	tests := []struct {
		name           string
		words          []string
		expectedResult bool
		minConfidence  float64
		maxConfidence  float64
		threshold      float64
		description    string
	}{
		{
			name:           "Top-ranked Spanish names should score very high",
			words:          []string{"José", "García"},
			expectedResult: true,
			minConfidence:  0.7, // Should be well above threshold now
			maxConfidence:  1.0,
			threshold:      0.7,
			description:    "José (rank #1) + García (rank #1) should get top-name bonus",
		},
		{
			name:           "Business terms with prepositions should score low",
			words:          []string{"Informe", "de", "Cliente"},
			expectedResult: false,
			minConfidence:  0.0,
			maxConfidence:  0.3, // Should be much lower than before
			threshold:      0.7,
			description:    "Should be penalized for 'de' preposition + poor rank for 'Cliente'",
		},
		{
			name:           "Document terms should score low",
			words:          []string{"Documento", "de", "Identidad"},
			expectedResult: false,
			minConfidence:  0.0,
			maxConfidence:  0.3,
			threshold:      0.7,
			description:    "Common document terminology should be rejected",
		},
		{
			name:           "Mixed valid name with preposition",
			words:          []string{"María", "de", "García"},
			expectedResult: true, // María and García are top names
			minConfidence:  0.4,
			maxConfidence:  1.0,
			threshold:      0.4, // Lower threshold to see if it works
			description:    "Should detect despite 'de' because María and García are top-ranked",
		},
		{
			name:           "Rare names should score much lower",
			words:          []string{"Cliente", "Nuevo"}, // If Nuevo exists and has poor rank
			expectedResult: false,
			minConfidence:  0.0,
			maxConfidence:  0.3,
			threshold:      0.7,
			description:    "Poor-ranked names should get minimal scores",
		},
		{
			name:           "Top names with different threshold",
			words:          []string{"José", "García"},
			expectedResult: true,
			minConfidence:  0.8, // Should score very high with new system
			maxConfidence:  1.0,
			threshold:      0.8, // Higher threshold
			description:    "Top-ranked names should easily pass higher thresholds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing: %s", tt.description)
			result := detector.DetectPIIWithThreshold(tt.words, tt.threshold)

			if result.IsLikelyName != tt.expectedResult {
				t.Errorf("Expected IsLikelyName=%v, got %v (confidence: %.3f)", 
					tt.expectedResult, result.IsLikelyName, result.Confidence)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %.3f, got %.3f", tt.minConfidence, result.Confidence)
			}

			if result.Confidence > tt.maxConfidence {
				t.Errorf("Expected confidence <= %.3f, got %.3f", tt.maxConfidence, result.Confidence)
			}

			t.Logf("Result: IsLikelyName=%v, Confidence=%.3f, Pattern=%s", 
				result.IsLikelyName, result.Confidence, result.Details.Pattern)
		})
	}
}

// Test the step-based popularity scoring function directly
func TestCalculatePopularityScore(t *testing.T) {
	dataset := createTestDataset()
	scorer := NewScorer(dataset, DefaultScoreConfig())

	// Test the new step-based scoring
	tests := []struct {
		rank     int32
		expected float64
		tier     string
	}{
		{1, 1.0, "top-10"},
		{5, 1.0, "top-10"},
		{10, 1.0, "top-10"},
		{25, 0.8, "top-50"},
		{50, 0.8, "top-50"},
		{100, 0.5, "top-200"},
		{200, 0.5, "top-200"},
		{500, 0.2, "top-1000"},
		{1000, 0.2, "top-1000"},
		{5000, 0.02, "rare"},
		{10000, 0.02, "rare"},
	}

	for _, tt := range tests {
		// Create mock name data
		nameData := &types.NameData{
			Rank: map[string]int32{"TEST": tt.rank},
		}

		score := scorer.calculatePopularityScore(nameData)
		if score != tt.expected {
			t.Errorf("Rank %d (%s tier): expected %.2f, got %.2f", 
				tt.rank, tt.tier, tt.expected, score)
		}
	}
}