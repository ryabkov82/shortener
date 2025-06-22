// Пакет logger предоставляет middleware для логирования HTTP-запросов.
package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// RequestLogging создает middleware для логирования HTTP-запросов.
//
// Middleware логирует:
// - HTTP-метод
// - Путь запроса
// - Статус ответа
// - Размер ответа в байтах
// - Время обработки запроса
//
// Параметры:
//
//	log - логгер zap для записи логов
//
// Возвращает:
//
//	func(next http.Handler) http.Handler - middleware функцию
func RequestLogging(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Создаем обертку для ResponseWriter для получения метрик ответа
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Фиксируем время начала обработки запроса
			t1 := time.Now()

			// Отложенное выполнение логирования после обработки запроса
			defer func() {
				log.Info("request completed",
					zap.String("method", r.Method),                  // HTTP-метод (GET, POST и т.д.)
					zap.String("path", r.URL.Path),                  // Путь запроса
					zap.Int("status", ww.Status()),                  // HTTP-статус ответа
					zap.Int("bytes", ww.BytesWritten()),             // Размер ответа в байтах
					zap.String("duration", time.Since(t1).String()), // Время обработки
				)
			}()

			// Передаем управление следующему обработчику
			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
