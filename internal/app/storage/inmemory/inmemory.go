// Package inmemory реализует in-memory хранилище для сервиса сокращения URL с персистентностью в файл.
//
// Основные особенности:
// - Хранение данных в памяти с синхронизацией через RWMutex
// - Сохранение данных в файл в формате JSON (append-only лог)
// - Поддержка транзакционности операций
// - Оптимизированное чтение для операций редиректа
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

// InMemoryStorage реализует интерфейс хранилища с in-memory кешем и файловой персистентностью.
//
// Структура использует:
// - userURLIndex: индекс для быстрого поиска по пользователю и оригинальному URL
// - shortCodeMap: основное хранилище сопоставлений
// - countRecords: счетчик записей для генерации UUID
// - file/encoder: для персистентного хранения
// - mu: RWMutex для синхронизации доступа
type InMemoryStorage struct {
	userURLIndex map[string]map[string]string
	shortCodeMap map[string]models.UserURLMapping
	file         *os.File
	encoder      *json.Encoder
	countRecords uint64
	mu           sync.RWMutex
}

// NewInMemoryStorage создает новое in-memory хранилище с файловой персистентностью.
//
// Параметры:
//
//	fileStoragePath - путь к файлу для хранения данных
//
// Возвращает:
//
//	*InMemoryStorage - инициализированное хранилище
//	error - ошибка при создании файла
//
// Пример:
//
//	storage, err := NewInMemoryStorage("data/storage.json")
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

// Load загружает данные из файла в память.
//
// Формат файла: JSON-строки (по одной на запись)
// В случае ошибки в строке она пропускается, но загрузка продолжается
//
// Параметры:
//
//	fileStoragePath - путь к файлу с данными
//
// Возвращает:
//
//	error - ошибка чтения файла
func (s *InMemoryStorage) Load(fileStoragePath string) error {
	file, err := os.OpenFile(fileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var countRecords uint64

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var url models.UserURLMapping
		if err := json.Unmarshal(line, &url); err != nil {
			continue
		}

		if url.UserID == "" || url.OriginalURL == "" || url.ShortURL == "" {
			continue
		}

		if _, ok := s.userURLIndex[url.UserID]; !ok {
			s.userURLIndex[url.UserID] = make(map[string]string)
		}

		s.userURLIndex[url.UserID][url.OriginalURL] = url.ShortURL
		s.shortCodeMap[url.ShortURL] = url
		countRecords++
	}

	s.countRecords = countRecords
	return scanner.Err()
}

// GetShortKey возвращает короткий ключ для оригинального URL пользователя.
//
// Параметры:
//
//	ctx - контекст с userID
//	originalURL - URL для поиска
//
// Возвращает:
//
//	models.URLMapping - найденное соответствие
//	error - ошибка поиска (storage.ErrURLNotFound если не найден)
func (s *InMemoryStorage) GetShortKey(ctx context.Context, originalURL string) (models.URLMapping, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID := ctx.Value(jwtauth.UserIDContextKey)
	if userID == nil {
		return models.URLMapping{}, errors.New("userID is not set")
	}

	shortKey, found := s.userURLIndex[userID.(string)][originalURL]
	if !found {
		return models.URLMapping{}, storage.ErrURLNotFound
	}

	return models.URLMapping{
		ShortURL:    shortKey,
		OriginalURL: originalURL,
	}, nil
}

// GetRedirectURL возвращает оригинальный URL для редиректа.
//
// Параметры:
//
//	ctx - контекст запроса
//	shortKey - короткий идентификатор URL
//
// Возвращает:
//
//	models.URLMapping - найденное соответствие
//	error:
//	  - storage.ErrURLNotFound если URL не существует
//	  - storage.ErrURLDeleted если URL помечен как удаленный
func (s *InMemoryStorage) GetRedirectURL(ctx context.Context, shortKey string) (models.URLMapping, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, found := s.shortCodeMap[shortKey]
	if !found {
		return models.URLMapping{}, storage.ErrURLNotFound
	}

	if url.DeletedFlag {
		return models.URLMapping{}, storage.ErrURLDeleted
	}

	return models.URLMapping{
		ShortURL:    url.ShortURL,
		OriginalURL: url.OriginalURL,
	}, nil
}

