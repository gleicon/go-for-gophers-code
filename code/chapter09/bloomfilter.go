package main

import (
	"fmt"
	"net/url"
	"strings"

	"math"

	"github.com/spaolacci/murmur3"
)

// BloomFilter represents a Bloom filter data structure
type BloomFilter struct {
	bitset []uint64 // Using uint64 for efficient bit operations
	size   uint     // Size of the bitset in bits
	k      uint     // Number of hash functions
}

// New creates a new Bloom filter optimized for expectedElements with falsePositiveRate
func NewBloomFilter(expectedElements int, falsePositiveRate float64) *BloomFilter {
	// Calculate optimal size and number of hash functions
	size := optimalBitSize(expectedElements, falsePositiveRate)
	k := optimalHashCount(size, expectedElements)

	// Create a bitset with enough uint64 elements
	bitsetSize := (size + 63) / 64 // Round up to nearest uint64
	return &BloomFilter{
		bitset: make([]uint64, bitsetSize),
		size:   size,
		k:      k,
	}
}

// optimalBitSize calculates the optimal size of the bitset
func optimalBitSize(n int, p float64) uint {
	return uint(math.Ceil(-float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
}

// optimalHashCount calculates the optimal number of hash functions
func optimalHashCount(size uint, n int) uint {
	return uint(math.Max(1, math.Round(float64(size)/float64(n)*math.Log(2))))
}

// Add adds an element to the Bloom filter
func (bf *BloomFilter) Add(data []byte) {
	for i := uint(0); i < bf.k; i++ {
		position := bf.getPosition(data, i)
		index, bit := position/64, position%64
		bf.bitset[index] |= 1 << bit
	}
}

// Contains checks if an element might be in the Bloom filter
func (bf *BloomFilter) Contains(data []byte) bool {
	for i := uint(0); i < bf.k; i++ {
		position := bf.getPosition(data, i)
		index, bit := position/64, position%64
		if bf.bitset[index]&(1<<bit) == 0 {
			return false
		}
	}
	return true
}

// getPosition calculates the bit position for a given element and hash function
func (bf *BloomFilter) getPosition(data []byte, hashNum uint) uint {
	// Create different hash functions using the seed value
	hash := murmur3.Sum64WithSeed(data, uint32(hashNum))
	return uint(hash % uint64(bf.size))
}

// Example usage of the Bloom filter
// This example demonstrates how to use the Bloom filter for a web crawler cache
// It normalizes URLs to ensure consistent representation and checks if a URL has been visited
// before marking it as visited. The Bloom filter is used to efficiently check for
// membership in the cache, allowing for a configurable false positive rate.
// The example also includes a function to normalize URLs by converting them to lowercase,
// removing trailing slashes, and stripping common tracking parameters.
// The Bloom filter is initialized with an expected number of URLs and a desired false positive rate.

// WebCrawlerCache uses a Bloom filter to remember visited URLs
type WebCrawlerCache struct {
	filter *BloomFilter
}

// NewWebCrawlerCache creates a new cache optimized for expectedURLs
func NewWebCrawlerCache(expectedURLs int) *WebCrawlerCache {
	// 0.01 = 1% false positive rate
	filter := NewBloomFilter(expectedURLs, 0.01)
	return &WebCrawlerCache{filter: filter}
}

// NormalizeURL normalizes URLs for consistent representation
func NormalizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Convert to lowercase
	u.Host = strings.ToLower(u.Host)
	u.Path = strings.ToLower(u.Path)

	// Remove trailing slash
	u.Path = strings.TrimSuffix(u.Path, "/")

	// Remove common tracking parameters
	q := u.Query()
	q.Del("utm_source")
	q.Del("utm_medium")
	q.Del("utm_campaign")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// HasVisited checks if a URL has been visited
func (wc *WebCrawlerCache) HasVisited(rawURL string) (bool, error) {
	normalized, err := NormalizeURL(rawURL)
	if err != nil {
		return false, err
	}

	return wc.filter.Contains([]byte(normalized)), nil
}

// MarkVisited marks a URL as visited
func (wc *WebCrawlerCache) MarkVisited(rawURL string) error {
	normalized, err := NormalizeURL(rawURL)
	if err != nil {
		return err
	}

	wc.filter.Add([]byte(normalized))
	return nil
}

func main() {
	// Create a cache expecting ~1 million URLs
	cache := NewWebCrawlerCache(1_000_000)

	// Simulate crawling
	urls := []string{
		"https://example.com/page1",
		"https://example.com/page2",
		"https://example.com/PAGE1", // Same as first URL after normalization
	}

	for _, u := range urls {
		visited, _ := cache.HasVisited(u)
		if !visited {
			fmt.Printf("Crawling: %s\n", u)
			cache.MarkVisited(u)
		} else {
			fmt.Printf("Skipping previously visited: %s\n", u)
		}
	}
}
