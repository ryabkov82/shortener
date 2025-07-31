package deluserurls

import (
	"context"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/base"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// URLHandler определяет контракт для обработки удаления URL.
type URLHandler interface {
	DeleteUserUrls(ctx context.Context, ids []string) error
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

func (h *Handler) DeleteUserURLs(
	ctx context.Context,
	req *pb.DeleteRequest,
) (*pb.DeleteResponse, error) {
	if len(req.ShortUrls) == 0 {
		h.Logger.Error("No short URLs provided for deletion")
		return nil, status.Error(codes.InvalidArgument, "No short URLs provided")
	}

	h.Logger.Debug("Processing DeleteUserURLs request",
		zap.Int("url_count", len(req.ShortUrls)))

	err := h.service.DeleteUserUrls(ctx, req.ShortUrls)
	if err != nil {
		h.Logger.Error("Failed to delete user URLs", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to delete user URLs")
	}

	// 202 Accepted, но в gRPC нет статус-кодов как в HTTP — просто возвращаем OK
	return &pb.DeleteResponse{}, nil
}
