package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	tasks := []string{"task1", "task2", "task3", "task4", "task5"}

	fmt.Println("Running tasks with error handling...")
	if err := runWithErrors(tasks); err != nil {
		fmt.Printf("Finished with error: %v\n", err)
	} else {
		fmt.Println("All tasks completed successfully.")
	}
}

// Simulated task processor with random failure
func processTask(task string) error {
	fmt.Printf("Processing %s...\n", task)
	time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)

	if rand.Float32() < 0.3 { // 30% chance to fail
		fmt.Printf("%s failed\n", task)
		return errors.New("failed: " + task)
	}

	fmt.Printf("%s succeeded\n", task)
	return nil
}

// Core pattern: WaitGroup + error channel
func runWithErrors(tasks []string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(tasks)) // Buffered to avoid blocking

	for _, task := range tasks {
		task := task // Capture range variable
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := processTask(task); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	// Return the first error encountered, if any
	for err := range errCh {
		return err
	}
	return nil
}
