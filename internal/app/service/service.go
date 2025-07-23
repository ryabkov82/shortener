// Package service реализует бизнес-логику сервиса сокращения URL.
//
// Основные функции:
// - Генерация коротких ключей
// - Управление хранилищем URL
// - Пакетная обработка запросов
// - Асинхронное удаление URL
package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/models"
	"github.com/ryabkov82/shortener/internal/app/workers/deleteurls"
)

// Repository определяет интерфейс для работы с хранилищем URL.
type Repository interface {
	GetShortKey(context.Context, string) (models.URLMapping, error)
	GetRedirectURL(context.Context, string) (models.URLMapping, error)
	SaveURL(context.Context, *models.URLMapping) error
	Ping(context.Context) error
	SaveNewURLs(context.Context, []models.URLMapping) error
	GetExistingURLs(context.Context, []string) (map[string]string, error)
	GetUserUrls(context.Context, string) ([]models.URLMapping, error)
	BatchMarkAsDeleted(userID string, urls []string) error
	Close() error
	CountURLs(ctx context.Context) (int, error)
	CountUsers(ctx context.Context) (int, error)
}

// Service реализует основной сервис приложения.
type Service struct {
	repo         Repository               // Хранилище данных
	deleteworker *deleteurls.DeleteWorker // Воркер для асинхронного удаления
}

// NewService создает новый экземпляр сервиса.
//
// Параметры:
//
//	storage - реализация интерфейса Repository
//
// Возвращает:
//
//	*Service - инициализированный сервис
func NewService(storage Repository) *Service {
	// Инициализация воркера для удаления:
	// - 1 воркер
	// - Буфер на 10 задач
	// - Задержка 500мс перед обработкой
	delworker := deleteurls.NewDeleteWorker(1, 10, 500*time.Millisecond, storage)
	delworker.Start()

	return &Service{
		repo:         storage,
		deleteworker: delworker,
	}
}

