package detector

import (
	"fmt"
	"strings"

	"github.com/montevive/go-name-detector/pkg/loader"
	"github.com/montevive/go-name-detector/pkg/types"
)

// Detector handles PII name detection
type Detector struct {
	scorer *Scorer
}

// New creates a new Detector with the given dataset
func New(dataset *types.NameDataset) *Detector {
	config := DefaultScoreConfig()
	scorer := NewScorer(dataset, config)
	
	return &Detector{
		scorer: scorer,
	}
}

// NewWithConfig creates a new Detector with custom scoring configuration
func NewWithConfig(dataset *types.NameDataset, config ScoreConfig) *Detector {
	scorer := NewScorer(dataset, config)
	
	return &Detector{
		scorer: scorer,
	}
}

// NewDefault creates a new Detector with embedded dataset - ready to use out of the box
func NewDefault() (*Detector, error) {
	l, err := loader.NewWithEmbeddedData()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize detector with embedded data: %w", err)
	}
	
	return New(l.GetDataset()), nil
}

// DetectPII analyzes words to determine if they represent a PII name
func (d *Detector) DetectPII(words []string) types.PIIResult {
	return d.DetectPIIWithThreshold(words, 0.7) // Default threshold
}

// DetectPIIWithThreshold analyzes words with a custom confidence threshold
func (d *Detector) DetectPIIWithThreshold(words []string, threshold float64) types.PIIResult {
	if len(words) < 2 || len(words) > 6 {
		return types.PIIResult{
			IsLikelyName: false,
			Confidence:   0.0,
			Details: types.NameDetails{
				Pattern: "invalid_length",
			},
		}
	}

	// Clean and normalize words
	cleanWords := d.cleanWords(words)
	if len(cleanWords) < 2 {
		return types.PIIResult{
			IsLikelyName: false,
			Confidence:   0.0,
			Details: types.NameDetails{
				Pattern: "insufficient_words",
			},
		}
	}

	// Generate all possible name combinations
	combinations := d.generateCombinations(cleanWords)
	
	// Score each combination and find the best one
	bestCombo, bestScore := d.findBestCombination(combinations)

	// Determine if it's likely a name
	isLikelyName := bestScore >= threshold

	// Build result details
	pattern := d.buildPattern(bestCombo)
	topCountry := d.scorer.GetTopCountry(bestCombo)
	gender := d.scorer.GetGender(bestCombo)

	return types.PIIResult{
		IsLikelyName: isLikelyName,
		Confidence:   bestScore,
		Details: types.NameDetails{
			FirstNames: bestCombo.FirstNames,
			Surnames:   bestCombo.Surnames,
			Pattern:    pattern,
			TopCountry: topCountry,
			Gender:     gender,
		},
	}
}

// cleanWords removes empty strings, trims whitespace, and filters invalid words
func (d *Detector) cleanWords(words []string) []string {
	var cleaned []string
	
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) == 0 {
			continue
		}
		
		// Skip words that are clearly not names (too short, numbers, special chars)
		if d.isValidNameWord(word) {
			cleaned = append(cleaned, word)
		}
	}
	
	return cleaned
}

// isValidNameWord checks if a word could plausibly be part of a name
func (d *Detector) isValidNameWord(word string) bool {
	// Must be at least 2 characters
	if len(word) < 2 {
		return false
	}
	
	// Must contain only letters (and possibly hyphens, apostrophes)
	for _, r := range word {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '-' || r == '\'' || r == '.') {
			return false
		}
	}
	
	// Skip common non-name words
	lowerWord := strings.ToLower(word)
	commonWords := map[string]bool{
		"the": true, "and": true, "or": true, "but": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "of": true, "with": true, "by": true,
		"is": true, "are": true, "was": true, "were": true, "be": true, "been": true,
		"have": true, "has": true, "had": true, "do": true, "does": true, "did": true,
		"will": true, "would": true, "could": true, "should": true, "may": true, "might": true,
		"can": true, "must": true, "shall": true, "this": true, "that": true, "these": true,
		"those": true, "a": true, "an": true, "it": true, "he": true, "she": true,
		"they": true, "we": true, "you": true, "i": true, "me": true, "him": true,
		"her": true, "them": true, "us": true, "my": true, "your": true, "his": true,
		"our": true, "their": true, "its": true,
	}
	
	return !commonWords[lowerWord]
}

// generateCombinations creates all possible splits of words into first names and surnames
func (d *Detector) generateCombinations(words []string) []types.NameCombination {
	var combinations []types.NameCombination
	
	// Try all possible splits where at least 1 word is first name and 1 is surname
	for i := 1; i < len(words); i++ {
		combo := types.NameCombination{
			FirstNames: words[:i],
			Surnames:   words[i:],
		}
		combinations = append(combinations, combo)
	}
	
	return combinations
}

// findBestCombination scores all combinations and returns the best one
func (d *Detector) findBestCombination(combinations []types.NameCombination) (types.NameCombination, float64) {
	var bestCombo types.NameCombination
	var bestScore float64
	
	for _, combo := range combinations {
		score := d.scorer.ScoreCombination(combo)
		if score > bestScore {
			bestScore = score
			bestCombo = combo
		}
	}
	
	return bestCombo, bestScore
}

// buildPattern creates a pattern string describing the name structure
func (d *Detector) buildPattern(combo types.NameCombination) string {
	firstCount := len(combo.FirstNames)
	lastCount := len(combo.Surnames)
	
	return fmt.Sprintf("%d_first_%d_last", firstCount, lastCount)
}

// GetDatasetStats returns statistics about the loaded dataset
func (d *Detector) GetDatasetStats() map[string]interface{} {
	if d.scorer == nil || d.scorer.dataset == nil {
		return map[string]interface{}{
			"error": "dataset not loaded",
		}
	}
	
	return map[string]interface{}{
		"first_names_count": len(d.scorer.dataset.FirstNames),
		"last_names_count":  len(d.scorer.dataset.LastNames),
	}
}