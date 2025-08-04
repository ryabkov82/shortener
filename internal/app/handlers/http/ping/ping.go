// Package ping предоставляет обработчик для проверки доступности базы данных.
//
// Пакет реализует:
// - Проверку соединения с базой данных
// - Возврат статуса доступности сервиса
// - Логирование результатов проверки
package ping

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

// URLHandler определяет контракт для проверки соединения с БД.
type URLHandler interface {
	// Ping проверяет соединение с базой данных.
	//
	// Параметры:
	//   ctx - контекст выполнения с таймаутом
	//
	// Возвращает:
	//   error - ошибка соединения или nil при успехе
	Ping(ctx context.Context) error
}

// GetHandler создаёт HTTP-обработчик для проверки доступности БД.
//
// Спецификация API:
//
//	Метод: GET
//	Путь: /ping
//
// Формат ответа:
//   - При успехе: текст "Connect to database is successful"
//   - При ошибке: текст ошибки
//
// Коды ответа:
//   - 200 OK - соединение с БД установлено
//   - 500 Internal Server Error - ошибка соединения с БД
//   - 503 Service Unavailable - сервис недоступен (может добавляться в будущих версиях)
//
// Параметры:
//
//	urlHandler - сервис для проверки соединения
//	log - логгер для записи событий
//
// Возвращает:
//
//	http.HandlerFunc - HTTP-обработчик
func GetHandler(urlHandler URLHandler, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := urlHandler.Ping(req.Context())
		if err != nil {
			http.Error(res, "Failed to connect to database", http.StatusInternalServerError)
			log.Error("Failed to connect to database",
				zap.Error(err),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
			return
		}

		log.Debug("Database connection check successful",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path))

		res.WriteHeader(http.StatusOK)
		res.Write([]byte("Connect to database is successful"))
	}
}
