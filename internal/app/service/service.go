package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/models"
	"github.com/ryabkov82/shortener/internal/app/workers/deleteurls"
)

type Repository interface {
	GetShortKey(context.Context, string) (models.URLMapping, error)
	GetRedirectURL(context.Context, string) (models.URLMapping, error)
	SaveURL(context.Context, *models.URLMapping) error
	Ping(context.Context) error
	SaveNewURLs(context.Context, []models.URLMapping) error
	GetExistingURLs(context.Context, []string) (map[string]string, error)
	GetUserUrls(context.Context, string) ([]models.URLMapping, error)
	BatchMarkAsDeleted(userID string, urls []string) error
}

type Service struct {
	repo         Repository
	deleteworker *deleteurls.DeleteWorker
}

func NewService(storage Repository) *Service {

	delworker := deleteurls.NewDeleteWorker(1, 10, 500*time.Millisecond, storage)
	delworker.Start()

	return &Service{
		repo:         storage,
		deleteworker: delworker,
	}
}

func (s *Service) GetShortKey(ctx context.Context, originalURL string) (string, error) {

	shortKey := generateShortKey()
	// Cохраняем переданный URL
	mapping := models.URLMapping{
		ShortURL:    shortKey,
		OriginalURL: originalURL,
	}

	err := s.repo.SaveURL(ctx, &mapping)

	return mapping.ShortURL, err
}

func (s *Service) GetRedirectURL(ctx context.Context, shortKey string) (string, error) {

	// Получаем адрес перенаправления
	mapping, err := s.repo.GetRedirectURL(ctx, shortKey)
	return mapping.OriginalURL, err
}

func (s *Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *Service) Batch(ctx context.Context, batchRequest []models.BatchRequest, baseURL string) ([]models.BatchResponse, error) {

	// Собираем все оригинальные URL для проверки
	originalURLs := make([]string, len(batchRequest))
	for i, item := range batchRequest {
		originalURLs[i] = item.OriginalURL
	}

	// Получаем существующие URL одним запросом
	existingURLs, err := s.repo.GetExistingURLs(ctx, originalURLs)
	if err != nil {
		return nil, err
	}

	var newURLs []models.URLMapping
	batchResponse := make([]models.BatchResponse, 0, len(batchRequest))

	for _, item := range batchRequest {
		// Проверяем, есть ли URL уже в базе
		if shortURL, ok := existingURLs[item.OriginalURL]; ok {
			batchResponse = append(batchResponse, models.BatchResponse{
				CorrelationID: item.CorrelationID,
				ShortURL:      baseURL + "/" + shortURL,
			})
			continue
		}

		// Генерируем новый короткий URL
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

	// Сохраняем новые URL пачкой
	if err := s.repo.SaveNewURLs(ctx, newURLs); err != nil {
		return nil, err
	}

	return batchResponse, nil

}

func (s *Service) GetUserUrls(ctx context.Context, baseURL string) ([]models.URLMapping, error) {

	// Получаем существующие URL одним запросом
	URLs, err := s.repo.GetUserUrls(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	return URLs, nil

}

func (s *Service) DeleteUserUrls(ctx context.Context, shortURLs []string) error {

	userID := ctx.Value(jwtauth.UserIDContextKey)

	deltask := deleteurls.DeleteTask{
		UserID:    userID.(string),
		ShortURLs: shortURLs,
	}

	err := s.deleteworker.Submit(deltask)

	return err

}

func (s *Service) GracefulStop(timeout time.Duration) {
	s.deleteworker.GracefulStop(timeout)
}

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
