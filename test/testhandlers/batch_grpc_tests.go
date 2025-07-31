package testhandlers

import (
	"context"
	"encoding/json"
	"testing"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/test/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/status"
)

func TestBatchGRPC(t *testing.T, grpcClient pb.ShortenerClient) {

	tests := CommonBatchTestCases()

	for _, tt := range tests {
		t.Run("gRPC_"+tt.Name, func(t *testing.T) {
			var request []pb.BatchCreateItem
			err := json.Unmarshal([]byte(tt.Request), &request)
			if err != nil {
				// в случае некорректного JSON пропускаем
				t.Skip("invalid request format for gRPC")
			}

			var items []*pb.BatchCreateItem
			for i := range request {
				items = append(items, &request[i])
			}

			resp, err := grpcClient.BatchCreate(context.Background(), &pb.BatchCreateRequest{Items: items})

			if tt.WantStatus != testutils.StatusCreated {
				require.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tt.WantStatus, testutils.GRPCCodeToStatusCode(st.Code()))
				return
			}

			require.NoError(t, err)
			assert.Len(t, resp.Items, tt.ExpectedLength)

		})
	}
}
