// Package userurls предоставляет обработчик для получения списка URL пользователя.
//
// Пакет реализует:
// - Получение всех сокращённых URL авторизованного пользователя
// - Возврат данных в JSON-формате
// - Обработку случая отсутствия URL
package userurls

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/models"
)

// URLHandler определяет контракт для получения URL пользователя.
type URLHandler interface {
	// GetUserUrls возвращает все сокращённые URL пользователя.
	//
	// Параметры:
	//   ctx - контекст выполнения (должен содержать идентификатор пользователя)
	//   baseURL - базовый адрес для построения полных коротких URL
	//
	// Возвращает:
	//   []models.URLMapping - список сопоставлений оригинальных и коротких URL
	//   error - ошибка выполнения (например, проблемы с хранилищем)
	GetUserUrls(ctx context.Context, baseURL string) ([]models.URLMapping, error)
}

// GetHandler создаёт HTTP-обработчик для получения URL пользователя.
//
// Спецификация API:
//
//	Метод: GET
//	Путь: /api/user/urls
//	Требуется: JWT-аутентификация
//
// Формат ответа:
//
//	[
//	  {
//	    "short_url": "http://short.ly/abc123",
//	    "original_url": "https://example.com/long/url"
//	  },
//	  ...
//	]
//
// Коды ответа:
//   - 200 OK: успешный запрос (возвращает список URL)
//   - 204 No Content: у пользователя нет сохранённых URL
//   - 400 Bad Request: ошибка аутентификации
//   - 500 Internal Server Error: внутренняя ошибка сервера
//
// Параметры:
//
//	urlHandler - сервис для получения URL
//	baseURL - базовый адрес сервиса
//	log - логгер для записи событий
//
// Возвращает:
//
//	http.HandlerFunc - HTTP-обработчик
func GetHandler(urlHandler URLHandler, baseURL string, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Получение данных из хранилища
		responseData, err := urlHandler.GetUserUrls(req.Context(), baseURL)
		if err != nil {
			http.Error(res, "Failed to get user URLs", http.StatusInternalServerError)
			log.Error("Failed to retrieve user URLs",
				zap.Error(err),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
			return
		}

		// Обработка случая отсутствия URL
		if len(responseData) == 0 {
			res.WriteHeader(http.StatusNoContent)
			log.Debug("No URLs found for user",
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
			return
		}

		// Формирование JSON-ответа
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)

		encoder := json.NewEncoder(res)
		encoder.SetIndent("", "  ") // Форматирование JSON для читаемости
		if err := encoder.Encode(responseData); err != nil {
			log.Error("Failed to encode response",
				zap.Error(err),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
		}

		log.Debug("Successfully returned user URLs",
			zap.Int("count", len(responseData)),
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path))
	}
}
