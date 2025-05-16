package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

var reqID int64

func handler(w http.ResponseWriter, r *http.Request) {
	id := atomic.AddInt64(&reqID, 1)

	fmt.Printf("[#%d] Start\n", id)

	time.Sleep(2 * time.Second) // Simulate some work

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	fmt.Fprintf(w, "Request #%d handled\n", id)
	fmt.Printf("[#%d] Done — Goroutines: %d, Alloc: %.2fMB\n",
		id,
		runtime.NumGoroutine(),
		float64(mem.Alloc)/1024/1024,
	)
}

// run with: go run concurrent-http-server.go
// Then open a browser and go to http://localhost:8080 or
// use curl to test the server:
// $ curl http://localhost:8080 &
// $ curl http://localhost:8080 &
// $ curl http://localhost:8080 &
func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
