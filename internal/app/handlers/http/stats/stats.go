// Package stats предоставляет обработчики HTTP для получения статистики сервиса.
package stats

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ryabkov82/shortener/internal/app/models"
	"go.uber.org/zap"
)

// URLHandler определяет интерфейс для получения статистики сервиса.
// Реализации этого интерфейса должны предоставлять данные о количестве URL и пользователей.
type URLHandler interface {
	// GetStats возвращает статистику сервиса.
	// Возвращает:
	//   - models.StatsResponse с количеством URL и пользователей
	//   - error в случае ошибки при получении данных
	GetStats(ctx context.Context) (models.StatsResponse, error)
}

// GetHandler создает HTTP-обработчик для получения статистики сервиса.
//
// Параметры:
//   - urlHandler: реализация интерфейса URLHandler для получения данных статистики
//   - log: логгер для записи событий и ошибок
//
// Возвращает:
//   - http.HandlerFunc, который обрабатывает GET-запросы по пути /api/internal/stats
//
// Поведение обработчика:
//   - Проверяет доступ по trusted_subnet (должен быть установлен через middleware)
//   - Возвращает:
//   - 200 OK и JSON с статистикой при успешном выполнении
//   - 403 Forbidden если IP не в доверенной подсети
//   - 500 Internal Server Error при ошибках получения данных
//   - Логирует все ошибки и успешные выполнения
//
// Пример ответа при успехе:
//
//	{
//	  "urls": 100,
//	  "users": 50
//	}
//
// Middleware:
//   - Должен быть установлен trustednet.CheckTrustedSubnet перед этим обработчиком
//   - Рекомендуется добавить mwlogger.RequestLogging для логирования запросов
func GetHandler(urlHandler URLHandler, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		ctx := req.Context()

		stats, err := urlHandler.GetStats(ctx)

		if err != nil {
			http.Error(res, "Failed to get stats", http.StatusInternalServerError)
			log.Error("Failed to get stats",
				zap.Error(err),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
			return
		}

		log.Debug("Stats received successfully",
			zap.String("method", req.Method),
			zap.String("path", req.URL.Path))

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		// Кодирование и отправка ответа
		if err := json.NewEncoder(res).Encode(stats); err != nil {
			log.Error("Failed to encode response",
				zap.Error(err),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path))
		}
	}
}
