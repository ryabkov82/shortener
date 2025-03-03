package storage

import (
	"sync"

	"github.com/ryabkov82/shortener/internal/app/models"
)

type InMemoryStorage struct {
	// Переменная для хранения редиректов ShortURL -> OriginalURL
	shortURLs map[string]string
	// Переменная для хранения значений OriginalURL -> ShortURL
	originalURLs map[string]string
	mu           sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {

	return &InMemoryStorage{
		shortURLs:    make(map[string]string),
		originalURLs: make(map[string]string),
	}
}

func (s *InMemoryStorage) GetShortKey(originalURL string) (models.URLMapping, bool) {

	s.mu.RLock()
	defer s.mu.RUnlock()
	shortKey, found := s.originalURLs[originalURL]

	if !found {
		shortKey = ""
	}

	mapping := models.URLMapping{
		ShortURL:    shortKey,
		OriginalURL: originalURL,
	}

	return mapping, found
}

func (s *InMemoryStorage) GetRedirectURL(shortKey string) (models.URLMapping, bool) {

	s.mu.RLock()
	defer s.mu.RUnlock()
	originalURL, found := s.shortURLs[shortKey]

	if !found {
		originalURL = ""
	}

	mapping := models.URLMapping{
		ShortURL:    shortKey,
		OriginalURL: originalURL,
	}

	return mapping, found

}

func (s *InMemoryStorage) SaveURL(mapping models.URLMapping) error {

	s.mu.Lock()
	defer s.mu.Unlock()
	s.shortURLs[mapping.ShortURL] = mapping.OriginalURL
	s.originalURLs[mapping.OriginalURL] = mapping.ShortURL

	return nil
}
