package main

import (
	"fmt"
	"sync"
	"time"
)

// fanOutFanIn distributes work across multiple workers and collects results
func fanOutFanIn(jobs []int, workerCount int) []int {
	jobCh := make(chan int)    // Channel to send jobs
	resultCh := make(chan int) // Channel to collect results
	var wg sync.WaitGroup

	// Fan-Out: Start workers
	for i := 0; i < workerCount; i++ {
		id := i + 1 // Worker IDs start at 1
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobCh {
				fmt.Printf("[Worker %d] Processing job: %d\n", workerID, job)
				time.Sleep(100 * time.Millisecond) // Simulate work
				result := job * 2
				fmt.Printf("[Worker %d] Finished job: %d -> %d\n", workerID, job, result)
				resultCh <- result
			}
		}(id)
	}

	// Feed jobs to workers (Fan-Out)
	go func() {
		for _, job := range jobs {
			jobCh <- job
		}
		close(jobCh) // Important to close, otherwise workers will hang
	}()

	// Close the result channel once all workers are done
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results from all workers (Fan-In)
	var results []int
	for res := range resultCh {
		results = append(results, res)
	}

	return results
}

func main() {
	jobs := make([]int, 20)
	for i := 0; i < len(jobs); i++ {
		jobs[i] = i + 1
	}

	workerCount := 4
	fmt.Println("Input:", jobs)
	results := fanOutFanIn(jobs, workerCount)
	fmt.Println("Output:", results)
}
