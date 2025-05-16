package main

import (
	"fmt"
	"sync"
	"time"
)

// Simulated shared configuration or connection
type Config struct {
	ConnectionString string
	Timestamp        time.Time
}

var (
	once   sync.Once
	config *Config
)

// InitConfig initializes the shared config safely and only once
func InitConfig() {
	once.Do(func() {
		fmt.Println("🔄 Initializing configuration...")
		time.Sleep(2 * time.Second) // Simulate expensive setup
		config = &Config{
			ConnectionString: "postgres://user:pass@localhost/db",
			Timestamp:        time.Now(),
		}
		fmt.Println("✅ Configuration initialized.")
	})
}

// GetConfig returns the shared config, initializing it if needed
func GetConfig() *Config {
	InitConfig()
	return config
}

func main() {
	var wg sync.WaitGroup

	// Simulate multiple goroutines needing the config at the same time
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cfg := GetConfig()
			fmt.Printf("[Goroutine %d] Got config: %s (created at %v)\n",
				id, cfg.ConnectionString, cfg.Timestamp.Format(time.Stamp))
		}(i)
	}

	wg.Wait()
}
