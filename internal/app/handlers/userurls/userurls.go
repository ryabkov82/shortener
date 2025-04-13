package userurls

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/models"
)

type URLHandler interface {
	GetUserUrls(context.Context, string) ([]models.URLMapping, error)
}

func GetHandler(urlHandler URLHandler, baseURL string, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		responseData, err := urlHandler.GetUserUrls(req.Context(), baseURL)

		if err != nil {
			http.Error(res, "Failed to get user urls", http.StatusBadRequest)
			log.Error("Failed to get user urls", zap.Error(err))
			return
		}

		if len(responseData) == 0 {
			// устанавливаем код 204
			res.WriteHeader(http.StatusNoContent)
		} else {
			res.Header().Set("content-type", "application/json")
			// устанавливаем код 200
			res.WriteHeader(http.StatusOK)
			// пишем тело ответа
			resp, err := json.Marshal(responseData)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				log.Error("Failed to encode response data", zap.Error(err))
				return
			}
			res.Write(resp)

		}

	}
}
