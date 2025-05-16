package main

import (
	"fmt"
	"sync"
	"time"
)

type SafeMap struct {
	mu sync.RWMutex
	m  map[string]string
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		m: make(map[string]string),
	}
}

func (s *SafeMap) Get(key string) (string, bool) {
	s.mu.RLock() // Allows multiple readers
	defer s.mu.RUnlock()
	val, ok := s.m[key]
	return val, ok
}

func (s *SafeMap) Set(key, value string) {
	s.mu.Lock() // Allows only one writer
	defer s.mu.Unlock()
	s.m[key] = value
}

func main() {
	sm := NewSafeMap()
	var wg sync.WaitGroup

	// Write to the map
	wg.Add(1)
	go func() {
		defer wg.Done()
		sm.Set("language", "Go")
	}()

	// Read from the map
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // Simulate slight delay
		if val, ok := sm.Get("language"); ok {
			fmt.Println("Read from SafeMap:", val)
		} else {
			fmt.Println("Key not found")
		}
	}()

	wg.Wait()
}
