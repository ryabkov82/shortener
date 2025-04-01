package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/ryabkov82/shortener/internal/app/models"
	"github.com/ryabkov82/shortener/internal/app/storage"
)

type Repository interface {
	GetShortKey(context.Context, string) (models.URLMapping, error)
	GetRedirectURL(context.Context, string) (models.URLMapping, error)
	SaveURL(context.Context, models.URLMapping) error
	Ping(context.Context) error
}

type Service struct {
	repo Repository
}

func NewService(storage Repository) *Service {
	return &Service{repo: storage}
}

func (s *Service) GetShortKey(ctx context.Context, originalURL string) (string, error) {

	// Возможно, shortURL уже сгенерирован...
	mapping, err := s.repo.GetShortKey(ctx, originalURL)
	if err != nil && err == storage.ErrURLNotFound {
		// Генерируем короткий URL
		/*
			generated := false
			shortKey := ""
			for i := 1; i < 3; i++ {
				shortKey = generateShortKey()
				// Возможно, shortURL был сгененрирован ранее
				_, found := urlHandler.GetRedirectURL(shortKey)
				if !found {
					generated = true
					break
				}
			}
			if generated {
				// Cохраняем переданный URL
				mapping = models.URLMapping{
					ShortURL:    shortKey,
					OriginalURL: originalURL,
				}

				err := urlHandler.SaveURL(mapping)
				if err != nil {
					http.Error(res, "Failed to save URL", http.StatusInternalServerError)
					log.Println("Failed to save URL", err)
					return
				}
			} else {
				// Не удалось сгененрировать новый shortURL
				http.Error(res, "Failed to generate a new shortURL", http.StatusBadRequest)
				log.Println("Failed to generate a new shortURL")
				return
			}
		*/
		shortKey := generateShortKey()
		// Cохраняем переданный URL
		mapping = models.URLMapping{
			ShortURL:    shortKey,
			OriginalURL: originalURL,
		}

		err = s.repo.SaveURL(ctx, mapping)
		if err != nil {
			return "", err
		}

	}
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
