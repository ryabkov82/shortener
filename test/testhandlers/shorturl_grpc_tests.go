package testhandlers

import (
	"context"
	"net/url"
	"testing"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/status"
)

func TestShortenURLGRPC(t *testing.T, grpcClient pb.ShortenerClient) {

	for _, tt := range CommonShortenURLTestCases() {
		t.Run("gRPC_"+tt.Name, func(t *testing.T) {

			token := tt.Cookie.Value
			ctx := testutils.ContextWithJWT(context.Background(), token)

			resp, err := grpcClient.CreateShortURL(ctx, &pb.CreateRequest{OriginalUrl: tt.OriginalURL})
			var shortenResult ShortenResult
			if resp != nil {
				shortenResult.ShortURL = resp.ShortUrl
				shortenResult.Status = testutils.StatusCreated
			}

			if err != nil {
				if s, ok := status.FromError(err); ok {
					shortenResult.Status = testutils.GRPCCodeToStatusCode(s.Code())
				} else {
					shortenResult.Status = testutils.StatusInternalError
				}
			}

			assert.Equal(t, tt.Want, shortenResult.Status)

			if tt.Want == testutils.StatusCreated {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				_, err = url.Parse(shortenResult.ShortURL)
				assert.NoError(t, err)
			}
		})
	}
}
