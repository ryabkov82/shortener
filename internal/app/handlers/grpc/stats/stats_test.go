package stats_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/config"
	grpchandlers "github.com/ryabkov82/shortener/internal/app/handlers/grpc"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/base"
	"github.com/ryabkov82/shortener/internal/app/handlers/grpc/stats"
	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/service/mocks"
	"github.com/ryabkov82/shortener/test/testutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGRPCGetStats(t *testing.T) {

	// Инициализация логгера
	if err := logger.Initialize("debug"); err != nil {
		t.Fatalf("logger initialization failed: %v", err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	serv := service.NewService(repo)

	tests := []struct {
		name           string
		ip             string
		trustedSubnet  string
		wantCode       codes.Code
		urls           int
		users          int
		countURLErr    error
		countUserErr   error
		expectResponse *pb.StatsResponse
	}{
		{
			name:           "positive test - allowed IP",
			ip:             "192.168.1.100",
			trustedSubnet:  "192.168.1.0/24",
			urls:           10,
			users:          5,
			wantCode:       codes.OK,
			expectResponse: &pb.StatsResponse{Urls: 10, Users: 5},
		},
		{
			name:          "denied IP - not in subnet",
			ip:            "10.0.0.1",
			trustedSubnet: "192.168.1.0/24",
			wantCode:      codes.PermissionDenied,
		},
		{
			name:          "allowed IP - CountURLs error",
			ip:            "192.168.1.100",
			trustedSubnet: "192.168.1.0/24",
			countURLErr:   errors.New("db error"),
			wantCode:      codes.Internal,
		},
		{
			name:          "allowed IP - CountUsers error",
			ip:            "192.168.1.100",
			trustedSubnet: "192.168.1.0/24",
			urls:          5,
			countUserErr:  errors.New("db error"),
			wantCode:      codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.wantCode != codes.PermissionDenied {
				repo.EXPECT().
					CountURLs(gomock.Any()).
					Return(tt.urls, tt.countURLErr).
					Times(1)

				if tt.countURLErr == nil {
					repo.EXPECT().
						CountUsers(gomock.Any()).
						Return(tt.users, tt.countUserErr).
						Times(1)
				}
			}

			baseHandler := &base.BaseHandler{Logger: logger.Log}

			cfg := &config.Config{TrustedSubnet: tt.trustedSubnet, JwtKey: string(testutils.TestSecretKey)}

			commonInterceptors := baseHandler.CommonInterceptors(cfg)

			statsHandler := stats.New(
				baseHandler,
				serv,
			)
			// Создаем тестовый клиент
			tc, err := testutils.NewTestGRPCClient(
				commonInterceptors,
				grpchandlers.NewServer(
					baseHandler,
					grpchandlers.WithGetStatsEndpoint(statsHandler),
				),
				logger.Log,
			)
			if err != nil {
				t.Fatal(err)
			}

			defer tc.Close()

			// Создаем клиент
			client := pb.NewShortenerClient(tc.Conn)

			md := metadata.New(map[string]string{
				"x-real-ip": tt.ip,
			})
			ctx := metadata.NewOutgoingContext(context.Background(), md)

			resp, err := client.GetStats(ctx, &pb.StatsRequest{})

			st, _ := status.FromError(err)
			assert.Equal(t, tt.wantCode, st.Code())

			if tt.wantCode == codes.OK {
				assert.Equal(t, tt.expectResponse.Urls, resp.Urls)
				assert.Equal(t, tt.expectResponse.Users, resp.Users)
			}

		})
	}
}
