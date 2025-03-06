package shorturl

import (
	"io"
	"log"
	"net/http"
	"net/url"
)

type URLHandler interface {
	GetShortKey(string) (string, error)
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
		shortURL, err := urlHandler.GetShortKey(originalURL)

		if err != nil {
			http.Error(res, "Failed to get short URL", http.StatusInternalServerError)
			log.Println("Failed to get short URL", err)
			return
		}

		log.Println("shortKey generate", shortURL)

		res.Header().Set("content-type", "text/plain")
		// устанавливаем код 201
		res.WriteHeader(http.StatusCreated)
		// пишем тело ответа
		res.Write([]byte(baseURL + "/" + shortURL))

	}
}
