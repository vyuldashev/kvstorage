package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

type store struct {
	mutex       sync.RWMutex
	filePath    *string
	items       map[string]*item
	sortedItems []*item
}

type item struct {
	Key       string    `json:"value"`
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewStore(filePath *string) *store {
	s := &store{
		filePath: filePath,
		items:    make(map[string]*item),
	}

	if filePath != nil {
		s.createFile()
		s.restoreFromDisk()
	}

	go s.expirationProcessor()

	return s
}

func (s *store) Put(key string, value string, ttl int) {
	item := &item{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}

	s.mutex.Lock()
	s.items[key] = item

	index := sort.Search(len(s.sortedItems), func(i int) bool { return s.sortedItems[i].ExpiresAt.UnixNano() > item.ExpiresAt.UnixNano() })
	s.sortedItems = append(s.sortedItems, item)
	copy(s.sortedItems[index+1:], s.sortedItems[index:])
	s.sortedItems[index] = item

	s.mutex.Unlock()
}

func (s *store) Get(key string) (*string, error) {
	s.mutex.RLock()
	item, ok := s.items[key]
	s.mutex.RUnlock()

	// if key not found or item is expired but not removed error should be returned
	if !ok || item.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("key not found")
	}

	return &item.Value, nil
}

func (s *store) Forget(key string) {
	s.mutex.Lock()

	if _, ok := s.items[key]; ok {
		delete(s.items, key)
	}

	s.mutex.Unlock()
}

func (s *store) Close() {
	s.persistToDisk()
}

func (s *store) expirationProcessor() {
	for now := range time.Tick(time.Second) {
		s.mutex.Lock()

		// здесь нужно чтобы ключи были отсортированы по TTL
		for k, item := range s.sortedItems {
			if item.ExpiresAt.After(now) {
				break
			}

			delete(s.items, item.Key)
			s.sortedItems = append(s.sortedItems[:k], s.sortedItems[k+1:]...)
		}

		s.mutex.Unlock()
	}
}

func (s *store) createFile() {
	if s.filePath == nil {
		return
	}

	_, err := os.Stat(*s.filePath)

	if os.IsNotExist(err) {
		f, err := os.Create(*s.filePath)

		defer f.Close()

		if err != nil {
			log.Fatalf("failed to open cache file (%v)", err)
		}

		_, err = f.WriteString("{}")

		if err != nil {
			log.Fatalf("failed to open cache file (%v)", err)
		}
	}
}

func (s *store) persistToDisk() {
	if s.filePath == nil {
		return
	}

	s.createFile()

	f, err := os.OpenFile(*s.filePath, os.O_WRONLY, 0644)

	defer f.Close()

	if err != nil {
		log.Fatal("failed to open file", err)

		return
	}

	os.Truncate(*s.filePath, 0)

	w := bufio.NewWriter(f)
	s.mutex.Lock()
	err = json.NewEncoder(w).Encode(s.items)

	if err != nil {
		log.Fatalf("failed to save cache to disk! (%v)", err)

		return
	}

	w.Flush()
	s.mutex.Unlock()
}

func (s *store) restoreFromDisk() {
	if s.filePath == nil {
		return
	}

	s.createFile()

	f, err := os.Open(*s.filePath)

	defer f.Close()

	if err != nil {
		log.Fatal("failed to restore cache from disk (failed to open file)", err)

		return
	}

	r := bufio.NewReader(f)

	items := make(map[string]*item)
	err = json.NewDecoder(r).Decode(&items)

	if err != nil {
		log.Fatalf("failed to restore cache from disk (%v)", err)

		return
	}

	s.mutex.Lock()
	s.items = items
	s.mutex.Unlock()
}
