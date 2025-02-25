package storage

type Storage struct {
	// Переменная для хранения редиректов ShortURL -> OriginalURL
	shortURLs map[string]string
	// Переменная для хранения значений OriginalURL -> ShortURL
	originalURLs map[string]string
}

func New() *Storage {

	var shortURLs = make(map[string]string)
	var originalURLs = make(map[string]string)

	return &Storage{shortURLs, originalURLs}
}

func (s *Storage) GetShortKey(originalURL string) (string, bool) {

	shortKey, found := s.originalURLs[originalURL]
	return shortKey, found

}

func (s *Storage) GetRedirectURL(shortKey string) (string, bool) {

	originalURL, found := s.shortURLs[shortKey]
	return originalURL, found

}

func (s *Storage) SaveURL(originalURL string, shortKey string) {

	s.shortURLs[shortKey] = originalURL
	s.originalURLs[originalURL] = shortKey

}
