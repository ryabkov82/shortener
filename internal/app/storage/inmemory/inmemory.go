package inmemory

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/models"
	"github.com/ryabkov82/shortener/internal/app/storage"
)

type InMemoryStorage struct {
	userURLIndex map[string]map[string]string     // userID -> originalURL -> shortCode
	shortCodeMap map[string]models.UserURLMapping // shortCode -> UserURLMapping
	countRecords uint64
	file         *os.File
	encoder      *json.Encoder
	mu           sync.RWMutex
}

func NewInMemoryStorage(fileStoragePath string) (*InMemoryStorage, error) {

	file, err := os.OpenFile(fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &InMemoryStorage{
		userURLIndex: make(map[string]map[string]string),
		shortCodeMap: make(map[string]models.UserURLMapping),
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
		line := scanner.Bytes()

		// Пропускаем пустые строки
		if len(line) == 0 {
			continue
		}

		var url models.UserURLMapping

		if err := json.Unmarshal(line, &url); err != nil {
			continue // Пропускаем некорректные записи, но продолжаем загрузку
		}

		// Валидация обязательных полей
		if url.UserID == "" || url.OriginalURL == "" || url.ShortURL == "" {
			continue
		}

		// Обновляем userURLIndex
		if _, ok := s.userURLIndex[url.UserID]; !ok {
			s.userURLIndex[url.UserID] = make(map[string]string)
		}

		// Для append-only лога последняя запись перезаписывает предыдущие
		s.userURLIndex[url.UserID][url.OriginalURL] = url.ShortURL
		s.shortCodeMap[url.ShortURL] = url

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

	var err error

	userID := ctx.Value(jwtauth.UserIDContextKey)

	if userID == nil {
		return models.URLMapping{}, errors.New("userID is not set")
	}

	// Проверка существования URL
	shortKey, found := s.userURLIndex[userID.(string)][originalURL]

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

	userID := ctx.Value(jwtauth.UserIDContextKey)
	if userID == nil {
		return models.URLMapping{}, errors.New("userID is not set")
	}

	url, found := s.shortCodeMap[shortKey]

	if !found {
		return models.URLMapping{}, storage.ErrURLNotFound
	}

	if url.UserID != userID {
		return models.URLMapping{}, storage.ErrURLNotFound
	}

	mapping := models.URLMapping{
		ShortURL:    url.ShortURL,
		OriginalURL: url.OriginalURL,
	}

	return mapping, nil

}

func (s *InMemoryStorage) SaveURL(ctx context.Context, mapping *models.URLMapping) error {

	// Устанавливаем блокировку
	s.mu.Lock()
	defer s.mu.Unlock()

	// После установки блокировки проверяем нет ли записи с таким ShortURL
	// Возможно, shortURL был сгененрирован ранее
	_, found := s.shortCodeMap[mapping.ShortURL]

	if found {
		return storage.ErrShortURLExists
	}

	userID := ctx.Value(jwtauth.UserIDContextKey)
	if userID == nil {
		return errors.New("userID is not set")
	}

	if _, ok := s.userURLIndex[userID.(string)]; !ok {
		s.userURLIndex[userID.(string)] = make(map[string]string)
	}

	// Проверка существования URL
	if shortURL, exists := s.userURLIndex[userID.(string)][mapping.OriginalURL]; exists {
		mapping.ShortURL = shortURL
		return storage.ErrURLExists
	}

	// Добавляем записи
	s.userURLIndex[userID.(string)][mapping.OriginalURL] = mapping.ShortURL

	s.countRecords++

	userURLMapping := models.UserURLMapping{
		UUID:        s.countRecords,
		ShortURL:    mapping.ShortURL,
		OriginalURL: mapping.OriginalURL,
		UserID:      userID.(string),
	}
	s.shortCodeMap[mapping.ShortURL] = userURLMapping

	err := s.encoder.Encode(userURLMapping)

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

func (s *InMemoryStorage) GetUserUrls(ctx context.Context, baseURL string) ([]models.URLMapping, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	userID := ctx.Value(jwtauth.UserIDContextKey)
	if userID == nil {
		return nil, errors.New("userID is not set")
	}

	// Проверяем существование пользователя в индексе
	userURLs, exists := s.userURLIndex[userID.(string)]
	if !exists {
		return nil, nil // Возвращаем nil вместо ошибки если пользователь не найден
	}

	var result []models.URLMapping
	// Итерируемся по всем URL пользователя
	for originalURL, shortCode := range userURLs {
		result = append(result, models.URLMapping{
			OriginalURL: originalURL,
			ShortURL:    baseURL + "/" + shortCode,
		})

	}
	return result, nil

}
