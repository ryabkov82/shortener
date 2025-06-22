// Пакет redirect предоставляет обработчик для перенаправления по коротким URL.
//
// Пакет реализует:
// - Поиск оригинального URL по короткому идентификатору
// - Обработку различных статусов URL (активен, удален, не найден)
// - Логирование всех операций перенаправления
package redirect

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"

	"github.com/ryabkov82/shortener/internal/app/storage"
)

// URLHandler определяет контракт для получения оригинального URL.
type URLHandler interface {
	// GetRedirectURL возвращает оригинальный URL для перенаправления.
	//
	// Параметры:
	//   ctx - контекст выполнения
	//   id - короткий идентификатор URL
	//
	// Возвращает:
	//   string - оригинальный URL
	//   error - возможные ошибки:
	//     - storage.ErrURLNotFound: URL не существует
	//     - storage.ErrURLDeleted: URL был удален
	//     - другие внутренние ошибки
	GetRedirectURL(ctx context.Context, id string) (string, error)
}

// GetHandler создаёт HTTP-обработчик для перенаправления по коротким URL.
//
// Спецификация API:
//
//	Метод: GET
//	Путь: /{id}
//
// Параметры пути:
//
//	id - короткий идентификатор URL (a-z, A-Z, 0-9)
//
// Ответы:
//   - 307 Temporary Redirect: успешное перенаправление (с Location header)
//   - 404 Not Found: короткий URL не существует
//   - 410 Gone: URL был удален
//   - 500 Internal Server Error: внутренняя ошибка сервера
//
// Особенности:
//   - Все запросы логируются с указанием shortKey
//   - Для удаленных URL возвращается специальный статус 410
//   - Поддерживается контекст для отмены операций
//
// Параметры:
//
//	urlHandler - сервис для получения URL
//	log - логгер для записи событий
//
// Возвращает:
//
//	http.HandlerFunc - HTTP-обработчик
func GetHandler(urlHandler URLHandler, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := chi.URLParam(req, "id")

		// Получаем адрес перенаправления
		originalURL, err := urlHandler.GetRedirectURL(req.Context(), id)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				http.Error(res, "Shortened key not found", http.StatusNotFound)
				log.Info("Shortened key not found",
					zap.String("shortKey", id),
					zap.String("method", req.Method),
					zap.String("path", req.URL.Path))
				return
			}
			if errors.Is(err, storage.ErrURLDeleted) {
				http.Error(res, "URL has been deleted", http.StatusGone)
				log.Info("URL has been deleted",
					zap.String("shortKey", id),
					zap.String("method", req.Method),
					zap.String("path", req.URL.Path))
				return
			}
			http.Error(res, "failed get redirect URL", http.StatusInternalServerError)
			log.Error("failed get redirect URL",
				zap.Error(err),
				zap.String("shortKey", id),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
			return
		}

		log.Info("Shortened key found",
			zap.String("shortKey", id),
			zap.String("redirect", originalURL),
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path))

		// Устанавливаем заголовок ответа Location
		res.Header().Set("Location", originalURL)
		// устанавливаем код 307
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
