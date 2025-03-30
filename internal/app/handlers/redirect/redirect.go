package redirect

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/ryabkov82/shortener/internal/app/storage"
)

type URLHandler interface {
	GetRedirectURL(string) (string, error)
}

func GetHandler(urlHandler URLHandler, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		id := chi.URLParam(req, "id")

		// Получаем адрес перенаправления
		originalURL, err := urlHandler.GetRedirectURL(id)
		if err != nil {
			if err == storage.ErrURLNotFound {
				http.Error(res, "Shortened key not found", http.StatusNotFound)
				log.Info("Shortened key not found", zap.String("shortKey", id))
				return
			}
			http.Error(res, "failed get redirect URL", http.StatusInternalServerError)
			log.Error("failed get redirect URL", zap.Error(err))
			return
		}
		log.Info("Shortened key found", zap.String("shortKey", id), zap.String("redirect", originalURL))

		// Устанавливаем заголовок ответа Location
		res.Header().Set("Location", originalURL)
		// устанавливаем код 307
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
