package redirect

import (
	"context"
	"errors"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/base"
	"github.com/ryabkov82/shortener/internal/app/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// GetOriginalURL обработчик gRPC для получения оригинального URL
func (h *Handler) GetOriginalURL(
	ctx context.Context,
	req *pb.GetRequest,
) (*pb.GetResponse, error) {
	// Валидация ID
	if req.ShortUrl == "" {
		h.Logger.Error("Empty ID in request")
		return nil, status.Error(codes.InvalidArgument, "ID parameter is missing")
	}

	h.Logger.Debug("Processing URL lookup",
		zap.String("shortID", req.ShortUrl))

	// Получаем оригинальный URL
	originalURL, err := h.service.GetRedirectURL(ctx, req.ShortUrl)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrURLNotFound):
			h.Logger.Info("Shortened key not found",
				zap.String("shortKey", req.ShortUrl))
			return nil, status.Error(codes.NotFound, "Shortened key not found")
		case errors.Is(err, storage.ErrURLDeleted):
			h.Logger.Info("URL has been deleted",
				zap.String("shortKey", req.ShortUrl))
			return nil, status.Error(codes.NotFound, "URL has been deleted")
		default:
			h.Logger.Error("Failed to get redirect URL",
				zap.Error(err),
				zap.String("shortKey", req.ShortUrl))
			return nil, status.Error(codes.Internal, "Failed to get redirect URL")
		}
	}

	h.Logger.Info("Shortened key found",
		zap.String("shortKey", req.ShortUrl),
		zap.String("redirect", originalURL))

	// Формируем ответ
	return &pb.GetResponse{
		OriginalUrl: originalURL,
	}, nil
}
