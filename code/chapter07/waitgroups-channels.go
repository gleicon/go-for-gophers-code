package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	jobs := make(chan int, 10)
	var wg sync.WaitGroup

	// Simulate some jobs
	for j := 1; j <= 10; j++ {
		jobs <- j
	}
	close(jobs) // Important: close channel so workers stop on range

	// Start a random number of workers (2 to 5)
	numWorkers := rand.Intn(4) + 2
	fmt.Println("Starting", numWorkers, "workers...")

	for i := 1; i <= numWorkers; i++ {
		// Dynamically add to WaitGroup *inside* goroutine (can be done before too)
		go func(id int) {
			wg.Add(1)
			defer wg.Done()

			for job := range jobs {
				fmt.Printf("Worker %d processing job %d\n", id, job)
				time.Sleep(time.Millisecond * 200)
			}
			fmt.Printf("Worker %d done\n", id)
		}(i)
	}

	wg.Wait() // blocks
	fmt.Println("All workers done.")
}
