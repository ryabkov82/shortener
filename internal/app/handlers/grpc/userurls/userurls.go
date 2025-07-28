package userurls

import (
	"context"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/base"
	"github.com/ryabkov82/shortener/internal/app/models"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// URLHandler определяет контракт для получения URL пользователя.
type URLHandler interface {
	GetUserUrls(ctx context.Context, baseURL string) ([]models.URLMapping, error)
}

type Handler struct {
	*base.BaseHandler // Встраиваем базовый обработчик
	service           URLHandler
	baseURL           string
}

func New(
	baseHandler *base.BaseHandler,
	service URLHandler,
	baseURL string,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler, // Инициализация базовых зависимостей
		service:     service,
		baseURL:     baseURL,
	}
}

func (h *Handler) GetUserURLs(
	ctx context.Context,
	_ *pb.UserURLsRequest,
) (*pb.UserURLsResponse, error) {

	h.Logger.Debug("Processing GetUserURLs request")

	// Получение URL пользователя
	urls, err := h.service.GetUserUrls(ctx, h.baseURL)
	if err != nil {
		h.Logger.Error("Failed to retrieve user URLs",
			zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve user URLs")
	}

	// Если нет данных — возвращаем пустой ответ
	if len(urls) == 0 {
		h.Logger.Debug("No URLs found for user")
		return &pb.UserURLsResponse{Urls: []*pb.UserURL{}}, nil
	}

	// Формирование ответа
	var pbUrls []*pb.UserURL
	for _, u := range urls {
		pbUrls = append(pbUrls, &pb.UserURL{
			ShortUrl:    u.ShortURL,
			OriginalUrl: u.OriginalURL,
		})
	}

	h.Logger.Debug("Successfully returned user URLs", zap.Int("count", len(pbUrls)))

	return &pb.UserURLsResponse{
		Urls: pbUrls,
	}, nil
}
