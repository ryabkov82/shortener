package redirect

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type URLHandler interface {
	GetRedirectURL(string) (string, bool)
}

func GetHandler(urlHandler URLHandler) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		id := chi.URLParam(req, "id")

		// Получаем адрес перенаправления
		originalURL, found := urlHandler.GetRedirectURL(id)
		if !found {
			http.Error(res, "Shortened key not found", http.StatusNotFound)
			log.Println("Shortened key not found", id)
			return
		}
		log.Println("Shortened key found", id, "redirect", originalURL)
		// Устанавливаем заголовок ответа Location
		res.Header().Set("Location", originalURL)
		// устанавливаем код 307
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
