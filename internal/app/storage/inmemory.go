package storage

import "sync"

type InMemoryStorage struct {
	// Переменная для хранения редиректов ShortURL -> OriginalURL
	shortURLs map[string]string
	// Переменная для хранения значений OriginalURL -> ShortURL
	originalURLs map[string]string
	mu           sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {

	var shortURLs = make(map[string]string)
	var originalURLs = make(map[string]string)

	return &InMemoryStorage{
		shortURLs:    shortURLs,
		originalURLs: originalURLs,
	}
}

func (s *InMemoryStorage) GetShortKey(originalURL string) (string, bool) {

	s.mu.RLock()
	defer s.mu.RUnlock()
	shortKey, found := s.originalURLs[originalURL]
	return shortKey, found

}

func (s *InMemoryStorage) GetRedirectURL(shortKey string) (string, bool) {

	s.mu.RLock()
	defer s.mu.RUnlock()
	originalURL, found := s.shortURLs[shortKey]
	return originalURL, found

}

func (s *InMemoryStorage) SaveURL(originalURL string, shortKey string) {

	s.mu.Lock()
	defer s.mu.Unlock()
	s.shortURLs[shortKey] = originalURL
	s.originalURLs[originalURL] = shortKey

}
