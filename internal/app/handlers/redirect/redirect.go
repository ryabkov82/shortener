package redirect

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

type URLHandler interface {
	GetRedirectURL(string) (string, bool)
}

func GetHandler(urlHandler URLHandler, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		id := chi.URLParam(req, "id")

		// Получаем адрес перенаправления
		originalURL, found := urlHandler.GetRedirectURL(id)
		if !found {
			http.Error(res, "Shortened key not found", http.StatusNotFound)
			log.Info("Shortened key not found", zap.String("shortKey", id))
			return
		}
		log.Info("Shortened key found", zap.String("shortKey", id), zap.String("redirect", originalURL))

		// Устанавливаем заголовок ответа Location
		res.Header().Set("Location", originalURL)
		// устанавливаем код 307
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
