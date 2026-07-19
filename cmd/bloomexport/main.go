// bloomexport builds a compact Bloom filter of popular names from the
// embedded dataset, for client-side membership checks (e.g. validating
// NER PERSON spans in the browser without any data leaving the device).
//
// Output format ("NBF1"):
//
//	bytes 0-3   magic "NBF1"
//	bytes 4-7   uint32 LE  k (number of probes)
//	bytes 8-15  uint64 LE  m (filter size in bits)
//	bytes 16-   bit array, LSB-first within each byte
//
// Lookup contract (identical on any client):
//
//	key  = lowercase(trim(NFD(name) minus combining marks))
//	h    = FNV-1a 32 over the UTF-8 bytes of key
//	h1   = fmix32(h)
//	h2   = fmix32(h ^ 0x5bd1e995) | 1
//	bits = (h1 + i*h2) mod m, for i in [0, k)
//
// fmix32 is the murmur3 finalizer; without it the two probe hashes are
// correlated and the false-positive rate degrades by an order of
// magnitude.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"unicode"

	"github.com/montevive/go-name-detector/pkg/loader"
	"github.com/montevive/go-name-detector/pkg/types"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Countries relevant to Spanish/European deployments; override with -countries.
const defaultCountries = "ES,MX,AR,CO,CL,PE,VE,EC,UY,PY,BO,CR,PA,DO,GT,HN,SV,NI,CU,US,GB,IE,FR,DE,IT,PT,BR,NL,BE,CH,AT,RO,PL,MA"

func main() {
	var (
		countriesCSV = flag.String("countries", defaultCountries, "comma-separated country codes whose rankings count")
		maxRank      = flag.Int("max-rank", 2000, "keep names ranked at or above this in at least one selected country (0 = keep all)")
		bitsPerEntry = flag.Float64("bits-per-entry", 11, "filter bits per entry (~11 bits + k=8 gives ~0.5% false positives)")
		k            = flag.Uint("k", 8, "number of probe hashes")
		names        = flag.String("names", "first", "which dataset to export: first or last")
		out          = flag.String("out", "names.bloom", "output file path")
	)
	flag.Parse()

	l, err := loader.NewWithEmbeddedData()
	if err != nil {
		fatal("loading embedded dataset: %v", err)
	}
	ds := l.GetDataset()

	source := ds.FirstNames
	if *names == "last" {
		source = ds.LastNames
	} else if *names != "first" {
		fatal("-names must be first or last")
	}

	countries := map[string]bool{}
	for _, c := range strings.Split(*countriesCSV, ",") {
		countries[strings.ToUpper(strings.TrimSpace(c))] = true
	}

	keys := selectNames(source, countries, int32(*maxRank))
	if len(keys) == 0 {
		fatal("no names selected — check -countries / -max-rank")
	}

	filter := newBloom(len(keys), *bitsPerEntry, uint32(*k))
	for key := range keys {
		filter.add(key)
	}
	if err := filter.write(*out); err != nil {
		fatal("writing %s: %v", *out, err)
	}
	fmt.Printf("selected %d names -> %s (%d bytes, k=%d, m=%d bits, ~%.2f%% expected FP)\n",
		len(keys), *out, 16+len(filter.bits), filter.k, filter.m, expectedFP(len(keys), filter)*100)
}

func selectNames(source map[string]*types.NameData, countries map[string]bool, maxRank int32) map[string]bool {
	keys := map[string]bool{}
	for name, data := range source {
		key := fold(name)
		if len(key) < 2 {
			continue
		}
		if maxRank <= 0 {
			keys[key] = true
			continue
		}
		for country, rank := range data.Rank {
			if countries[country] && rank > 0 && rank <= maxRank {
				keys[key] = true
				break
			}
		}
	}
	return keys
}

func fold(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, err := transform.String(t, s)
	if err != nil {
		out = s
	}
	return strings.ToLower(strings.TrimSpace(out))
}

func fnv1a32(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func fmix32(h uint32) uint32 {
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

type bloom struct {
	k    uint32
	m    uint64
	bits []byte
}

func newBloom(n int, bitsPerEntry float64, k uint32) *bloom {
	m := uint64(math.Ceil(float64(n) * bitsPerEntry))
	return &bloom{k: k, m: m, bits: make([]byte, (m+7)/8)}
}

func (b *bloom) add(s string) {
	h := fnv1a32(s)
	h1 := fmix32(h)
	h2 := fmix32(h^0x5bd1e995) | 1
	for i := uint32(0); i < b.k; i++ {
		bit := (uint64(h1) + uint64(i)*uint64(h2)) % b.m
		b.bits[bit/8] |= 1 << (bit % 8)
	}
}

func (b *bloom) write(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString("NBF1"); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, b.k); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, b.m); err != nil {
		return err
	}
	_, err = f.Write(b.bits)
	return err
}

func expectedFP(n int, b *bloom) float64 {
	fill := 1 - math.Exp(-float64(uint64(b.k)*uint64(n))/float64(b.m))
	return math.Pow(fill, float64(b.k))
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
