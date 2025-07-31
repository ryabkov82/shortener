package grpcserver

import (
	"net"

	"github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/config"
	grpchandlers "github.com/ryabkov82/shortener/internal/app/handlers/grpc"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/base"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/batch"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/deluserurls"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/ping"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/redirect"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/shorturl"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/stats"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/userurls"
	"github.com/ryabkov82/shortener/internal/app/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// StartGRPCServer создает и запускает gRPC сервер
func StartGRPCServer(log *zap.Logger, cfg *config.Config, srv *service.Service) *grpc.Server {

	// Создаем базовый обработчик с общими зависимостями
	baseHandler := base.NewBaseHandler(log)

	// Инициализация конкретных обработчиков
	shorturlHandler := shorturl.New(
		baseHandler,
		srv,
		cfg.BaseURL,
	)

	redirectHandler := redirect.New(
		baseHandler,
		srv,
	)

	batchHandler := batch.New(
		baseHandler,
		srv,
		cfg.BaseURL,
	)

	deluserurlsHandler := deluserurls.New(
		baseHandler,
		srv,
	)

	userurlsHandler := userurls.New(
		baseHandler,
		srv,
		cfg.BaseURL,
	)

	statsHandler := stats.New(
		baseHandler,
		srv,
	)

	pingHandler := ping.New(
		baseHandler,
		srv,
	)

	// Создаем агрегированный сервер
	aggregateHandler := grpchandlers.NewServer(
		baseHandler,
		grpchandlers.WithCreateShortURLEndpoint(shorturlHandler),
		grpchandlers.WithGetOriginalURLEndpoint(redirectHandler),
		grpchandlers.WithBatchCreateEndpoint(batchHandler),
		grpchandlers.WithDeleteUserURLsEndpoint(deluserurlsHandler),
		grpchandlers.WithGetUserURLsEndpoint(userurlsHandler),
		grpchandlers.WithGetStatsEndpoint(statsHandler),
		grpchandlers.WithPingEndpoint(pingHandler),
	)

	commonInterceptors := baseHandler.CommonInterceptors(cfg)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(commonInterceptors...),
	)

	// Регистрация gRPC сервиса
	api.RegisterShortenerServer(grpcServer, aggregateHandler)

	lis, err := net.Listen("tcp", cfg.GRPCServerAddr)
	if err != nil {
		log.Fatal("Failed to listen gRPC", zap.Error(err))
	}

	go func() {
		log.Info("Starting gRPC server", zap.String("address", cfg.GRPCServerAddr))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server failed", zap.Error(err))
		}
	}()

	return grpcServer
}
