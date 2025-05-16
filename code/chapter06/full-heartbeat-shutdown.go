package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	done := make(chan struct{})

	// Simulate shutdown after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		done <- struct{}{}
	}()

	count := 0
	for {
		select {
		case t := <-ticker.C:
			fmt.Println("Tick at", t)
			count++
			if count >= 5 {
				return // or: done <- struct{}{} to let goroutine exit
			}
		case <-done:
			fmt.Println("Received shutdown signal")
			return
		}
	}
}
