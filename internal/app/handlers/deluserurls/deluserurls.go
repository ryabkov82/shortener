// Пакет deluserurls предоставляет обработчик для массового удаления URL пользователя.
//
// Пакет реализует:
// - Приём списка URL для удаления в JSON-формате
// - Асинхронное удаление URL
// - Подтверждение принятия запроса
package deluserurls

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// URLHandler определяет контракт для обработки удаления URL.
type URLHandler interface {
	// DeleteUserUrls удаляет указанные URL, принадлежащие пользователю.
	//
	// Параметры:
	//   ctx - контекст выполнения
	//   urls - список коротких URL для удаления (только идентификаторы)
	//
	// Возвращает:
	//   error - ошибка выполнения (не влияет на HTTP-статус ответа)
	DeleteUserUrls(ctx context.Context, urls []string) error
}

// GetHandler создаёт HTTP-обработчик для массового удаления URL пользователя.
//
// Спецификация API:
//
//	Метод: DELETE
//	Content-Type: application/json
//	Путь: /api/user/urls
//
// Формат запроса:
//
//	["url1", "url2", ...]
//
// Формат ответа:
//
//	Тело ответа пустое
//
// Коды ответа:
//   - 202 Accepted - запрос принят в обработку
//   - 400 Bad Request - невалидный JSON
//   - 401 Unauthorized - пользователь не аутентифицирован
//   - 500 Internal Server Error - внутренняя ошибка сервера
//
// Особенности:
//   - Удаление происходит асинхронно
//   - Ответ 202 не гарантирует успешного удаления
//   - Для аутентификации используется JWT-токен в Cookie
//
// Параметры:
//
//	urlHandler - сервис для обработки URL
//	baseURL - базовый адрес сервиса (не используется в текущей реализации)
//	log - логгер для записи событий
//
// Возвращает:
//
//	http.HandlerFunc - HTTP-обработчик
func GetHandler(urlHandler URLHandler, baseURL string, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var shortURLs []string
		if err := json.NewDecoder(req.Body).Decode(&shortURLs); err != nil {
			http.Error(res, "Invalid request body", http.StatusBadRequest)
			return
		}

		err := urlHandler.DeleteUserUrls(req.Context(), shortURLs)

		if err != nil {
			http.Error(res, "Failed to delete user urls", http.StatusBadRequest)
			log.Error("Failed to delete user urls", zap.Error(err))
			return
		}

		res.WriteHeader(http.StatusAccepted)
	}
}
