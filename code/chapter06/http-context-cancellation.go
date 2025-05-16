package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type key string

const traceIDKey key = "traceID"

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract metadata from headers (e.g., trace ID)
	traceID := r.Header.Get("X-Trace-ID")
	if traceID == "" {
		traceID = "unknown"
	}

	// Store trace ID in context
	ctx = context.WithValue(ctx, traceIDKey, traceID)

	// Simulate work with cancellation awareness
	select {
	case <-time.After(3 * time.Second):
		// Simulated long-running operation
		id := ctx.Value(traceIDKey).(string)
		fmt.Fprintf(w, "Processed request. Trace ID: %s\n", id)
	case <-ctx.Done():
		// Request was canceled by the client
		log.Println("Request canceled:", ctx.Err())
		http.Error(w, "Request canceled", http.StatusRequestTimeout)
	}
}

/*
Test it with:
$ curl -H "X-Trace-ID: abc123" localhost:8080

Abort the request mid-way to trigger cancellation:
$ curl -m 1 localhost:8080
*/

func main() {
	http.HandleFunc("/", handler)

	srv := &http.Server{
		Addr: ":8080",
	}

	log.Println("Starting server on :8080")
	log.Fatal(srv.ListenAndServe())
}
