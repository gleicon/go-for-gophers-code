package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	tasks := []string{"task1", "task2", "task3", "task4", "task5"}

	// Set a timeout shorter than total work time to trigger cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel() // Important to release resources

	fmt.Println("Running tasks with context cancellation...")
	runWithContext(ctx, tasks)
	fmt.Println("All done.")
}

func process(task string) {
	delay := time.Duration(200+rand.Intn(400)) * time.Millisecond
	fmt.Printf("🛠️  %s started, will take %v\n", task, delay)
	time.Sleep(delay)
	fmt.Printf("✅ %s done\n", task)
}

func runWithContext(ctx context.Context, tasks []string) {
	var wg sync.WaitGroup

	for _, task := range tasks {
		task := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				log.Printf("⚠️ Task %s canceled: %v\n", task, ctx.Err())
			default:
				process(task)
			}
		}()
	}

	wg.Wait()
}
