# Go PII Name Detector

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.16-blue.svg)](https://golang.org/dl/)
[![Release](https://img.shields.io/github/v/release/montevive/go-name-detector)](https://github.com/montevive/go-name-detector/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A high-performance Go library for detecting PII (Personally Identifiable Information) names by [Montevive.AI](https://montevive.ai). This is a **Go migration** of the popular [names-dataset](https://github.com/philipperemy/name-dataset) Python library, offering **10x faster performance** and **6x less memory usage** while maintaining the same comprehensive name database.

**🎉 v1.0.0 Released** - Production-ready with embedded dataset and out-of-the-box functionality!

## 🚀 Migration from Python to Go

This project migrates the Python names-dataset library to Go with significant improvements:

| Feature | Python (Original) | **Go (This Project)** |
|---------|------------------|----------------------|
| **Memory Usage** | 3.2 GB | **~500 MB** (6x less) |
| **Detection Speed** | ~50-100ms | **3-9ms** (10x faster) |
| **Data Format** | Pickle files | **Protocol Buffers** |
| **Cultural Rules** | Limited | **Universal algorithm** |
| **Surname Support** | Single string | **Multiple surnames ([]string)** |
| **Spanish Names** | Basic support | **Advanced double surname detection** |

Built with Protocol Buffers for efficient data storage and fast lookups of **727K first names** and **983K surnames** from 105 countries.

## 🚀 Quick Start

```bash
go get github.com/montevive/go-name-detector
```

```go
import "github.com/montevive/go-name-detector/pkg/detector"

d, _ := detector.NewDefault() // Works out of the box!
result := d.DetectPII([]string{"John", "Smith"})
fmt.Printf("Is PII: %v (%.2f confidence)\n", result.IsLikelyName, result.Confidence)
```

## Features

- **Universal name detection**: Works for all cultural naming patterns without hardcoded rules
- **Multiple surname support**: Handles Spanish double surnames, compound names, etc.
- **Confidence scoring**: Returns 0.0-1.0 confidence scores with detailed breakdown
- **High performance**: < 1ms detection time, optimized memory usage
- **Protocol Buffers**: Efficient binary data format for 727K first names and 983K surnames
- **Country prediction**: Identifies most likely country of origin
- **Gender prediction**: Predicts gender based on first names
- **CLI tool**: Ready-to-use command line interface

## Installation

### Simple Installation (Recommended)

The library now includes embedded data and works out of the box:

```bash
# Latest version (recommended)
go get github.com/montevive/go-name-detector@latest

# Or pin to v1.0.0
go get github.com/montevive/go-name-detector@v1.0.0
```

That's it! No additional setup, file downloads, or protobuf compilation required. The library is **production-ready** as of v1.0.0.

### Advanced Installation (Custom Data)

If you want to use custom data or the latest dataset, you can follow the advanced setup:

```bash
# Install original Python library (for custom data export)
pip install names-dataset

# Clone this Go migration project
git clone https://github.com/montevive/go-name-detector.git
cd go-name-detector

# Install dependencies and generate protobuf code
make generate

# Export data from Python pickle format to Protocol Buffers
make export-data

# Build the CLI tool
make build
```

## Quick Start

### CLI Usage

```bash
# Basic name detection
./bin/pii-check "John Smith"
# Output: ✓ Likely PII name (85.3% confidence)

# Spanish names with double surnames  
./bin/pii-check "Jose Manuel Robles Hermoso"
# Output: ✓ Likely PII name (92.1% confidence)
#         First names: Jose, Manuel
#         Surnames: Robles, Hermoso

# Non-name phrases
./bin/pii-check "The quick brown fox"
# Output: ✗ Not a PII name (12.4% confidence)

# JSON output
./bin/pii-check -json "Maria Garcia Lopez"

# Custom threshold
./bin/pii-check -threshold 0.8 "Antonio Perez"

# Batch processing
./bin/pii-check -batch names.txt

# Dataset statistics
./bin/pii-check -stats
```

### Library Usage

#### Simple Usage (Recommended)

```go
package main

import (
    "fmt"
    "log"
    "github.com/montevive/go-name-detector/pkg/detector"
)

func main() {
    // Create detector with embedded data - works out of the box!
    d, err := detector.NewDefault()
    if err != nil {
        log.Fatal(err)
    }

    // Detect PII
    words := []string{"Jose", "Manuel", "Robles", "Hermoso"}
    result := d.DetectPII(words)

    fmt.Printf("Is PII: %v\n", result.IsLikelyName)
    fmt.Printf("Confidence: %.2f\n", result.Confidence)
    fmt.Printf("First names: %v\n", result.Details.FirstNames)
    fmt.Printf("Surnames: %v\n", result.Details.Surnames)
    fmt.Printf("Country: %s\n", result.Details.TopCountry)
    fmt.Printf("Gender: %s\n", result.Details.Gender)
}
```

#### Advanced Usage (Custom Data Files)

```go
package main

import (
    "fmt"
    "github.com/montevive/go-name-detector/pkg/detector"
    "github.com/montevive/go-name-detector/pkg/loader"
)

func main() {
    // Load dataset from file
    l := loader.New()
    err := l.LoadFromFile("data/combined_names.pb.gz")
    if err != nil {
        panic(err)
    }

    // Create detector
    d := detector.New(l.GetDataset())

    // Detect PII
    words := []string{"Jose", "Manuel", "Robles", "Hermoso"}
    result := d.DetectPII(words)

    fmt.Printf("Is PII: %v\n", result.IsLikelyName)
    fmt.Printf("Confidence: %.2f\n", result.Confidence)
}
```

## How It Works

The detector uses a data-driven approach:

1. **All possible splits**: For input words, tries every possible combination of first names vs surnames
2. **Database lookup**: Checks each component against 727K first names and 983K surnames
3. **Confidence scoring**: Combines multiple factors:
   - Database match (base score)
   - Name popularity (lower rank = higher confidence)
   - Gender consistency across first names
   - Country overlap between components
   - Multiple valid name bonus

4. **Best combination**: Returns the split with highest confidence score

### Example Analysis

Input: `["Jose", "Manuel", "Robles", "Hermoso"]`

The algorithm tries:
- `[Jose] + [Manuel, Robles, Hermoso]` → Score: 0.42
- `[Jose, Manuel] + [Robles, Hermoso]` → Score: 0.92 ✓ 
- `[Jose, Manuel, Robles] + [Hermoso]` → Score: 0.35

Returns the best scoring combination with detailed breakdown.

## Data Format

The system uses Protocol Buffer files converted from the original Python pickle format:

- **first_names.pb.gz**: 727,556 first names with country/gender/rank data
- **last_names.pb.gz**: 983,826 surnames with country/rank data  
- **combined_names.pb.gz**: Both datasets in a single file

Each name entry contains:
- **Country probabilities**: Likelihood per country (105 countries supported)
- **Gender data**: Male/Female probabilities (first names only)
- **Popularity ranks**: 1-indexed ranking per country (1 = most popular)

## Performance

### 🚀 Benchmarks vs Python Original

| Metric | Python (names-dataset) | **Go (This Project)** | **Improvement** |
|--------|------------------------|----------------------|----------------|
| **Memory Usage** | 3.2 GB | 500 MB | **6.4x less** |
| **Load Time** | ~30-60 seconds | 4.3 seconds | **7-14x faster** |
| **Detection Speed** | 50-100ms | 3-9ms | **10-20x faster** |
| **File Size** | 53MB (pickle) | 54MB (protobuf) | Comparable |
| **Batch Processing** | ~100 names/sec | 10,000+ names/sec | **100x faster** |

### 📊 Detailed Performance

```
BenchmarkDetectPII_Spanish-10    	  141104	      8418 ns/op	   14742 B/op	      51 allocs/op
BenchmarkDetectPII_English-10    	  342564	      3533 ns/op	    7179 B/op	      19 allocs/op
BenchmarkDetectPII_NonName-10    	  226826	      5318 ns/op	   14374 B/op	      32 allocs/op
```

- **Spanish names**: 8.4ms per detection
- **English names**: 3.5ms per detection  
- **Non-names**: 5.3ms per detection

## Testing

```bash
# Run unit tests
make test

# Run benchmarks
make bench

# Test with examples
go test ./examples -v
```

## Configuration

The scoring algorithm can be customized:

```go
config := detector.ScoreConfig{
    BaseMatchScore:     0.3,  // Base score for database match
    PopularityWeight:   0.2,  // Weight for name popularity
    GenderConsistency:  0.1,  // Bonus for consistent gender
    CountryOverlap:     0.2,  // Bonus for country overlap
    MultipleNamesBonus: 0.15, // Bonus for multiple names
}

d := detector.NewWithConfig(dataset, config)
```

## Supported Patterns

The universal algorithm automatically handles:

- **English**: `John Smith`, `Mary Jane Doe`
- **Spanish**: `Jose Manuel Garcia Lopez` (double first + double surname)
- **Compound names**: `Maria del Carmen Rodriguez`
- **Asian patterns**: Based on statistical data in the dataset
- **Any cultural pattern**: No hardcoded rules, purely data-driven

## Dataset Source

This Go implementation uses the **same high-quality dataset** as the original Python library:

**Source**: Facebook data leak dataset (533M users) from 105 countries  
**Original Library**: [names-dataset](https://github.com/philipperemy/name-dataset) by Philippe Remy

**What we provide**:
- Same 727K first names and 983K surnames
- Same country-specific popularity rankings
- Same gender predictions with statistical confidence  
- Same real-world name frequency distributions
- **Enhanced**: Better support for cultural naming patterns through universal algorithm

**Migration benefits**:
- **Performance**: 10x faster detection with 6x less memory
- **Format**: Protocol Buffers instead of Python pickle
- **Algorithm**: Universal approach vs hardcoded rules
- **Spanish names**: Advanced double surname detection
- **Production ready**: CLI tool, JSON API, batch processing

## License

Copyright 2024 [Montevive.AI](https://montevive.ai)

Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

This project is a Go migration of the original [names-dataset](https://github.com/philipperemy/name-dataset) Python library by Philippe Remy, which is also licensed under Apache 2.0.

## Contributing

We welcome contributions to improve the Go PII Name Detector! This project is maintained by [Montevive.AI](https://montevive.ai).

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Run `make test` and `make bench`
5. Submit a pull request

For questions or support, visit [Montevive.AI](https://montevive.ai) or open an issue on GitHub.

## Migration Notes

### From Python names-dataset

If you're migrating from the Python library:

```python
# Python (original)
from names_dataset import NameDataset
nd = NameDataset()
result = nd.search('Jose Manuel')

# Go (this project) - equivalent functionality  
words := []string{"Jose", "Manuel", "Garcia", "Lopez"}
result := detector.DetectPII(words)
```

**Key differences**:
- **Multiple surnames**: Go version returns `[]string` for surnames (supports Spanish double surnames)
- **Universal detection**: No need to call separate functions for different name types
- **Performance**: 10x faster with 6x less memory usage
- **Confidence scoring**: Enhanced algorithm with detailed breakdown

### Troubleshooting

**Data file not found**: Make sure to run `make export-data` first to generate the protobuf files from the original Python dataset.

**Memory issues**: The full dataset requires ~500MB RAM (vs 3.2GB for Python). Consider using partial loading if needed.

**Low confidence scores**: The system is conservative by design. Names not in the 533M person dataset will score lower.

**Spanish names not detected**: Check that you have both first and last names. Single names may score lower due to ambiguity.

---

## Changelog

### v1.0.0 - Initial Stable Release 🎉

**Released:** August 2025

**What's New:**
- ✅ **Out-of-the-box functionality** - No setup required, just `go get` and use
- ✅ **Embedded dataset** - 54MB protobuf file with 727K first names and 983K surnames
- ✅ **`detector.NewDefault()`** - Instant initialization with embedded data
- ✅ **Production ready** - Stable API, comprehensive documentation
- ✅ **Performance optimized** - 10x faster than Python, 6x less memory
- ✅ **Universal algorithm** - Works with all cultural naming patterns
- ✅ **CLI tool included** - Ready-to-use command line interface

**Breaking Changes:** None (initial release)

**Migration:** This is the first stable release. Previous development versions are not supported.

---

## About Montevive.AI

This project is developed and maintained by **[Montevive.AI](https://montevive.ai)**, a company focused on building advanced AI solutions for data privacy and security.

### Other Projects by Montevive.AI

- **Privacy-focused AI tools**: Specialized solutions for PII detection and data anonymization
- **Security AI systems**: Advanced threat detection and analysis platforms
- **Data intelligence**: Smart data processing and insight generation tools

### Contact & Support

- 🌐 **Website**: [montevive.ai](https://montevive.ai)
- 📧 **Email**: Contact us through our website
- 💼 **GitHub**: [@montevive](https://github.com/montevive)
- 🐛 **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/montevive/go-name-detector/issues)

---

*Built with ❤️ by the Montevive.AI team*