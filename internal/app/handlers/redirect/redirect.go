package redirect

import (
	"log"
	"net/http"

	"github.com/ryabkov82/shortener/internal/app/storage"
)

func GetHandler(storage *storage.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.PathValue("id")

		// Получаем адрес перенаправления
		originalURL, found := storage.GetRedirectURL(id)
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
