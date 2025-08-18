package detector

import (
	"math"
	"strings"

	"github.com/montevive/go-name-detector/pkg/types"
)

// ScoreConfig holds configuration for the scoring algorithm
type ScoreConfig struct {
	BaseMatchScore     float64 // Base score for finding a name in database
	PopularityWeight   float64 // Weight for popularity (lower rank = higher score)
	GenderConsistency  float64 // Bonus for consistent gender across first names
	CountryOverlap     float64 // Bonus for country overlap between components
	MultipleNamesBonus float64 // Bonus for finding multiple valid names
}

// DefaultScoreConfig returns the default scoring configuration
func DefaultScoreConfig() ScoreConfig {
	return ScoreConfig{
		BaseMatchScore:     0.3,
		PopularityWeight:   0.2,
		GenderConsistency:  0.1,
		CountryOverlap:     0.2,
		MultipleNamesBonus: 0.15,
	}
}

// Scorer handles confidence scoring for name combinations
type Scorer struct {
	config  ScoreConfig
	dataset *types.NameDataset
}

// NewScorer creates a new scorer with the given dataset and config
func NewScorer(dataset *types.NameDataset, config ScoreConfig) *Scorer {
	return &Scorer{
		config:  config,
		dataset: dataset,
	}
}

// ScoreCombination calculates a confidence score for a name combination
func (s *Scorer) ScoreCombination(combo types.NameCombination) float64 {
	if len(combo.FirstNames) == 0 || len(combo.Surnames) == 0 {
		return 0.0
	}

	var totalScore float64
	var componentCount int

	// Score first names
	firstNamesScore, firstNamesData := s.scoreNames(combo.FirstNames, true)
	totalScore += firstNamesScore
	componentCount += len(combo.FirstNames)

	// Score surnames
	surnamesScore, surnamesData := s.scoreNames(combo.Surnames, false)
	totalScore += surnamesScore
	componentCount += len(combo.Surnames)

	if componentCount == 0 {
		return 0.0
	}

	// Average base score
	averageScore := totalScore / float64(componentCount)

	// Add bonus for gender consistency among first names
	if len(firstNamesData) > 1 {
		genderBonus := s.calculateGenderConsistency(firstNamesData)
		averageScore += genderBonus
	}

	// Add bonus for country overlap between first names and surnames
	if len(firstNamesData) > 0 && len(surnamesData) > 0 {
		countryBonus := s.calculateCountryOverlap(firstNamesData, surnamesData)
		averageScore += countryBonus
	}

	// Add bonus for multiple valid names
	if componentCount > 2 {
		averageScore += s.config.MultipleNamesBonus
	}

	// Clamp to [0, 1]
	if averageScore > 1.0 {
		averageScore = 1.0
	}

	return averageScore
}

// scoreNames scores a list of names (either first names or surnames)
func (s *Scorer) scoreNames(names []string, isFirstNames bool) (float64, []*types.NameData) {
	var totalScore float64
	var nameDataList []*types.NameData

	targetMap := s.dataset.LastNames
	if isFirstNames {
		targetMap = s.dataset.FirstNames
	}

	for _, name := range names {
		normalizedName := strings.ToUpper(strings.TrimSpace(name))
		nameData, exists := targetMap[normalizedName]

		if !exists {
			// Name not found in database
			continue
		}

		nameDataList = append(nameDataList, nameData)

		// Base score for match
		score := s.config.BaseMatchScore

		// Add popularity bonus (lower rank = higher score)
		popularityScore := s.calculatePopularityScore(nameData)
		score += popularityScore * s.config.PopularityWeight

		totalScore += score
	}

	return totalScore, nameDataList
}

// calculatePopularityScore calculates score based on name popularity
func (s *Scorer) calculatePopularityScore(nameData *types.NameData) float64 {
	if len(nameData.Rank) == 0 {
		return 0.0
	}

	// Find the best (lowest) rank across all countries
	minRank := int32(math.MaxInt32)
	for _, rank := range nameData.Rank {
		if rank > 0 && rank < minRank {
			minRank = rank
		}
	}

	if minRank == int32(math.MaxInt32) {
		return 0.0
	}

	// Convert rank to score: 1 / log(rank + 1)
	// This gives higher scores to more popular names (lower ranks)
	return 1.0 / math.Log(float64(minRank)+1.0)
}

