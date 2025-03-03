package shorturl

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type URLHandler interface {
	GetShortKey(string) (string, bool)
	GetRedirectURL(string) (string, bool)
	SaveURL(string, string) error
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
		shortKey, found := urlHandler.GetShortKey(originalURL)
		if !found {
			// Генерируем короткий URL
			generated := false
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
				err := urlHandler.SaveURL(originalURL, shortKey)
				if err != nil {
					http.Error(res, "Failed to save URL", http.StatusInternalServerError)
					return
				}
			} else {
				// Не удалось сгененрировать новый shortURL
				http.Error(res, "Failed to generate a new shortURL", http.StatusBadRequest)
				log.Println("Failed to generate a new shortURL")
				return
			}
		}

		log.Println("shortKey generate", shortKey)

		res.Header().Set("content-type", "text/plain")
		// устанавливаем код 201
		res.WriteHeader(http.StatusCreated)
		// пишем тело ответа
		baseURLParsed, _ := url.Parse(baseURL)
		var u = url.URL{
			Scheme: baseURLParsed.Scheme,
			Host:   baseURLParsed.Host,
			Path:   shortKey,
		}
		resp := fmt.Sprint(u.String())
		res.Write([]byte(resp))

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
