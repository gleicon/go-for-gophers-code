package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context, out chan<- int) {
	defer close(out)
	for i := 0; i < 100; i++ {
		select {
		case <-ctx.Done():
			return
		case out <- i:
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := make(chan int)
	go worker(ctx, ch)

	for val := range ch {
		fmt.Println("Received:", val)
	}
}
