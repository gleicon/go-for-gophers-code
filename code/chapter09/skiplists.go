package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	maxLevel = 16   // Maximum level for the skip list
	p        = 0.25 // Probability of inserting at higher level
)

// Initialize the random number generator
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

// Node represents a node in the skip list
type Node[K comparable, V any] struct {
	key     K
	value   V
	forward []*Node[K, V] // Array of pointers for each level
}

// SkipList is a generic skip list implementation
type SkipList[K comparable, V any] struct {
	head  *Node[K, V]     // Head node (sentinel)
	level int             // Current maximum level
	less  func(K, K) bool // Comparison function
}

// New creates a new skip list with the specified comparison function
func NewSkipList[K comparable, V any](less func(K, K) bool) *SkipList[K, V] {
	var zeroK K
	var zeroV V

	head := &Node[K, V]{
		key:     zeroK,
		value:   zeroV,
		forward: make([]*Node[K, V], maxLevel),
	}

	return &SkipList[K, V]{
		head:  head,
		level: 1,
		less:  less,
	}
}

// randomLevel determines a random level for a new node
func randomLevel() int {
	lvl := 1
	for rnd.Float64() < p && lvl < maxLevel {
		lvl++
	}
	return lvl
}

// Insert adds or updates a key-value pair
func (sl *SkipList[K, V]) Insert(key K, value V) {
	// Create update array and initialize it
	update := make([]*Node[K, V], maxLevel)
	current := sl.head

	// Find position to insert
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && sl.less(current.forward[i].key, key) {
			current = current.forward[i]
		}
		update[i] = current
	}

	// Get first element that might be equal to our key
	current = current.forward[0]

	// Update existing node if key exists
	if current != nil && !sl.less(current.key, key) && !sl.less(key, current.key) {
		current.value = value
		return
	}

	// Otherwise, create new node with random level
	level := randomLevel()

	// Update the skip list level if necessary
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			update[i] = sl.head
		}
		sl.level = level
	}

	// Create new node
	newNode := &Node[K, V]{
		key:     key,
		value:   value,
		forward: make([]*Node[K, V], level),
	}

	// Insert the node at all levels
	for i := 0; i < level; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}
}

// Search looks for a key and returns its value and success flag
func (sl *SkipList[K, V]) Search(key K) (V, bool) {
	var zeroV V
	current := sl.head

	// Start from the highest level and work down
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && sl.less(current.forward[i].key, key) {
			current = current.forward[i]
		}
	}

	// Move to the actual element
	current = current.forward[0]

	// Check if we found the key
	if current != nil && !sl.less(current.key, key) && !sl.less(key, current.key) {
		return current.value, true
	}

	return zeroV, false
}

// Delete removes a key from the skip list
func (sl *SkipList[K, V]) Delete(key K) bool {
	update := make([]*Node[K, V], maxLevel)
	current := sl.head

	// Find the node to delete
	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && sl.less(current.forward[i].key, key) {
			current = current.forward[i]
		}
		update[i] = current
	}

	current = current.forward[0]

	// If found, remove it from all levels
	if current != nil && !sl.less(current.key, key) && !sl.less(key, current.key) {
		for i := 0; i < sl.level; i++ {
			if update[i].forward[i] != current {
				break
			}
			update[i].forward[i] = current.forward[i]
		}

		// Update the level if needed
		for sl.level > 1 && sl.head.forward[sl.level-1] == nil {
			sl.level--
		}

		return true
	}

	return false
}

// Example usage
// This example demonstrates a simple time-to-live (TTL) cache using a skip list
// with a cleanup mechanism to remove expired items.

// CacheItem represents a value in the cache with expiration time
type CacheItem struct {
	value      interface{}
	expiration time.Time
}

// TTLCache is a time-to-live cache using a skip list for efficient access
type TTLCache struct {
	items       *SkipList[string, CacheItem]
	defaultTTL  time.Duration
	cleanupFreq time.Duration
	stopCleanup chan struct{}
}

// NewTTLCache creates a new cache with default TTL and cleanup frequency
func NewTTLCache(defaultTTL, cleanupFreq time.Duration) *TTLCache {
	cache := &TTLCache{
		items:       NewSkipList[string, CacheItem](func(a, b string) bool { return a < b }),
		defaultTTL:  defaultTTL,
		cleanupFreq: cleanupFreq,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Set adds or updates a key with the default TTL
func (c *TTLCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL adds or updates a key with a specific TTL
func (c *TTLCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	expiration := time.Now().Add(ttl)
	item := CacheItem{
		value:      value,
		expiration: expiration,
	}
	c.items.Insert(key, item)
}

// Get retrieves a value from the cache
func (c *TTLCache) Get(key string) (interface{}, bool) {
	item, found := c.items.Search(key)
	if !found {
		return nil, false
	}

	// Check if the item has expired
	if time.Now().After(item.expiration) {
		c.items.Delete(key)
		return nil, false
	}

	return item.value, true
}

// Delete removes a key from the cache
func (c *TTLCache) Delete(key string) {
	c.items.Delete(key)
}

// cleanupLoop periodically removes expired items
func (c *TTLCache) cleanupLoop() {
	ticker := time.NewTicker(c.cleanupFreq)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanup removes all expired items
func (c *TTLCache) cleanup() {
	//now := time.Now()

	// This is a simplified approach - in a real implementation,
	// we would use the skip list more efficiently
	keysToDelete := []string{}

	// Iterate through all keys to find expired ones
	// This would be implemented with a proper iterator in a real skiplist

	for _, key := range keysToDelete {
		c.items.Delete(key)
	}
}

// Close stops the cleanup goroutine
func (c *TTLCache) Close() {
	close(c.stopCleanup)
}

func main() {
	// Create a cache with 1 minute default TTL, cleanup every 10 seconds
	cache := NewTTLCache(1*time.Minute, 10*time.Second)
	defer cache.Close()

	// Add some items
	cache.Set("user:1001", map[string]string{"name": "Alice", "role": "admin"})
	cache.Set("user:1002", map[string]string{"name": "Bob", "role": "user"})
	cache.SetWithTTL("session:abc123", "token-data", 30*time.Second)

	// Retrieve and use the data
	key := "user:1001"
	if userData, found := cache.Get(key); found {
		fmt.Printf("Found key: %s user: %v\n", key, userData)
	}

	// Wait for the short TTL item to expire
	fmt.Println("Waiting item expiration")

	time.Sleep(35 * time.Second)

	key = "session:abc123"
	if ud, found := cache.Get(key); !found {
		fmt.Println("Session expired as expected")
	} else {
		fmt.Printf("oops, found %s user: %v\n", key, ud)
	}
}
