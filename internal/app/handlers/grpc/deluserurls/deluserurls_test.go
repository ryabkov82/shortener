package deluserurls_test

import (
	"testing"

	pb "github.com/ryabkov82/shortener/api"
	grpchandlers "github.com/ryabkov82/shortener/internal/app/handlers/grpc"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/base"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/deluserurls"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/redirect"
	"github.com/ryabkov82/shortener/internal/app/server/grpc/interceptors"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/test/testhandlers"
	"github.com/ryabkov82/shortener/test/testutils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TestDelUserURLsGRPC(t *testing.T) {
	// Инициализация
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}

	st, err := testutils.InitializeInMemoryStorage()
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	serv := service.NewService(st)
	baseHandler := &base.BaseHandler{Logger: logger}

	interceptors := []grpc.UnaryServerInterceptor{
		interceptors.LoggingInterceptor(logger),
		interceptors.JWTAutoIssueGRPC(testutils.TestSecretKey, logger),
	}

	// Создаем тестовый клиент
	tc, err := testutils.NewTestGRPCClient(
		interceptors,
		grpchandlers.NewServer(
			baseHandler,
			grpchandlers.WithDeleteUserURLsEndpoint(deluserurls.New(baseHandler, serv)),
			grpchandlers.WithGetOriginalURLEndpoint(redirect.New(baseHandler, serv)),
		),
		logger,
	)

	if err != nil {
		t.Fatal(err)
	}

	defer tc.Close()

	// Создаем клиент
	client := pb.NewShortenerClient(tc.Conn)

	testhandlers.TestDelUserUrlsGRPC(t, serv, client)

}
