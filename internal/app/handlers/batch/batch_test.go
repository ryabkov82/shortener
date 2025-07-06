package batch_test

import (
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/service/mocks"

	"github.com/ryabkov82/shortener/internal/app/handlers/batch"
	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/test/testhandlers"
	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/go-chi/chi/v5"

	"github.com/golang/mock/gomock"
)

func TestGetHandler(t *testing.T) {

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockRepository(ctrl)

	m.EXPECT().GetExistingURLs(gomock.Any(), gomock.Any()).Return(nil, nil)
	m.EXPECT().SaveNewURLs(gomock.Any(), gomock.Any()).Return(nil)

	service := service.NewService(m)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	baseURL := "http://localhost:8080/"

	tc := testutils.NewTestClient(func(r chi.Router) {
		r.Use(mwlogger.RequestLogging(logger.Log))
		r.Use(mwgzip.Gzip)
		r.Use(auth.JWTAutoIssue(testutils.TestSecretKey))

		r.Post("/api/shorten/batch", batch.GetHandler(service, baseURL, logger.Log))
	})
	defer tc.Close()

	testhandlers.TestBatch(t, tc.Client)
}

func TestGetHandler_InMemory(t *testing.T) {

	st, err := testutils.InitializeInMemoryStorage()
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	defer os.Remove(st.FilePath())

	service := service.NewService(st)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	baseURL := "http://localhost:8080/"

	tc := testutils.NewTestClient(func(r chi.Router) {
		r.Use(mwlogger.RequestLogging(logger.Log))
		r.Use(mwgzip.Gzip)
		r.Use(auth.JWTAutoIssue(testutils.TestSecretKey))

		r.Post("/api/shorten/batch", batch.GetHandler(service, baseURL, logger.Log))
	})
	defer tc.Close()

	testhandlers.TestBatch(t, tc.Client)
}
