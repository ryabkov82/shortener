package stats

import (
	"context"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/base"
	"github.com/ryabkov82/shortener/internal/app/models"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

type Handler struct {
	*base.BaseHandler // Встраиваем базовый обработчик
	service           URLHandler
}

func New(
	baseHandler *base.BaseHandler,
	service URLHandler,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler, // Инициализация базовых зависимостей
		service:     service,
	}
}

// GetStats реализует gRPC хендлер получения статистики
func (h *Handler) GetStats(
	ctx context.Context,
	_ *pb.StatsRequest,
) (*pb.StatsResponse, error) {

	// Получение статистики из сервиса
	stats, err := h.service.GetStats(ctx)
	if err != nil {
		h.Logger.Error("Failed to get stats", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get stats")
	}

	h.Logger.Debug("Stats received successfully")

	// Формирование и возврат ответа
	return &pb.StatsResponse{
		Urls:  int64(stats.URLs),
		Users: int64(stats.Users),
	}, nil
}
