package testutils

import (
	"context"
	"fmt"
	"net"
	"time"

	pb "github.com/ryabkov82/shortener/api"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// TestGRPCClient представляет тестовый gRPC клиент с ассоциированным сервером.
type TestGRPCClient struct {
	Conn   *grpc.ClientConn
	Server *grpc.Server
	Lis    *bufconn.Listener
}

// NewTestGRPCClient создает новое тестовое окружение для gRPC тестов.
//
// Параметры:
//   - interceptors: список gRPC интерцепторов
//   - service: реализация gRPC сервера
//   - logger: логгер для записи событий
//
// Возвращает:
//   - *TestGRPCClient - готовый к использованию тестовый клиент с сервером
//   - error - ошибка инициализации
func NewTestGRPCClient(
	interceptors []grpc.UnaryServerInterceptor,
	service pb.ShortenerServer,
	logger *zap.Logger,
) (*TestGRPCClient, error) {

	// Создаем виртуальное соединение
	lis := bufconn.Listen(1024 * 1024)

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...),
	)

	// Регистрируем сервисы
	pb.RegisterShortenerServer(srv, service)

	// Канал для ожидания запуска сервера
	serverErr := make(chan error, 1) // Буферизированный канал
	ready := make(chan struct{})

	go func() {
		close(ready) // Сообщаем, что вот-вот будет вызван Serve
		if err := srv.Serve(lis); err != nil {
			serverErr <- err
		}
		close(serverErr)
	}()

	<-ready // Дождались запуска горутины

	// Создаем dial функцию для bufconn
	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return lis.DialContext(ctx)
	}

	conn, err := grpc.NewClient(
		"passthrough:///inmem",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// 1. Явно инициируем соединение
	conn.Connect()

	// 3. Проверяем готовность с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := waitForConnectionReady(ctx, conn, logger); err != nil {
		// Создаем контекст с таймаутом 2 секунды для всего shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel() // Важно вызывать cancel для освобождения ресурсов

		// Канал для отслеживания завершения GracefulStop
		gracefulDone := make(chan struct{})

		// Запускаем graceful shutdown в отдельной горутине
		go func() {
			defer close(gracefulDone)
			srv.GracefulStop() // Пытаемся остановиться корректно
		}()

		// Ожидаем либо успешного завершения, либо таймаута
		select {
		case <-gracefulDone:
			logger.Info("Server stopped gracefully")
		case <-shutdownCtx.Done():
			logger.Warn("Forcing shutdown after timeout")
		}

		// Всегда закрываем ресурсы
		lis.Close()
		conn.Close()

		return nil, fmt.Errorf("connection failed: %w", err)
	}

	logger.Debug("gRPC test environment initialized successfully")

	return &TestGRPCClient{
		Conn:   conn,
		Lis:    lis,
		Server: srv,
	}, nil
}

// Close освобождает ресурсы тестового клиента.
func (tc *TestGRPCClient) Close() {
	tc.Conn.Close()
	tc.Server.Stop()
	tc.Lis.Close()
}

// waitForConnectionReady ожидает готовности соединения с таймаутом
func waitForConnectionReady(ctx context.Context, conn *grpc.ClientConn, logger *zap.Logger) error {
	connectionReady := make(chan struct{})
	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				state := conn.GetState()
				if state == connectivity.Ready {
					close(connectionReady)
					return
				}
				if !conn.WaitForStateChange(ctx, state) {
					return
				}
			}
		}
	}()

	select {
	case <-connectionReady:
		logger.Debug("gRPC connection established",
			zap.String("state", conn.GetState().String()))
		return nil
	case <-ctx.Done():
		return fmt.Errorf("connection timeout: last state %s", conn.GetState())
	}
}
