// Package shorturl предоставляет обработчик для создания сокращённых URL через текстовый интерфейс.
//
// Пакет реализует:
// - Приём оригинального URL в текстовом формате
// - Валидацию входящего URL
// - Генерацию короткого ключа
// - Возврат результата в текстовом формате
package shorturl

import (
	"context"
	"errors"
	"io"
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
	//   ctx - контекст выполнения (должен включать таймаут)
	//   originalURL - валидный URL для сокращения
	//
	// Возвращает:
	//   string - короткий ключ
	//   error - возможные ошибки:
	//     - storage.ErrURLExists: URL уже существует
	//     - другие внутренние ошибки
	GetShortKey(ctx context.Context, originalURL string) (string, error)
}

// GetHandler создаёт HTTP-обработчик для текстового интерфейса сокращения URL.
//
// Спецификация API:
//
//	Метод: POST
//	Content-Type: text/plain
//	Путь: /
//
// Формат запроса:
//
//	Текстовое тело с оригинальным URL (например: "https://example.com/long/url")
//
// Формат ответа:
//
//	Текстовое тело с сокращённым URL (например: "http://short.ly/abc123")
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
		// Чтение и валидация тела запроса
		// Использование io.LimitReader, минимизация аллокаций
		body, err := io.ReadAll(io.LimitReader(req.Body, 1<<20)) // Ограничение 1MB
		if err != nil {
			http.Error(res, "Failed to read request body", http.StatusBadRequest)
			log.Error("Failed to read request body",
				zap.Error(err),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
			return
		}
		defer req.Body.Close()

		originalURL := string(body)
		if originalURL == "" {
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			log.Error("Empty URL in request",
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
			return
		}

		// Валидация URL
		if _, err = url.ParseRequestURI(originalURL); err != nil {
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
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		shortURL := baseURL + "/" + shortKey

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

		if _, err := res.Write([]byte(shortURL)); err != nil {
			log.Error("Failed to write response",
				zap.Error(err),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
		}
	}
}
