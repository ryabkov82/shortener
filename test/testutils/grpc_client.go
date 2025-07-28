package testutils

import (
	"context"
	"fmt"
	"net"
	"time"

	pb "github.com/ryabkov82/shortener/api"

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
//
// Возвращает:
//   - *TestGRPCClient - готовый к использованию тестовый клиент с сервером
func NewTestGRPCClient(
	interceptors []grpc.UnaryServerInterceptor,
	service pb.ShortenerServer,
) *TestGRPCClient {

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
		panic(err)
	}

	// 1. Явно инициируем соединение
	conn.Connect()

	// 2. Проверяем соединение с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 3. Механизм проверки готовности соединения
	connectionReady := make(chan struct{})
	go func() {
		for {
			state := conn.GetState()
			if state == connectivity.Ready {
				close(connectionReady)
				return
			}

			// Ждем изменения состояния
			if !conn.WaitForStateChange(ctx, state) {
				return // Таймаут или отмена контекста
			}
		}
	}()

	// Ждем либо готовности, либо таймаута
	select {
	case <-connectionReady:
		// Соединение готово к работе
		fmt.Println("Connection established successfully")
	case <-ctx.Done():
		panic("connection timeout exceeded: server not responding")
	}

	return &TestGRPCClient{
		Conn:   conn,
		Lis:    lis,
		Server: srv,
	}
}

// Close освобождает ресурсы тестового клиента.
func (tc *TestGRPCClient) Close() {
	tc.Conn.Close()
	tc.Server.Stop()
	tc.Lis.Close()
}
