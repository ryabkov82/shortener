package batch

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/models"
)

type URLHandler interface {
	Batch(context.Context, []models.BatchRequest, string) ([]models.BatchResponse, error)
}

func GetHandler(urlHandler URLHandler, baseURL string, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		// Декодируем тело запроса
		var requestData []models.BatchRequest
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&requestData)
		if err != nil {
			http.Error(res, "Failed to read request body", http.StatusBadRequest)
			log.Error("Failed to read request body", zap.Error(err))
			return
		}

		responseData, err := urlHandler.Batch(req.Context(), requestData, baseURL)

		if err != nil {
			http.Error(res, "Failed to proccessing request data", http.StatusBadRequest)
			log.Error("Failed to proccessing request data", zap.Error(err))
			return
		}

		res.Header().Set("content-type", "application/json")
		// устанавливаем код 201
		res.WriteHeader(http.StatusCreated)
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
