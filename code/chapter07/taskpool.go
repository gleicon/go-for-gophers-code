package main

import (
	"fmt"
	"sync"
)

func runPool(jobs []string, workers int) {
	var wg sync.WaitGroup
	jobsCh := make(chan string)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobsCh {
				process(id, job)
			}
		}(i)
	}

	for _, job := range jobs {
		jobsCh <- job
	}
	close(jobsCh) // Signal that no more jobs are coming
	wg.Wait()     // Wait for all workers to complete
}

func process(workerID int, job string) {
	fmt.Printf("Worker %d processing: %s\n", workerID, job)
}

func main() {
	jobs := []string{"job1", "job2", "job3", "job4", "job5"}
	workers := 3
	runPool(jobs, workers)
}
