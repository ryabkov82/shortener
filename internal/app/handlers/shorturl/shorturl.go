package shorturl

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/ryabkov82/shortener/internal/app/storage"
	"go.uber.org/zap"
)

type URLHandler interface {
	GetShortKey(context.Context, string) (string, error)
}

func GetHandler(urlHandler URLHandler, baseURL string, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Failed to read request body", http.StatusBadRequest)
			log.Error("Failed to read request body", zap.Error(err))
			return
		}
		originalURL := string(body)

		if originalURL == "" {
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			log.Error("URL parameter is missing")
			return
		}

		// Проверяем, что передан URL
		_, err = url.ParseRequestURI(originalURL)

		if err != nil {
			http.Error(res, "invalid request", http.StatusBadRequest)
			log.Error("invalid request", zap.Error(err))
			return
		}

		log.Debug("get URL", zap.String("originalURL", originalURL))

		// Возможно, shortURL уже сгенерирован...
		shortURL, err := urlHandler.GetShortKey(req.Context(), originalURL)

		if err != nil {
			if !errors.Is(err, storage.ErrURLExists) {
				http.Error(res, "Failed to get short URL", http.StatusInternalServerError)
				log.Error("Failed to get short URL", zap.Error(err))
				return
			}
		}

		res.Header().Set("content-type", "text/plain")
		if err == nil {
			log.Debug("shortKey generate", zap.String("shortKey", shortURL))
			// устанавливаем код 201
			res.WriteHeader(http.StatusCreated)
		} else {
			log.Debug("url exists, shortKey", zap.String("shortKey", shortURL))
			// устанавливаем код 409 Conflict
			res.WriteHeader(http.StatusConflict)
		}
		// пишем тело ответа
		res.Write([]byte(baseURL + "/" + shortURL))

	}
}
