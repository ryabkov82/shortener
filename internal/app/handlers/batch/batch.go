// Package batch предоставляет обработчик для пакетного создания сокращённых URL.
//
// Пакет реализует:
// - Приём массива URL в JSON-формате
// - Параллельную обработку запросов
// - Возврат результатов в коррелируемом формате
package batch

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/models"
)

// URLHandler определяет контракт для обработки пакетных запросов.
type URLHandler interface {
	// Batch обрабатывает массив запросов на сокращение URL.
	//
	// Параметры:
	//   ctx - контекст выполнения
	//   requests - массив запросов
	//   baseURL - базовый URL для генерации коротких ссылок
	//
	// Возвращает:
	//   []models.BatchResponse - массив результатов
	//   error - ошибка выполнения
	Batch(ctx context.Context, requests []models.BatchRequest, baseURL string) ([]models.BatchResponse, error)
}

// GetHandler создаёт HTTP-обработчик для пакетного создания URL.
//
// Спецификация API:
//
//	Метод: POST
//	Content-Type: application/json
//	Путь: /api/shorten/batch
//
// Формат запроса:
//
//	[
//	  {
//	    "correlation_id": "уникальный_идентификатор",
//	    "original_url": "https://example.com"
//	  },
//	  ...
//	]
//
// Формат ответа:
//
//	[
//	  {
//	    "correlation_id": "уникальный_идентификатор",
//	    "short_url": "http://short.ly/abc"
//	  },
//	  ...
//	]
//
// Коды ответа:
//   - 201 Created - успешная обработка
//   - 400 Bad Request - невалидный JSON
//   - 500 Internal Server Error - внутренняя ошибка сервера
//
// Параметры:
//
//	urlHandler - сервис для обработки URL
//	baseURL - базовый адрес для коротких ссылок
//	log - логгер для записи событий
//
// Возвращает:
//
//	http.HandlerFunc - HTTP-обработчик
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
