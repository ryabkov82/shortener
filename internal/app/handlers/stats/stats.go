package stats

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ryabkov82/shortener/internal/app/models"
	"go.uber.org/zap"
)

type URLHandler interface {
	GetStats(ctx context.Context) (models.StatsResponse, error)
}

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
