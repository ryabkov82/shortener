// Package shortenapi предоставляет JSON API для создания сокращённых URL.
//
// Пакет реализует:
// - Приём оригинального URL в JSON-формате
// - Валидацию входящего URL
// - Генерацию короткого ключа
// - Возврат результата в стандартизированном JSON-формате
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

// URLHandler определяет контракт для генерации коротких URL.
type URLHandler interface {
	// GetShortKey возвращает короткий ключ для оригинального URL.
	//
	// Параметры:
	//   ctx - контекст выполнения
	//   originalURL - URL для сокращения (должен быть валидным)
	//
	// Возвращает:
	//   string - короткий ключ
	//   error - возможные ошибки:
	//     - storage.ErrURLExists: URL уже существует
	//     - другие внутренние ошибки
	GetShortKey(ctx context.Context, originalURL string) (string, error)
}

// Request представляет структуру входящего JSON-запроса.
type Request struct {
	URL string `json:"url"` // Оригинальный URL для сокращения
}

// Response представляет структуру исходящего JSON-ответа.
type Response struct {
	Result string `json:"result"` // Полный сокращённый URL
}

// GetHandler создаёт HTTP-обработчик для API сокращения URL.
//
// Спецификация API:
//
//	Метод: POST
//	Content-Type: application/json
//	Путь: /api/shorten
//
// Формат запроса:
//
//	{
//	  "url": "https://example.com/very/long/url"
//	}
//
// Формат ответа:
//
//	{
//	  "result": "http://short.ly/abc123"
//	}
//
// Коды ответа:
//   - 201 Created: URL успешно сокращён
//   - 400 Bad Request: невалидный запрос
//   - 409 Conflict: URL уже существует
//   - 500 Internal Server Error: внутренняя ошибка сервера
//
// Параметры:
//
//	urlHandler - сервис для генерации коротких ключей
//	baseURL - базовый адрес для построения полного короткого URL
//	log - логгер для записи событий
//
// Возвращает:
//
//	http.HandlerFunc - HTTP-обработчик
func GetHandler(urlHandler URLHandler, baseURL string, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var request Request

		// Декодируем JSON-тело запроса
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			http.Error(res, "Failed to read request body", http.StatusBadRequest)
			log.Error("Failed to decode request body",
				zap.Error(err),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
			return
		}

		originalURL := request.URL

		// Валидация обязательного поля URL
		if originalURL == "" {
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			log.Error("Empty URL in request",
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
			return
		}

		// Проверка валидности URL
		if _, err := url.ParseRequestURI(originalURL); err != nil {
			http.Error(res, "Invalid URL format", http.StatusBadRequest)
			log.Error("Invalid URL in request",
				zap.String("url", originalURL),
				zap.Error(err),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
			return
		}

		log.Debug("Processing URL shortening",
			zap.String("originalURL", originalURL),
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path))

		// Генерация короткого ключа
		shortKey, err := urlHandler.GetShortKey(req.Context(), originalURL)

		// Обработка ошибок
		if err != nil && !errors.Is(err, storage.ErrURLExists) {
			http.Error(res, "Failed to generate short URL", http.StatusInternalServerError)
			log.Error("Short URL generation failed",
				zap.Error(err),
				zap.String("originalURL", originalURL),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
			return
		}

		// Формирование ответа
		response := Response{
			Result: baseURL + "/" + shortKey,
		}

		res.Header().Set("Content-Type", "application/json")

		// Установка соответствующего HTTP-статуса
		if errors.Is(err, storage.ErrURLExists) {
			res.WriteHeader(http.StatusConflict)
			log.Debug("URL already exists",
				zap.String("shortKey", shortKey),
				zap.String("originalURL", originalURL))
		} else {
			res.WriteHeader(http.StatusCreated)
			log.Debug("URL successfully shortened",
				zap.String("shortKey", shortKey),
				zap.String("originalURL", originalURL))
		}

		// Кодирование и отправка ответа
		if err := json.NewEncoder(res).Encode(response); err != nil {
			log.Error("Failed to encode response",
				zap.Error(err),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
		}
	}
}
