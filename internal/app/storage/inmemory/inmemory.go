package inmemory

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/ryabkov82/shortener/internal/app/models"
	"github.com/ryabkov82/shortener/internal/app/storage"
)

type InMemoryStorage struct {
	// Переменная для хранения редиректов ShortURL -> OriginalURL
	shortURLs map[string]string
	// Переменная для хранения значений OriginalURL -> ShortURL
	originalURLs map[string]string
	countRecords uint64
	file         *os.File
	encoder      *json.Encoder
	mu           sync.RWMutex
}

// структура хранения записей в файле
type record struct {
	UUID        uint64 `json:"uuid"`
	ShortURL    string `json:"short_url"`    // Короткий URL
	OriginalURL string `json:"original_url"` // Оригинальный URL
}

func NewInMemoryStorage(fileStoragePath string) (*InMemoryStorage, error) {

	file, err := os.OpenFile(fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &InMemoryStorage{
		shortURLs:    make(map[string]string),
		originalURLs: make(map[string]string),
		countRecords: 0,
		file:         file,
		encoder:      json.NewEncoder(file),
	}, nil
}

func (s *InMemoryStorage) Load(fileStoragePath string) error {

	file, err := os.OpenFile(fileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	// Создаем сканер для чтения файла построчно
	scanner := bufio.NewScanner(file)

	var countRecords uint64

	// Читаем файл построчно
	for scanner.Scan() {
		line := scanner.Text()

		// Пропускаем пустые строки
		if len(line) == 0 {
			continue
		}

		// Декодируем JSON-строку в структуру
		var record record
		err := json.Unmarshal([]byte(line), &record)
		if err != nil {
			continue // Пропускаем некорректные строки и продолжаем чтение
		}

		s.shortURLs[record.ShortURL] = record.OriginalURL
		s.originalURLs[record.OriginalURL] = record.ShortURL

		countRecords++

	}

	s.countRecords = countRecords

	// Проверяем, не возникла ли ошибка при сканировании
	err = scanner.Err()

	return err
}

func (s *InMemoryStorage) GetShortKey(ctx context.Context, originalURL string) (models.URLMapping, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()
	shortKey, found := s.originalURLs[originalURL]

	var err error
	if !found {
		shortKey = ""
		err = storage.ErrURLNotFound
	}

	mapping := models.URLMapping{
		ShortURL:    shortKey,
		OriginalURL: originalURL,
	}

	return mapping, err
}

func (s *InMemoryStorage) GetRedirectURL(ctx context.Context, shortKey string) (models.URLMapping, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()
	originalURL, found := s.shortURLs[shortKey]

	var err error
	if !found {
		originalURL = ""
		err = storage.ErrURLNotFound
	}

	mapping := models.URLMapping{
		ShortURL:    shortKey,
		OriginalURL: originalURL,
	}

	return mapping, err

}

func (s *InMemoryStorage) SaveURL(ctx context.Context, mapping *models.URLMapping) error {

	// Устанавливаем блокировку
	s.mu.Lock()
	defer s.mu.Unlock()

	// После установки блокировки проверяем нет ли записи с таким ShortURL
	// Возможно, shortURL был сгененрирован ранее
	_, found := s.shortURLs[mapping.ShortURL]

	if found {
		return storage.ErrShortURLExists
	}

	// Проверяем существует ли уже короткий url для данного OriginalURL
	shortURL, found := s.originalURLs[mapping.OriginalURL]

	if found {
		mapping.ShortURL = shortURL
		return storage.ErrURLExists
	}

	s.shortURLs[mapping.ShortURL] = mapping.OriginalURL
	s.originalURLs[mapping.OriginalURL] = mapping.ShortURL

	s.countRecords++

	// сохраняем данные в файл
	record := record{UUID: s.countRecords, ShortURL: mapping.ShortURL, OriginalURL: mapping.OriginalURL}
	err := s.encoder.Encode(record)

	return err
}

func (s *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *InMemoryStorage) GetExistingURLs(ctx context.Context, originalURLs []string) (map[string]string, error) {

	existing := make(map[string]string)

	if len(originalURLs) == 0 {
		return existing, nil
	}

	for _, originalURL := range originalURLs {
		mapping, err := s.GetShortKey(ctx, originalURL)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				continue
			} else {
				return nil, err
			}
		}
		existing[mapping.OriginalURL] = mapping.ShortURL
	}

	return existing, nil

}

func (s *InMemoryStorage) SaveNewURLs(ctx context.Context, urls []models.URLMapping) error {
	if len(urls) == 0 {
		return nil
	}

	for _, url := range urls {
		err := s.SaveURL(ctx, &url)
		if err != nil {
			return err
		}
	}

	return nil
}
