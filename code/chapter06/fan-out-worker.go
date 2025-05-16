package main

import (
	"fmt"
	"time"
)

func worker(id int, jobs <-chan int) {
	for job := range jobs {
		fmt.Printf("Worker %d processing job %d\n", id, job)
		time.Sleep(500 * time.Millisecond) // Simulate work
	}
}

func main() {
	jobs := make(chan int)

	// Start 3 workers (fan-out)
	for w := 1; w <= 3; w++ {
		go worker(w, jobs)
	}

	// Send 5 jobs
	for j := 1; j <= 5; j++ {
		jobs <- j
	}

	close(jobs)                 // Close channel so workers exit after processing
	time.Sleep(2 * time.Second) // Wait for all workers to finish
}
