package main

import (
	"container/list"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type entry struct{ key, val string }

type LRUCache struct {
	cap  int
	list *list.List
	data map[string]*list.Element
	mu   sync.Mutex
}

func NewLRU(cap int) *LRUCache {
	return &LRUCache{
		cap:  cap,
		list: list.New(),
		data: make(map[string]*list.Element),
	}
}

func (c *LRUCache) Get(k string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.data[k]; ok {
		c.list.MoveToFront(e)
		return e.Value.(entry).val, true
	}
	return "", false
}

func (c *LRUCache) Set(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.data[k]; ok {
		c.list.MoveToFront(e)
		e.Value = entry{k, v}
		return
	}
	if c.list.Len() >= c.cap {
		old := c.list.Back()
		c.list.Remove(old)
		delete(c.data, old.Value.(entry).key)
	}
	e := c.list.PushFront(entry{k, v})
	c.data[k] = e
}

type LRUSQLiteBackend struct {
	cache *LRUCache
	db    *sql.DB
}

func NewLRUSQLiteBackend(dbPath string, cacheSize int) *LRUSQLiteBackend {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS kv (key TEXT PRIMARY KEY, val TEXT)`)
	return &LRUSQLiteBackend{
		cache: NewLRU(cacheSize),
		db:    db,
	}
}

func (s *LRUSQLiteBackend) Get(k string) (string, error) {
	// First check the cache
	if val, ok := s.cache.Get(k); ok {
		fmt.Printf("[Cache Hit] Key: %s, Value: %s\n", k, val)
		return val, nil
	}

	// If not in cache, check the database
	var val string
	err := s.db.QueryRow("SELECT val FROM kv WHERE key = ?", k).Scan(&val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("[Cache Miss] Key: %s not found\n", k)
			return "", errors.New("not found")
		}
		return "", err
	}

	// Cache the value for future access
	s.cache.Set(k, val)
	fmt.Printf("[Cache Miss] Key: %s found in DB, caching it\n", k)
	return val, nil
}

func (s *LRUSQLiteBackend) Set(k, v string) error {
	// Write to the database
	_, err := s.db.Exec("INSERT OR REPLACE INTO kv(key, val) VALUES (?, ?)", k, v)
	if err != nil {
		return err
	}
	// Write to the cache
	s.cache.Set(k, v)
	fmt.Printf("[Write] Key: %s, Value: %s\n", k, v)
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Initialize the backend
	backend := NewLRUSQLiteBackend("kv_store.db", 5)

	// Wordlist to populate the database
	words := []string{
		"apple", "banana", "cherry", "date", "elderberry",
		"fig", "grape", "honeydew", "kiwi", "lemon",
		"mango", "nectarine", "orange", "papaya", "quince",
		"raspberry", "strawberry", "tangerine", "ugli", "watermelon",
	}

	// Insert all words into the backend
	for _, word := range words {
		backend.Set(word, fmt.Sprintf("Definition of %s", word))
	}

	fmt.Println("\n--- Random Access Demonstration ---")

	// Randomly access words to trigger cache hits and misses
	for i := 0; i < 10; i++ {
		word := words[rand.Intn(len(words))]
		val, err := backend.Get(word)
		if err != nil {
			fmt.Println("[ERROR]", err)
		} else {
			fmt.Printf("Fetched -> %s: %s\n", word, val)
		}
		time.Sleep(200 * time.Millisecond)
	}
}