// SaveURL сохраняет новое соответствие URL.
//
// Параметры:
//
//	ctx - контекст с userID
//	mapping - сохраняемое соответствие URL
//
// Возвращает:
//
//	error:
//	  - storage.ErrShortURLExists если shortURL уже существует
//	  - storage.ErrURLExists если originalURL уже существует
func (s *InMemoryStorage) SaveURL(ctx context.Context, mapping *models.URLMapping) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.shortCodeMap[mapping.ShortURL]; found {
		return storage.ErrShortURLExists
	}

	userID := ctx.Value(jwtauth.UserIDContextKey)
	if userID == nil {
		return errors.New("userID is not set")
	}

	if _, ok := s.userURLIndex[userID.(string)]; !ok {
		s.userURLIndex[userID.(string)] = make(map[string]string)
	}

	if shortURL, exists := s.userURLIndex[userID.(string)][mapping.OriginalURL]; exists {
		mapping.ShortURL = shortURL
		return storage.ErrURLExists
	}

	s.userURLIndex[userID.(string)][mapping.OriginalURL] = mapping.ShortURL
	s.countRecords++

	userURLMapping := models.UserURLMapping{
		UUID:        s.countRecords,
		ShortURL:    mapping.ShortURL,
		OriginalURL: mapping.OriginalURL,
		UserID:      userID.(string),
		DeletedFlag: false,
	}
	s.shortCodeMap[mapping.ShortURL] = userURLMapping

	return s.encoder.Encode(userURLMapping)
}

// Ping проверяет доступность хранилища (всегда возвращает nil).
func (s *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}

// GetExistingURLs возвращает существующие URL из списка.
//
// Параметры:
//
//	ctx - контекст с userID
//	originalURLs - список URL для проверки
//
// Возвращает:
//
//	map[string]string - карта существующих URL (originalURL -> shortURL)
//	error - ошибка операции
func (s *InMemoryStorage) GetExistingURLs(ctx context.Context, originalURLs []string) (map[string]string, error) {
	existing := make(map[string]string)

	if len(originalURLs) == 0 {
		return existing, nil
	}

	for _, originalURL := range originalURLs {
		mapping, err := s.GetShortKey(ctx, originalURL)
		if err != nil && !errors.Is(err, storage.ErrURLNotFound) {
			return nil, err
		}
		if err == nil {
			existing[mapping.OriginalURL] = mapping.ShortURL
		}
	}

	return existing, nil
}

// SaveNewURLs сохраняет список новых URL.
//
// Параметры:
//
//	ctx - контекст с userID
//	urls - список URL для сохранения
//
// Возвращает:
//
//	error - первая ошибка при сохранении
func (s *InMemoryStorage) SaveNewURLs(ctx context.Context, urls []models.URLMapping) error {
	for _, url := range urls {
		if err := s.SaveURL(ctx, &url); err != nil && !errors.Is(err, storage.ErrURLExists) {
			return err
		}
	}
	return nil
}

// GetUserUrls возвращает все URL пользователя.
//
// Параметры:
//
//	ctx - контекст с userID
//	baseURL - базовый URL для построения полных коротких URL
//
// Возвращает:
//
//	[]models.URLMapping - список URL пользователя
//	error - ошибка операции
func (s *InMemoryStorage) GetUserUrls(ctx context.Context, baseURL string) ([]models.URLMapping, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID := ctx.Value(jwtauth.UserIDContextKey)
	if userID == nil {
		return nil, errors.New("userID is not set")
	}

	userURLs, exists := s.userURLIndex[userID.(string)]
	if !exists {
		return nil, nil
	}

	var result []models.URLMapping
	for originalURL, shortCode := range userURLs {
		result = append(result, models.URLMapping{
			OriginalURL: originalURL,
			ShortURL:    baseURL + "/" + shortCode,
		})
	}
	return result, nil
}

// CountURLs возвращает количество сокращённых URL в сервисе.
//
// Параметры:
//
//	ctx - контекст
//
// Возвращает:
//
//	 int - количество сокращённых URL в сервисе
//		error - ошибка операции
func (s *InMemoryStorage) CountURLs(ctx context.Context) (int, error) {
	count := len(s.shortCodeMap)
	return count, nil
}

// CountUsers возвращает количество пользователей в сервисе.
//
// Параметры:
//
//	ctx - контекст
//
// Возвращает:
//
//	 int - количество пользователей в сервисе
//		error - ошибка операции
func (s *InMemoryStorage) CountUsers(ctx context.Context) (int, error) {
	count := len(s.userURLIndex)
	return count, nil
}

// BatchMarkAsDeleted помечает URL пользователя как удаленные.
//
// Параметры:
//
//	userID - идентификатор пользователя
//	urls - список коротких URL для удаления
//
// Возвращает:
//
//	error - ошибка операции
func (s *InMemoryStorage) BatchMarkAsDeleted(userID string, urls []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, code := range urls {
		if mapping, exists := s.shortCodeMap[code]; exists && mapping.UserID == userID {
			mapping.DeletedFlag = true
			s.shortCodeMap[code] = mapping
			if err := s.encoder.Encode(mapping); err != nil {
				return err
			}
		}
	}
	return nil
}

// FilePath возвращает путь к файлу, используемому хранилищем.
// Если файл не открыт, возвращает пустую строку.
func (s *InMemoryStorage) FilePath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.file == nil {
		return ""
	}
	return s.file.Name()
}

// Close освобождает ресурсы
func (s *InMemoryStorage) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
