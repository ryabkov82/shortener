package shorturl

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
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
					log.Println("Failed to save URL")
					return
				}
			} else {
				// Не удалось сгененрировать новый shortURL
				http.Error(res, "Failed to generate a new shortURL", http.StatusBadRequest)
				log.Println("Failed to generate a new shortURL")
				return
			}
		}

		log.Println("shortKey generate", mapping.ShortURL)

		res.Header().Set("content-type", "text/plain")
		// устанавливаем код 201
		res.WriteHeader(http.StatusCreated)
		// пишем тело ответа
		fullURL := JoinURL(baseURL, mapping.ShortURL)

		res.Write([]byte(fullURL))

	}
}

// JoinURL корректно соединяет базовый URL и короткий URL.
func JoinURL(baseURL, shortURL string) string {
	// Убедимся, что базовый URL не заканчивается на "/"
	baseURL = strings.TrimSuffix(baseURL, "/")
	// Убедимся, что короткий URL не начинается с "/"
	shortURL = strings.TrimPrefix(shortURL, "/")
	return baseURL + "/" + shortURL
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
