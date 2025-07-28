package testhandlers

import (
	"context"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/models"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/status"
)

func TestRedirectGRPC(t *testing.T, repo service.Repository, grpcClient pb.ShortenerClient) {

	const (
		shortKey    = "EYm7J2zF"
		originalURL = "https://practicum.yandex.ru/"
	)

	mapping := models.URLMapping{
		ShortURL:    shortKey,
		OriginalURL: originalURL,
	}

	cookie, userID := testutils.CreateSignedCookie()
	ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, userID)
	repo.SaveURL(ctx, &mapping)

	tests := CommonRedirectTestCases(shortKey, originalURL)

	for _, tt := range tests {
		t.Run("gRPC_"+tt.Name, func(t *testing.T) {

			token := cookie.Value
			ctx := testutils.ContextWithJWT(context.Background(), token)

			resp, err := grpcClient.GetOriginalURL(ctx, &pb.GetRequest{ShortUrl: tt.ShortKey})

			var redirectStatus testutils.StatusCode
			if err != nil {
				if s, ok := status.FromError(err); ok {
					redirectStatus = testutils.GRPCCodeToStatusCode(s.Code())
				} else {
					redirectStatus = testutils.StatusInternalError
				}
			} else {
				redirectStatus = testutils.StatusTemporaryRedirect
			}

			// Проверяем статус ответа
			assert.Equal(t, tt.ExpectedStatus, redirectStatus)
			if tt.ExpectedStatus == testutils.StatusTemporaryRedirect {
				assert.NoError(t, err)
				assert.Equal(t, tt.ExpectedURL, resp.OriginalUrl)
			} else {
				assert.Error(t, err)
			}

		})
	}
}
