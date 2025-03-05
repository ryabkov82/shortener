package service

import (
	"math/rand"
	"time"

	"github.com/ryabkov82/shortener/internal/app/models"
)

type Repository interface {
	GetShortKey(string) (models.URLMapping, bool)
	GetRedirectURL(string) (models.URLMapping, bool)
	SaveURL(models.URLMapping) error
}

type Service struct {
	repo Repository
}

func NewService(storage Repository) *Service {
	return &Service{repo: storage}
}

func (s *Service) GetShortKey(originalURL string) (string, error) {

	// Возможно, shortURL уже сгенерирован...
	mapping, found := s.repo.GetShortKey(originalURL)
	if !found {
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

		err := s.repo.SaveURL(mapping)
		if err != nil {
			return "", err
		}

	}
	return mapping.ShortURL, nil
}

func (s *Service) GetRedirectURL(shortKey string) (string, bool) {

	// Получаем адрес перенаправления
	mapping, found := s.repo.GetRedirectURL(shortKey)
	if !found {
		return "", found
	}
	return mapping.OriginalURL, found
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
