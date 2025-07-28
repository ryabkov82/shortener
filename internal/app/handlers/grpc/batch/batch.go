package batch

import (
	"context"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/base"
	"github.com/ryabkov82/shortener/internal/app/models"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// URLHandler определяет контракт для обработки пакетных запросов.
type URLHandler interface {
	Batch(ctx context.Context, requests []models.BatchRequest, baseURL string) ([]models.BatchResponse, error)
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

func (h *Handler) BatchCreate(
	ctx context.Context,
	req *pb.BatchCreateRequest,
) (*pb.BatchCreateResponse, error) {
	if len(req.Items) == 0 {
		h.Logger.Error("Empty batch request")
		return nil, status.Error(codes.InvalidArgument, "Request contains no items")
	}

	// Преобразуем в []models.BatchRequest
	batchReq := make([]models.BatchRequest, 0, len(req.Items))
	for _, item := range req.Items {
		batchReq = append(batchReq, models.BatchRequest{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
		})
	}

	// Обработка
	batchResp, err := h.service.Batch(ctx, batchReq, h.baseURL)
	if err != nil {
		h.Logger.Error("Failed to process batch create", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to process batch create")
	}

	// Формируем ответ
	resp := &pb.BatchCreateResponse{}
	for _, item := range batchResp {
		resp.Items = append(resp.Items, &pb.BatchCreateResult{
			CorrelationId: item.CorrelationID,
			ShortUrl:      item.ShortURL,
		})
	}

	h.Logger.Debug("Batch create processed", zap.Int("count", len(resp.Items)))
	return resp, nil
}
