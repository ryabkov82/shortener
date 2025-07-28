package ping_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/config"
	grpchandlers "github.com/ryabkov82/shortener/internal/app/handlers/grpc"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/base"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/ping"
	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/service/mocks"
	"github.com/ryabkov82/shortener/test/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGRPCPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем mock-репозиторий и сервис
	mockRepo := mocks.NewMockRepository(ctrl)
	service := service.NewService(mockRepo)

	// Инициализируем логгер
	require.NoError(t, logger.Initialize("debug"))

	baseHandler := &base.BaseHandler{Logger: logger.Log}

	cfg := &config.Config{JwtKey: string(testutils.TestSecretKey)}

	commonInterceptors := baseHandler.CommonInterceptors(cfg)

	pingHandler := ping.New(
		baseHandler,
		service,
	)

	// Создаем тестовый клиент
	tc := testutils.NewTestGRPCClient(
		commonInterceptors,
		grpchandlers.NewServer(
			baseHandler,
			grpchandlers.WithPingEndpoint(pingHandler),
		),
	)
	defer tc.Close()

	// Создаем клиент
	client := pb.NewShortenerClient(tc.Conn)

	tests := []struct {
		name         string
		mockError    error
		wantGRPCCode codes.Code
		wantMessage  string
	}{
		{
			name:         "positive test #1",
			mockError:    nil,
			wantGRPCCode: codes.OK,
			wantMessage:  "Connect to database is successful",
		},
		{
			name:         "negative test #2",
			mockError:    errors.New("db is down"),
			wantGRPCCode: codes.Internal,
			wantMessage:  "", // тело неважно, главное — код
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.EXPECT().
				Ping(gomock.Any()).
				Return(tt.mockError)

			_, err := client.Ping(context.Background(), &pb.PingRequest{})

			st, ok := status.FromError(err)
			if tt.wantGRPCCode == codes.OK {
				require.NoError(t, err)
				//assert.Equal(t, tt.wantMessage, resp.GetMessage())
			} else {
				require.True(t, ok)
				assert.Equal(t, tt.wantGRPCCode, st.Code())
			}
		})
	}
}