// calculateGenderConsistency calculates bonus for consistent gender across first names
func (s *Scorer) calculateGenderConsistency(firstNamesData []*types.NameData) float64 {
	if len(firstNamesData) < 2 {
		return 0.0
	}

	// Count votes for each gender
	genderVotes := make(map[string]float32)
	totalVotes := float32(0)

	for _, nameData := range firstNamesData {
		for gender, probability := range nameData.Gender {
			genderVotes[gender] += probability
			totalVotes += probability
		}
	}

	if totalVotes == 0 {
		return 0.0
	}

	// Find the dominant gender
	var maxVotes float32
	for _, votes := range genderVotes {
		if votes > maxVotes {
			maxVotes = votes
		}
	}

	// Calculate consistency ratio
	consistency := float64(maxVotes / totalVotes)

	// Only give bonus if consistency is high (> 0.7)
	if consistency > 0.7 {
		return s.config.GenderConsistency * consistency
	}

	return 0.0
}

// calculateCountryOverlap calculates bonus for country overlap between first names and surnames
func (s *Scorer) calculateCountryOverlap(firstNamesData, surnamesData []*types.NameData) float64 {
	// Aggregate country probabilities for first names
	firstCountries := make(map[string]float32)
	for _, nameData := range firstNamesData {
		for country, prob := range nameData.Country {
			firstCountries[country] += prob
		}
	}

	// Aggregate country probabilities for surnames
	lastCountries := make(map[string]float32)
	for _, nameData := range surnamesData {
		for country, prob := range nameData.Country {
			lastCountries[country] += prob
		}
	}

	// Calculate overlap score
	var overlapScore float64
	for country, firstProb := range firstCountries {
		if lastProb, exists := lastCountries[country]; exists {
			// Both first and last names have this country
			// Score is the minimum of the two probabilities
			overlap := math.Min(float64(firstProb), float64(lastProb))
			overlapScore += overlap
		}
	}

	return s.config.CountryOverlap * overlapScore
}

// GetTopCountry returns the most likely country for a name combination
func (s *Scorer) GetTopCountry(combo types.NameCombination) string {
	countryScores := make(map[string]float64)

	// Add scores from first names
	for _, name := range combo.FirstNames {
		normalizedName := strings.ToUpper(strings.TrimSpace(name))
		if nameData, exists := s.dataset.FirstNames[normalizedName]; exists {
			for country, prob := range nameData.Country {
				countryScores[country] += float64(prob)
			}
		}
	}

	// Add scores from surnames
	for _, name := range combo.Surnames {
		normalizedName := strings.ToUpper(strings.TrimSpace(name))
		if nameData, exists := s.dataset.LastNames[normalizedName]; exists {
			for country, prob := range nameData.Country {
				countryScores[country] += float64(prob)
			}
		}
	}

	// Find the country with highest score
	var topCountry string
	var maxScore float64
	for country, score := range countryScores {
		if score > maxScore {
			maxScore = score
			topCountry = country
		}
	}

	return topCountry
}

// GetGender returns the predicted gender for the first names in a combination
func (s *Scorer) GetGender(combo types.NameCombination) string {
	if len(combo.FirstNames) == 0 {
		return ""
	}

	genderScores := make(map[string]float64)

	// Aggregate gender scores from all first names
	for _, name := range combo.FirstNames {
		normalizedName := strings.ToUpper(strings.TrimSpace(name))
		if nameData, exists := s.dataset.FirstNames[normalizedName]; exists {
			for gender, prob := range nameData.Gender {
				genderScores[gender] += float64(prob)
			}
		}
	}

	// Find the gender with highest score
	var predictedGender string
	var maxScore float64
	for gender, score := range genderScores {
		if score > maxScore {
			maxScore = score
			if gender == "M" {
				predictedGender = "Male"
			} else if gender == "F" {
				predictedGender = "Female"
			}
		}
	}

	return predictedGender
}