package main

import (
	"fmt"
	"sync"
	"time"
)

// Simulated file contents
var fakeFiles = map[string]string{
	"file1.txt": "Hello from file 1",
	"file2.txt": "Greetings from file 2",
	"file3.txt": "This is file 3",
	"file4.txt": "Data from file 4",
	"file5.txt": "Another one: file 5",
}

// Indexed Fan-Out/Fan-In for File Reads
func readFiles(paths []string) []string {
	var wg sync.WaitGroup
	results := make([]string, len(paths)) // Preserves order

	for i, path := range paths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			fmt.Printf("[Worker %d] Reading %s...\n", i+1, path)
			time.Sleep(100 * time.Millisecond) // Simulate file read delay

			content, ok := fakeFiles[path]
			if ok {
				results[i] = content
				fmt.Printf("[Worker %d] Finished %s: %s\n", i+1, path, content)
			} else {
				results[i] = "error: file not found"
				fmt.Printf("[Worker %d] Failed %s: file not found\n", i+1, path)
			}
		}(i, path)
	}

	wg.Wait()
	return results
}

func main() {
	paths := []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt"}
	fmt.Println("Reading files in parallel...")
	results := readFiles(paths)
	fmt.Println("\nFinal Results:")
	for i, r := range results {
		fmt.Printf("%s -> %s\n", paths[i], r)
	}
}
