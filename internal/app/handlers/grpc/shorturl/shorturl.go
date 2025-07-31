package shorturl

import (
	"context"
	"errors"
	"net/url"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/base"
	"github.com/ryabkov82/shortener/internal/app/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// URLHandler определяет контракт для генерации коротких URL.
type URLHandler interface {
	// GetShortKey возвращает короткий ключ для оригинального URL.
	//
	// Параметры:
	//   ctx - контекст выполнения (должен включать таймаут)
	//   originalURL - валидный URL для сокращения
	//
	// Возвращает:
	//   string - короткий ключ
	//   error - возможные ошибки:
	//     - storage.ErrURLExists: URL уже существует
	//     - другие внутренние ошибки
	GetShortKey(ctx context.Context, originalURL string) (string, error)
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

// CreateShortURL обработчик gRPC для создания короткой ссылки
func (h *Handler) CreateShortURL(
	ctx context.Context,
	req *pb.CreateRequest,
) (*pb.CreateResponse, error) {
	// Валидация URL
	if req.OriginalUrl == "" {
		h.Logger.Error("Empty URL in request")
		return nil, status.Error(codes.InvalidArgument, "URL parameter is missing")
	}

	if _, err := url.ParseRequestURI(req.OriginalUrl); err != nil {
		h.Logger.Error("Invalid URL in request",
			zap.String("url", req.OriginalUrl),
			zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, "Invalid URL format")
	}

	h.Logger.Debug("Processing URL shortening",
		zap.String("originalURL", req.OriginalUrl))

	// Генерация короткого ключа
	shortKey, err := h.service.GetShortKey(ctx, req.OriginalUrl)
	if err != nil && !errors.Is(err, storage.ErrURLExists) {
		h.Logger.Error("Short URL generation failed",
			zap.Error(err),
			zap.String("originalURL", req.OriginalUrl))
		return nil, status.Error(codes.Internal, "Failed to generate short URL")
	}

	// Формирование ответа
	response := &pb.CreateResponse{
		ShortUrl: h.baseURL + "/" + shortKey,
	}

	if errors.Is(err, storage.ErrURLExists) {
		h.Logger.Debug("URL already exists",
			zap.String("shortKey", shortKey),
			zap.String("originalURL", req.OriginalUrl))
		return response, status.Error(codes.AlreadyExists, "URL already exists")
	}

	h.Logger.Debug("URL successfully shortened",
		zap.String("shortKey", shortKey),
		zap.String("originalURL", req.OriginalUrl))

	return response, nil
}
