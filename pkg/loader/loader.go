package loader

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	names "github.com/montevive/go-name-detector/pkg/proto"
	"github.com/montevive/go-name-detector/pkg/types"
	"google.golang.org/protobuf/proto"
)

// Loader handles loading and caching of name data
type Loader struct {
	dataset *types.NameDataset
	loaded  bool
}

// New creates a new Loader instance
func New() *Loader {
	return &Loader{
		dataset: &types.NameDataset{
			FirstNames: make(map[string]*types.NameData),
			LastNames:  make(map[string]*types.NameData),
		},
		loaded: false,
	}
}

// LoadFromFile loads name data from a protobuf file
func (l *Loader) LoadFromFile(filename string) error {
	if l.loaded {
		return nil // Already loaded
	}

	// Read and decompress file
	data, err := l.readFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Parse protobuf
	var pbDataset names.CombinedNameDataset
	if err := proto.Unmarshal(data, &pbDataset); err != nil {
		return fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	// Convert to internal format
	l.convertToInternalFormat(&pbDataset)
	l.loaded = true

	return nil
}

// LoadFromBytes loads name data from a byte array (supports gzip compression)
func (l *Loader) LoadFromBytes(data []byte) error {
	if l.loaded {
		return nil // Already loaded
	}

	// Check if data is gzip compressed (magic bytes: 0x1f, 0x8b)
	var decompressedData []byte
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		// Decompress gzip data
		reader := bytes.NewReader(data)
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()

		decompressedData, err = io.ReadAll(gzipReader)
		if err != nil {
			return fmt.Errorf("failed to decompress data: %w", err)
		}
	} else {
		// Data is not compressed
		decompressedData = data
	}

	// Parse protobuf
	var pbDataset names.CombinedNameDataset
	if err := proto.Unmarshal(decompressedData, &pbDataset); err != nil {
		return fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	// Convert to internal format
	l.convertToInternalFormat(&pbDataset)
	l.loaded = true

	return nil
}

// LoadSeparateFiles loads first and last names from separate files
func (l *Loader) LoadSeparateFiles(firstNamesFile, lastNamesFile string) error {
	if l.loaded {
		return nil // Already loaded
	}

	// Load first names
	if err := l.loadSingleDataset(firstNamesFile, true); err != nil {
		return fmt.Errorf("failed to load first names: %w", err)
	}

	// Load last names
	if err := l.loadSingleDataset(lastNamesFile, false); err != nil {
		return fmt.Errorf("failed to load last names: %w", err)
	}

	l.loaded = true
	return nil
}

// loadSingleDataset loads a single name dataset
func (l *Loader) loadSingleDataset(filename string, isFirstNames bool) error {
	data, err := l.readFile(filename)
	if err != nil {
		return err
	}

	var pbDataset names.NameDataset
	if err := proto.Unmarshal(data, &pbDataset); err != nil {
		return fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	// Convert entries
	targetMap := l.dataset.LastNames
	if isFirstNames {
		targetMap = l.dataset.FirstNames
	}

	for _, entry := range pbDataset.Entries {
		nameData := &types.NameData{
			Country: entry.Country,
			Gender:  entry.Gender,
			Rank:    entry.Rank,
		}

		// Store with normalized key for case-insensitive lookup
		normalizedName := strings.ToUpper(strings.TrimSpace(entry.Name))
		targetMap[normalizedName] = nameData
	}

	return nil
}

// readFile reads a file (with optional gzip decompression)
func (l *Loader) readFile(filename string) ([]byte, error) {
	if strings.HasSuffix(filename, ".gz") {
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()

		// Use io.ReadAll instead of strings.Builder.ReadFrom
		data, err := io.ReadAll(gzipReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read decompressed data: %w", err)
		}

		return data, nil
	}

	return os.ReadFile(filename)
}

// convertToInternalFormat converts protobuf data to internal format
func (l *Loader) convertToInternalFormat(pbDataset *names.CombinedNameDataset) {
	// Convert first names
	for _, entry := range pbDataset.FirstNames.Entries {
		nameData := &types.NameData{
			Country: entry.Country,
			Gender:  entry.Gender,
			Rank:    entry.Rank,
		}
		
		// Store with normalized key for case-insensitive lookup
		normalizedName := strings.ToUpper(strings.TrimSpace(entry.Name))
		l.dataset.FirstNames[normalizedName] = nameData
	}

	// Convert last names
	for _, entry := range pbDataset.LastNames.Entries {
		nameData := &types.NameData{
			Country: entry.Country,
			Gender:  entry.Gender,
			Rank:    entry.Rank,
		}
		
		// Store with normalized key for case-insensitive lookup
		normalizedName := strings.ToUpper(strings.TrimSpace(entry.Name))
		l.dataset.LastNames[normalizedName] = nameData
	}
}

// GetDataset returns the loaded dataset
func (l *Loader) GetDataset() *types.NameDataset {
	return l.dataset
}

// IsLoaded returns whether the dataset has been loaded
func (l *Loader) IsLoaded() bool {
	return l.loaded
}

// GetStats returns statistics about the loaded dataset
func (l *Loader) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"first_names_count": len(l.dataset.FirstNames),
		"last_names_count":  len(l.dataset.LastNames),
		"loaded":            l.loaded,
	}
}

// LoadEmbedded loads the embedded dataset (convenience method)
func (l *Loader) LoadEmbedded() error {
	return l.LoadFromBytes(EmbeddedData)
}

// NewWithEmbeddedData creates a new Loader with embedded data pre-loaded
func NewWithEmbeddedData() (*Loader, error) {
	l := New()
	if err := l.LoadEmbedded(); err != nil {
		return nil, fmt.Errorf("failed to load embedded data: %w", err)
	}
	return l, nil
}