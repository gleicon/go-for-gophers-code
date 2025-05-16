package main

import (
	"fmt"
	"sync"
)

func mapReduce(inputs []int, mapper func(int) int, reducer func([]int) int) int {
	var wg sync.WaitGroup
	results := make(chan int, len(inputs))

	// Map phase: Launch a goroutine for each input to compute partial results
	for _, input := range inputs {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			results <- mapper(val)
		}(input)
	}

	// Wait for all mappers to finish and close the channel
	wg.Wait()
	close(results)

	// Collect all results for the reduce phase
	var mapped []int
	for r := range results {
		mapped = append(mapped, r)
	}

	// Reduce phase: Aggregate all mapped results
	return reducer(mapped)
}

func square(n int) int { return n * n }

func sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func main() {
	inputs := []int{1, 2, 3, 4, 5}
	result := mapReduce(inputs, square, sum)
	fmt.Println("Sum of squares:", result)
}
