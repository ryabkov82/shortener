package base

import (
	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/server/grpc/interceptors"
	"google.golang.org/grpc"

	"go.uber.org/zap"
)

// BaseHandler содержит общие зависимости для всех обработчиков
type BaseHandler struct {
	Logger *zap.Logger // Общий логгер
}

// NewBaseHandler создает базовый обработчик
func NewBaseHandler(logger *zap.Logger) *BaseHandler {
	return &BaseHandler{
		Logger: logger,
	}
}

// CommonInterceptors возвращает цепочку общих интерцепторов
func (h *BaseHandler) CommonInterceptors(cfg *config.Config) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		interceptors.LoggingInterceptor(h.Logger),
		interceptors.JWTAutoIssueGRPC([]byte(cfg.JwtKey)),
		interceptors.TrustedSubnetInterceptor(interceptors.TrustedSubnetConfig{
			TrustedSubnet: cfg.TrustedSubnet,
			ProtectedMethods: map[string]bool{
				"/shortener.Shortener/GetStats": true,
			},
			DenyIfNotConfigured: true, // Блокировать если подсеть не настроена
		}),
	}

}
