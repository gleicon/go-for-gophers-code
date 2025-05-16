package main

import (
	"context"
	"fmt"
	"time"
)

func monitor(ctx context.Context, hb chan<- struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Monitor stopped")
			return
		case <-ticker.C:
			hb <- struct{}{}
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	hb := make(chan struct{})

	go monitor(ctx, hb)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Shutting down main")
			return
		case <-hb:
			fmt.Println("Heartbeat received")
		}
	}
}
