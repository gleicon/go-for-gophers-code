package main

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

var sampleTexts = []string{
	"Go is an open-source programming language.",
	"It makes it easy to build simple, reliable, and efficient software.",
	"Concurrency is not parallelism.",
	"Goroutines are lightweight threads managed by the Go runtime.",
	"Channels are pipes that connect concurrent goroutines.",
	"Select lets a goroutine wait on multiple communication operations.",
	"The standard library is one of Go's most valuable features.",
	"Interfaces are a way to define behavior.",
	"Error handling is explicit and simple.",
	"Go compiles to a single binary with no dependencies.",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// WordCount takes a slice of text and counts the total number of words.
func WordCount(texts []string) int {
	total := 0
	for _, text := range texts {
		words := strings.Fields(text)
		total += len(words)
	}
	return total
}

// Benchmark for WordCount
func BenchmarkWordCount(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		WordCount(sampleTexts)
	}
}

// Sub-benchmarks to see the difference with varying input sizes.
func BenchmarkWordCountVariations(b *testing.B) {
	// Small input
	b.Run("Small", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			WordCount(sampleTexts[:3])
		}
	})

	// Medium input
	b.Run("Medium", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			WordCount(sampleTexts[:6])
		}
	})

	// Large input
	b.Run("Large", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			WordCount(sampleTexts)
		}
	})
}

// Main function to demonstrate normal execution
func main() {
	fmt.Println("Word Count Demonstration:")
	fmt.Println("=========================")

	textSamples := generateTextSamples(5)
	for i, text := range textSamples {
		fmt.Printf("Sample %d: %s\n", i+1, text)
	}

	fmt.Println("\nTotal Words in Sample Texts:")
	fmt.Println(WordCount(textSamples))
}

// generateTextSamples creates a random slice of text samples
func generateTextSamples(n int) []string {
	var samples []string
	for i := 0; i < n; i++ {
		index := rand.Intn(len(sampleTexts))
		samples = append(samples, sampleTexts[index])
	}
	return samples
}

// To run the main function, use the command:
// go run benchmarktesting.go

// To run the benchmarks, use the command:
// go test -bench=. -benchmem
