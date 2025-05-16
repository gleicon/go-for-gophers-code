package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/spaolacci/murmur3"
)

// CountMinSketch represents a Count-Min Sketch data structure
type CountMinSketch struct {
	matrix [][]uint32
	width  uint
	depth  uint
}

// New creates a new Count-Min Sketch with the specified error parameters
// epsilon: error in the count (ε)
// delta: probability of error (δ)
func NewCountMinSketch(epsilon, delta float64) *CountMinSketch {
	width := uint(math.Ceil(math.E / epsilon))
	depth := uint(math.Ceil(math.Log(1 / delta)))

	// Create and initialize the matrix
	matrix := make([][]uint32, depth)
	for i := uint(0); i < depth; i++ {
		matrix[i] = make([]uint32, width)
	}

	return &CountMinSketch{
		matrix: matrix,
		width:  width,
		depth:  depth,
	}
}

// Increment adds a count for the given data
func (cms *CountMinSketch) Increment(data []byte, count uint32) {
	for i := uint(0); i < cms.depth; i++ {
		position := cms.getPosition(data, i)
		cms.matrix[i][position] += count
	}
}

// Count estimates the count for the given data
func (cms *CountMinSketch) Count(data []byte) uint32 {
	var min uint32 = math.MaxUint32

	for i := uint(0); i < cms.depth; i++ {
		position := cms.getPosition(data, i)
		if cms.matrix[i][position] < min {
			min = cms.matrix[i][position]
		}
	}

	return min
}

// getPosition calculates the array position for a given element and hash function
func (cms *CountMinSketch) getPosition(data []byte, hashNum uint) uint {
	hash := murmur3.Sum64WithSeed(data, uint32(hashNum))
	return uint(hash % uint64(cms.width))
}

// Let's create a simple analytics tracker using Count-Min Sketch
// This will track search query frequencies and identify trending terms
// The threshold for heavy hitters is set to 5, meaning any term with a count of 5 or more
// will be considered a heavy hitter
// This is a simple implementation and can be extended with more features
// to include more sophisticated analytics, such as time-based trends or user segmentation

// SearchAnalytics tracks search query frequencies
type SearchAnalytics struct {
	sketch       *CountMinSketch
	heavyHitters map[string]uint32 // Store actual counts for potential heavy hitters
	threshold    uint32
}

// NewSearchAnalytics creates a new analytics tracker
func NewSearchAnalytics(errorRate, confidence float64, threshold uint32) *SearchAnalytics {
	return &SearchAnalytics{
		sketch:       NewCountMinSketch(errorRate, 1-confidence),
		heavyHitters: make(map[string]uint32),
		threshold:    threshold,
	}
}

// RecordQuery records a search query
func (sa *SearchAnalytics) RecordQuery(query string) {
	// Normalize the query
	query = strings.ToLower(strings.TrimSpace(query))

	// Skip empty queries
	if query == "" {
		return
	}

	// Update the sketch
	sa.sketch.Increment([]byte(query), 1)

	// Check if this might be a heavy hitter
	count := sa.sketch.Count([]byte(query))
	if count >= sa.threshold {
		// Keep exact count for potential heavy hitters
		sa.heavyHitters[query] = count
	}
}

// GetTrendingTerms returns the top N trending search terms
func (sa *SearchAnalytics) GetTrendingTerms(n int) []string {
	type queryCount struct {
		query string
		count uint32
	}

	// Convert map to slice for sorting
	counts := make([]queryCount, 0, len(sa.heavyHitters))
	for query, count := range sa.heavyHitters {
		counts = append(counts, queryCount{query, count})
	}

	// Sort by count (descending)
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	// Take top N
	result := make([]string, 0, n)
	for i := 0; i < n && i < len(counts); i++ {
		result = append(result, counts[i].query)
	}

	return result
}

func main() {
	// Create analytics with 0.01 error rate, 0.99 confidence, threshold of 5
	analytics := NewSearchAnalytics(0.01, 0.99, 5)

	// Simulate search queries
	queries := []string{
		"go programming", "probabilistic data structures",
		"go programming", "golang tutorial", "count min sketch",
		"go programming", "probabilistic data structures",
		"bloom filter example", "count min sketch",
		"go programming", "golang jobs", "probabilistic data structures",
		"count min sketch", "go programming", "golang tutorial",
	}

	for _, query := range queries {
		analytics.RecordQuery(query)
	}

	// Get top 3 trending terms
	trending := analytics.GetTrendingTerms(3)
	fmt.Println("Top trending search terms:")
	for i, term := range trending {
		count := analytics.sketch.Count([]byte(term))
		fmt.Printf("%d. %s (approx. %d times)\n", i+1, term, count)
	}
}
