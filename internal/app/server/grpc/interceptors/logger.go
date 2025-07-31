package interceptors

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor возвращает gRPC-интерцептор для логирования вызовов методов.
//
// Интерцептор фиксирует:
//   - Полное имя вызываемого метода (info.FullMethod)
//   - Параметры запроса (req)
//   - Статус обработки (код и текстовое описание)
//   - Время выполнения вызова
//
// Параметры:
//   - log: логгер zap для записи сообщений (должен быть предварительно настроен)
//
// Возвращает:
//   - grpc.UnaryServerInterceptor: настроенный интерцептор логирования
//
// Особенности работы:
//   - Логирование происходит после выполнения основного обработчика
//   - Время выполнения измеряется с наносекундной точностью
//   - Ошибки преобразуются в gRPC-статусы для единообразного логирования
//   - Запросы логируются на уровне INFO
//
// Пример использования:
//
//	logger, _ := zap.NewProduction()
//	server := grpc.NewServer(
//	    grpc.ChainUnaryInterceptor(
//	        LoggingInterceptor(logger),
//	    ),
//	)
//
// Формат логов:
//
//	{
//	  "method": "/service.Name/MethodName",
//	  "request": {...},
//	  "status_code": 2,
//	  "status": "UNKNOWN",
//	  "duration": "12.345ms"
//	}
func LoggingInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		// Фиксируем время начала обработки
		startTime := time.Now()

		// Отложенное логирование после обработки
		defer func() {
			duration := time.Since(startTime)
			st, _ := status.FromError(err)

			log.Info("gRPC request completed",
				zap.String("method", info.FullMethod),     // Полное имя метода (например /shortener.Shortener/CreateShortURL)
				zap.Any("request", req),                   // Входные параметры
				zap.Int("status_code", int(st.Code())),    // Код статуса gRPC
				zap.String("status", st.Code().String()),  // Текстовый статус
				zap.String("duration", duration.String()), // Время обработки
			)
		}()

		// Вызываем следующий обработчик
		return handler(ctx, req)
	}
}
