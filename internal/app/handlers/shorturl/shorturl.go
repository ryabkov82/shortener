package shorturl

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/ryabkov82/shortener/internal/app/models"
)

type URLHandler interface {
	GetShortKey(string) (models.URLMapping, bool)
	GetRedirectURL(string) (models.URLMapping, bool)
	SaveURL(models.URLMapping) error
}

func GetHandler(urlHandler URLHandler, baseURL string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Failed to read request body", http.StatusBadRequest)
			log.Println(err)
			return
		}
		originalURL := string(body)

		if originalURL == "" {
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			log.Println("URL parameter is missing")
			return
		}

		// Проверяем, что передан URL
		parsedURL, err := url.ParseRequestURI(originalURL)

		if err != nil {
			http.Error(res, "invalid request", http.StatusBadRequest)
			log.Println(err)
			return
		}

		log.Println("get URL", parsedURL)

		// Возможно, shortURL уже сгенерирован...
		mapping, found := urlHandler.GetShortKey(originalURL)
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

			err := urlHandler.SaveURL(mapping)
			if err != nil {
				http.Error(res, "Failed to save URL", http.StatusInternalServerError)
				log.Println("Failed to save URL", err)
				return
			}

		}

		log.Println("shortKey generate", mapping.ShortURL)

		res.Header().Set("content-type", "text/plain")
		// устанавливаем код 201
		res.WriteHeader(http.StatusCreated)
		// пишем тело ответа
		res.Write([]byte(baseURL + "/" + mapping.ShortURL))

	}
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
