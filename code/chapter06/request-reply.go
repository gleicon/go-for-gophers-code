package main

import (
	"fmt"
)

type Request struct {
	Payload string
	ReplyTo chan string
}

func responder(reqs <-chan Request) {
	for req := range reqs {
		go func(r Request) {
			result := fmt.Sprintf("Processed: %s", r.Payload)
			r.ReplyTo <- result
		}(req)
	}
}

func main() {
	reqs := make(chan Request)
	go responder(reqs)

	reply := make(chan string)
	reqs <- Request{Payload: "data", ReplyTo: reply}

	fmt.Println("Response:", <-reply)
}
