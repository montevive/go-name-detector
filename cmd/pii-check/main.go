package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/montevive/go-name-detector/pkg/detector"
	"github.com/montevive/go-name-detector/pkg/loader"
	"github.com/montevive/go-name-detector/pkg/types"
)

var (
	dataPath   = flag.String("data", "data/combined_names.pb.gz", "Path to the protobuf data file")
	threshold  = flag.Float64("threshold", 0.7, "Confidence threshold for PII detection")
	jsonOutput = flag.Bool("json", false, "Output results in JSON format")
	batch      = flag.String("batch", "", "Process names from a file (one per line)")
	stats      = flag.Bool("stats", false, "Show dataset statistics")
	help       = flag.Bool("help", false, "Show help information")
)

func main() {
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load the dataset
	fmt.Fprintf(os.Stderr, "Loading dataset from %s...\n", *dataPath)
	startTime := time.Now()

	l := loader.New()
	if err := l.LoadFromFile(*dataPath); err != nil {
		log.Fatalf("Failed to load dataset: %v", err)
	}

	loadTime := time.Since(startTime)
	fmt.Fprintf(os.Stderr, "Dataset loaded in %v\n", loadTime)

	// Create detector
	d := detector.New(l.GetDataset())

	// Show stats if requested
	if *stats {
		showStats(d, l)
		return
	}

	// Process batch file if specified
	if *batch != "" {
		processBatchFile(*batch, d)
		return
	}

	// Process command line arguments
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No input provided. Use -help for usage information.\n")
		os.Exit(1)
	}

	// Join all arguments as a single string and split by spaces
	input := strings.Join(args, " ")
	words := strings.Fields(input)

	// Detect PII
	result := d.DetectPIIWithThreshold(words, *threshold)

	// Output result
	if *jsonOutput {
		outputJSON(result)
	} else {
		outputHuman(result, words)
	}
}

func showHelp() {
	fmt.Printf(`PII Name Detector

Usage:
  pii-check [OPTIONS] <words>
  pii-check -batch <file>

Examples:
  pii-check "John Smith"
  pii-check "Jose Manuel Robles Hermoso"
  pii-check -threshold 0.8 "Maria Garcia Lopez"
  pii-check -json "Antonio Perez"
  pii-check -batch names.txt
  pii-check -stats

Options:
  -data <path>       Path to protobuf data file (default: data/combined_names.pb.gz)
  -threshold <val>   Confidence threshold for PII detection (default: 0.7)
  -json             Output in JSON format
  -batch <file>     Process names from file (one per line)
  -stats            Show dataset statistics
  -help             Show this help

The tool analyzes 2-6 words to determine if they represent a PII name.
It returns a confidence score and detailed breakdown of the analysis.
`)
}

func showStats(d *detector.Detector, l *loader.Loader) {
	loaderStats := l.GetStats()
	detectorStats := d.GetDatasetStats()

	fmt.Printf("Dataset Statistics:\n")
	fmt.Printf("  First names: %v\n", detectorStats["first_names_count"])
	fmt.Printf("  Last names:  %v\n", detectorStats["last_names_count"])
	fmt.Printf("  Loaded:      %v\n", loaderStats["loaded"])
	
	total := detectorStats["first_names_count"].(int) + detectorStats["last_names_count"].(int)
	fmt.Printf("  Total names: %d\n", total)
}

func processBatchFile(filename string, d *detector.Detector) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open batch file: %v", err)
	}
	defer file.Close()

	// Read file contents
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read batch file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var processed, detected int

	fmt.Printf("Processing %d lines from %s...\n", len(lines), filename)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		words := strings.Fields(line)
		if len(words) < 2 || len(words) > 6 {
			continue
		}

		result := d.DetectPIIWithThreshold(words, *threshold)
		processed++

		if result.IsLikelyName {
			detected++
		}

		if *jsonOutput {
			output := map[string]interface{}{
				"line":   i + 1,
				"input":  line,
				"result": result,
			}
			jsonBytes, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(jsonBytes))
		} else {
			status := "NOT_PII"
			if result.IsLikelyName {
				status = "PII"
			}
			fmt.Printf("Line %d: %s (%.2f) - %s\n", i+1, status, result.Confidence, line)
		}
	}

	fmt.Fprintf(os.Stderr, "\nSummary: %d processed, %d detected as PII (%.1f%%)\n", 
		processed, detected, float64(detected)/float64(processed)*100)
}

func outputJSON(result types.PIIResult) {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}
	fmt.Println(string(jsonBytes))
}

func outputHuman(result types.PIIResult, words []string) {
	input := strings.Join(words, " ")
	
	if result.IsLikelyName {
		fmt.Printf("✓ Likely PII name (%.1f%% confidence)\n", result.Confidence*100)
		fmt.Printf("  Input: %s\n", input)
		if len(result.Details.FirstNames) > 0 {
			fmt.Printf("  First names: %s\n", strings.Join(result.Details.FirstNames, ", "))
		}
		if len(result.Details.Surnames) > 0 {
			fmt.Printf("  Surnames: %s\n", strings.Join(result.Details.Surnames, ", "))
		}
		if result.Details.Pattern != "" {
			fmt.Printf("  Pattern: %s\n", result.Details.Pattern)
		}
		if result.Details.TopCountry != "" {
			fmt.Printf("  Most likely country: %s\n", result.Details.TopCountry)
		}
		if result.Details.Gender != "" {
			fmt.Printf("  Predicted gender: %s\n", result.Details.Gender)
		}
	} else {
		fmt.Printf("✗ Not a PII name (%.1f%% confidence)\n", result.Confidence*100)
		fmt.Printf("  Input: %s\n", input)
		if result.Details.Pattern != "" {
			fmt.Printf("  Reason: %s\n", result.Details.Pattern)
		}
	}
}

// Helper function to check if path exists
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func init() {
	// Try to find data file in common locations
	possiblePaths := []string{
		"data/combined_names.pb.gz",
		"../data/combined_names.pb.gz",
		"../../data/combined_names.pb.gz",
	}
	
	for _, path := range possiblePaths {
		if pathExists(path) {
			*dataPath = path
			break
		}
	}
}