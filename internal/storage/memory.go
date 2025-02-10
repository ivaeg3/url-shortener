package storage

import (
	"sync"

	"github.com/ivaeg3/url-shortener/pkg/utils"
)

type MemoryStorage struct {
	mu              sync.RWMutex
	shortToOriginal map[string]string
	originalToShort map[string]string
	counter         uint64
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		shortToOriginal: make(map[string]string),
		originalToShort: make(map[string]string),
	}
}

func (s *MemoryStorage) Save(originalURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if shortURL, exists := s.originalToShort[originalURL]; exists {
		return shortURL, nil
	}

	shortURL := utils.Encode(s.counter)

	s.shortToOriginal[shortURL] = originalURL
	s.originalToShort[originalURL] = shortURL

	s.counter++
	return shortURL, nil
}

func (s *MemoryStorage) Get(shortURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	originalURL, exists := s.shortToOriginal[shortURL]
	if !exists {
		return "", ErrNotFound
	}
	return originalURL, nil
}
