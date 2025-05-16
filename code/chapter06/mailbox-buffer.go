package main

import (
	"context"
	"fmt"
	"time"
)

type Message struct {
	ID   string
	Data string
}

func mailbox(ctx context.Context, in <-chan Message, flush func([]Message)) {
	const maxBatch = 3
	var batch []Message
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				flush(batch)
			}
			return
		case msg := <-in:
			batch = append(batch, msg)
			if len(batch) >= maxBatch {
				flush(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				flush(batch)
				batch = nil
			}
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	in := make(chan Message)
	go mailbox(ctx, in, func(msgs []Message) {
		fmt.Println("Flushing batch:")
		for _, m := range msgs {
			fmt.Printf(" - %s: %s\n", m.ID, m.Data)
		}
	})

	for i := 1; i <= 7; i++ {
		in <- Message{ID: fmt.Sprintf("msg%d", i), Data: "payload"}
		time.Sleep(800 * time.Millisecond)
	}
}
