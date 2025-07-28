package ping

import (
	"context"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/base"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (h *Handler) Ping(
	ctx context.Context,
	req *pb.PingRequest,
) (*pb.PingResponse, error) {
	err := h.service.Ping(ctx)
	if err != nil {
		h.Logger.Error("Failed to connect to database",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, "Failed to connect to database")
	}

	h.Logger.Debug("Database connection check successful")

	return &pb.PingResponse{
		Ok: true,
	}, nil
}
