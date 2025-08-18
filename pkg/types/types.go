package types

// NameData represents the metadata for a single name
type NameData struct {
	Country map[string]float32 // Country code → probability
	Gender  map[string]float32 // "M"/"F" → probability (first names only)
	Rank    map[string]int32   // Country code → rank (1 = most popular)
}

// NameDataset holds the complete name databases
type NameDataset struct {
	FirstNames map[string]*NameData
	LastNames  map[string]*NameData
}

// NameCombination represents a potential split of words into first names and surnames
type NameCombination struct {
	FirstNames []string
	Surnames   []string
}

// PIIResult represents the result of PII name detection
type PIIResult struct {
	IsLikelyName bool    `json:"is_likely_name"`
	Confidence   float64 `json:"confidence"` // 0.0 to 1.0
	Details      NameDetails
}

// NameDetails provides detailed information about the detected name
type NameDetails struct {
	FirstNames []string `json:"first_names"` // Can be multiple: ["Jose", "Manuel"]
	Surnames   []string `json:"surnames"`    // Can be multiple: ["Robles", "Hermoso"]
	Pattern    string   `json:"pattern"`     // e.g., "2_first_2_last"
	TopCountry string   `json:"top_country"` // Most likely country of origin
	Gender     string   `json:"gender"`      // Predicted gender if applicable
}