// GetShortKey генерирует и сохраняет короткий ключ для URL.
//
// Параметры:
//
//	ctx - контекст с идентификатором пользователя
//	originalURL - URL для сокращения
//
// Возвращает:
//
//	string - сгенерированный короткий ключ
//	error - ошибка при сохранении:
//	  - storage.ErrURLExists если URL уже существует
func (s *Service) GetShortKey(ctx context.Context, originalURL string) (string, error) {
	shortKey := generateShortKey()
	mapping := models.URLMapping{
		ShortURL:    shortKey,
		OriginalURL: originalURL,
	}

	err := s.repo.SaveURL(ctx, &mapping)
	return mapping.ShortURL, err
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
//	string - оригинальный URL
//	error:
//	  - storage.ErrURLNotFound если URL не существует
//	  - storage.ErrURLDeleted если URL помечен как удаленный
func (s *Service) GetRedirectURL(ctx context.Context, shortKey string) (string, error) {
	mapping, err := s.repo.GetRedirectURL(ctx, shortKey)
	return mapping.OriginalURL, err
}

// Ping проверяет доступность хранилища.
func (s *Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

// Batch обрабатывает пакетный запрос на сокращение URL.
//
// Параметры:
//
//	ctx - контекст с идентификатором пользователя
//	batchRequest - список URL для обработки
//	baseURL - базовый адрес для построения полных коротких URL
//
// Возвращает:
//
//	[]models.BatchResponse - результаты обработки
//	error - ошибка при сохранении
func (s *Service) Batch(ctx context.Context, batchRequest []models.BatchRequest, baseURL string) ([]models.BatchResponse, error) {
	originalURLs := make([]string, len(batchRequest))
	for i, item := range batchRequest {
		originalURLs[i] = item.OriginalURL
	}

	existingURLs, err := s.repo.GetExistingURLs(ctx, originalURLs)
	if err != nil {
		return nil, err
	}

	var newURLs []models.URLMapping
	batchResponse := make([]models.BatchResponse, 0, len(batchRequest))

	for _, item := range batchRequest {
		if shortURL, ok := existingURLs[item.OriginalURL]; ok {
			batchResponse = append(batchResponse, models.BatchResponse{
				CorrelationID: item.CorrelationID,
				ShortURL:      baseURL + "/" + shortURL,
			})
			continue
		}

		shortURL := generateShortKey()
		newURLs = append(newURLs, models.URLMapping{
			OriginalURL: item.OriginalURL,
			ShortURL:    shortURL,
		})

		batchResponse = append(batchResponse, models.BatchResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      baseURL + "/" + shortURL,
		})
	}

	if err := s.repo.SaveNewURLs(ctx, newURLs); err != nil {
		return nil, err
	}

	return batchResponse, nil
}

// GetUserUrls возвращает все сокращенные URL пользователя.
//
// Параметры:
//
//	ctx - контекст с идентификатором пользователя
//	baseURL - базовый адрес для построения полных коротких URL
//
// Возвращает:
//
//	[]models.URLMapping - список URL пользователя
//	error - ошибка при получении
func (s *Service) GetUserUrls(ctx context.Context, baseURL string) ([]models.URLMapping, error) {
	return s.repo.GetUserUrls(ctx, baseURL)
}

// GetStats возвращает статистику сервиса по количеству URL и пользователей.
//
// Параметры:
//   - ctx: контекст выполнения, может использоваться для передачи таймаутов
//
// Возвращает:
//   - models.StatsResponse: структура с полями:
//   - URLs: общее количество сокращенных URL в сервисе
//   - Users: количество уникальных пользователей в сервисе
//   - error: ошибка, если не удалось получить статистику:
//   - Ошибка базы данных при запросе CountURLs
//   - Ошибка базы данных при запросе CountUsers
//
// Логика работы:
//  1. Запрашивает общее количество URL через s.repo.CountURLs
//  2. При ошибке на этом шаге сразу возвращает ошибку
//  3. Запрашивает количество пользователей через s.repo.CountUsers
//  4. При ошибке на этом шаге возвращает ошибку
//  5. Формирует и возвращает структуру StatsResponse с полученными данными
//
// Пример использования:
//
//	stats, err := service.GetStats(context.Background())
//	if err != nil {
//	    // обработка ошибки
//	}
//	fmt.Printf("Stats: %d URLs, %d Users\n", stats.URLs, stats.Users)
//
// Особенности:
//   - Метод атомарен - при ошибке любого из запросов статистика не возвращается
//   - Для работы требует корректной инициализации s.repo
//   - Контекст передается в нижележащие репозитории
//
// Взаимодействие с другими компонентами:
//   - Используется в stats.GetHandler для обработки HTTP-запросов
//   - Получает данные через интерфейс Repository
func (s *Service) GetStats(ctx context.Context) (models.StatsResponse, error) {

	urlCount, err := s.repo.CountURLs(ctx)
	if err != nil {
		return models.StatsResponse{}, err
	}

	userCount, err := s.repo.CountUsers(ctx)
	if err != nil {
		return models.StatsResponse{}, err
	}

	return models.StatsResponse{URLs: urlCount, Users: userCount}, nil

}

// DeleteUserUrls помечает URL пользователя как удаленные (асинхронно).
//
// Параметры:
//
//	ctx - контекст с идентификатором пользователя
//	shortURLs - список коротких URL для удаления
//
// Возвращает:
//
//	error - ошибка при постановке задачи в очередь
func (s *Service) DeleteUserUrls(ctx context.Context, shortURLs []string) error {
	userID := ctx.Value(jwtauth.UserIDContextKey).(string)
	return s.deleteworker.Submit(deleteurls.DeleteTask{
		UserID:    userID,
		ShortURLs: shortURLs,
	})
}

// GracefulStop корректно останавливает сервис.
//
// Параметры:
//
//	timeout - максимальное время ожидания завершения операций
func (s *Service) GracefulStop(timeout time.Duration) {
	s.deleteworker.GracefulStop(timeout)
}

// Close освобождает ресурсы
func (s *Service) Close() error {
	return s.repo.Close()
}

// generateShortKey генерирует случайный короткий ключ.
//
// Возвращает:
//
//	string - 8-символьный ключ из [a-zA-Z0-9]
func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 8

	rand.New(rand.NewSource(time.Now().UnixNano()))
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}
