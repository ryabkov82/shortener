package shortenapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/storage"
)

type URLHandler interface {
	GetShortKey(context.Context, string) (string, error)
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func GetHandler(urlHandler URLHandler, baseURL string, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var request Request

		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			http.Error(res, "Failed to read request body", http.StatusBadRequest)
			log.Error("Failed to read request body", zap.Error(err))
			return
		}

		originalURL := string(request.URL)

		if originalURL == "" {
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			log.Error("URL parameter is missing")
			return
		}

		// Проверяем, что передан URL
		_, err := url.ParseRequestURI(originalURL)

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

		log.Debug("shortKey generate", zap.String("shortKey", shortURL))

		res.Header().Set("content-type", "application/json")
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
		response := Response{Result: baseURL + "/" + shortURL}
		resp, err := json.Marshal(response)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			log.Error("Failed to get short URL", zap.Error(err))
			return
		}
		res.Write(resp)

	}
}
