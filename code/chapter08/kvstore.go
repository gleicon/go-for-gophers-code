package main

import (
	"container/list"
	"database/sql"
	"errors"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// KVStore defines a simple key-value interface
type KVStore interface {
	Get(key string) (string, error)
	Set(key, val string) error
	Delete(key string) error
}

// MemStore is an in-memory backend
type MemStore struct {
	data map[string]string
}

func NewMemStore() *MemStore {
	return &MemStore{data: make(map[string]string)}
}

func (m *MemStore) Get(k string) (string, error) {
	v, ok := m.data[k]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}

func (m *MemStore) Set(k, v string) error {
	m.data[k] = v
	return nil
}

func (m *MemStore) Delete(k string) error {
	delete(m.data, k)
	return nil
}

// SQLiteStore uses a local sqlite database
type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) *SQLiteStore {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("failed to open sqlite: %v", err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS kv (key TEXT PRIMARY KEY, val TEXT)")
	if err != nil {
		log.Fatalf("failed to create kv table: %v", err)
	}
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) Get(k string) (string, error) {
	var v string
	err := s.db.QueryRow("SELECT val FROM kv WHERE key = ?", k).Scan(&v)
	if err != nil {
		return "", errors.New("not found")
	}
	return v, nil
}

func (s *SQLiteStore) Set(k, v string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO kv(key, val) VALUES (?, ?)", k, v)
	return err
}

func (s *SQLiteStore) Delete(k string) error {
	_, err := s.db.Exec("DELETE FROM kv WHERE key = ?", k)
	return err
}

// LRUCache is a fixed-size key-value cache
type entry struct{ key, val string }

type LRUCache struct {
	cap  int
	list *list.List
	data map[string]*list.Element
}

func NewLRU(cap int) *LRUCache {
	return &LRUCache{cap, list.New(), make(map[string]*list.Element)}
}

func (c *LRUCache) Get(k string) (string, bool) {
	if e, ok := c.data[k]; ok {
		c.list.MoveToFront(e)
		return e.Value.(entry).val, true
	}
	return "", false
}

func (c *LRUCache) Set(k, v string) {
	if e, ok := c.data[k]; ok {
		c.list.MoveToFront(e)
		e.Value = entry{k, v}
		return
	}
	if c.list.Len() == c.cap {
		old := c.list.Back()
		if old != nil {
			c.list.Remove(old)
			delete(c.data, old.Value.(entry).key)
			log.Printf("[cache] evicted key: %s", old.Value.(entry).key)
		}
	}
	e := c.list.PushFront(entry{k, v})
	c.data[k] = e
}

func main() {
	var store KVStore
	var storeName string

	if os.Getenv("BACKEND") == "sqlite" {
		store = NewSQLiteStore("kv.db")
		storeName = "sqlite"
	} else {
		store = NewMemStore()
		storeName = "memory"
	}

	cache := NewLRU(3)
	keys := []string{"site", "lang", "version", "os", "arch"}

	log.Printf("[setup] Using %s backend\n", storeName)

	// First pass: set values directly
	for _, k := range keys {
		val := strings.ToUpper(k)
		log.Printf("[store] Writing: %s -> %s", k, val)
		if err := store.Set(k, val); err != nil {
			log.Printf("[error] failed to set %s: %v", k, err)
		}
	}

	// Second pass: try reading with cache in front
	log.Println("[read] Reading keys via cache -> fallback -> populate cache")

	for _, k := range keys {
		if val, ok := cache.Get(k); ok {
			log.Printf("[cache hit] %s = %s", k, val)
		} else {
			log.Printf("[cache miss] %s -> checking store...", k)
			val, err := store.Get(k)
			if err != nil {
				log.Printf("[store miss] %s not found", k)
				continue
			}
			cache.Set(k, val)
			log.Printf("[store hit] %s = %s", k, val)
		}
	}

	log.Println("[done] Final cache contents:")
	for k, e := range cache.data {
		log.Printf("%s -> %s", k, e.Value.(entry).val)
	}
}
