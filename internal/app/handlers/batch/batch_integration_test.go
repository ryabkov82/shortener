package batch

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/service"

	"github.com/ryabkov82/shortener/internal/app/logger"

	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
	"github.com/ryabkov82/shortener/internal/testutils"

	"github.com/ryabkov82/shortener/internal/app/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetHandler_InMemory(t *testing.T) {

	st, err := testutils.InitializeInMemoryStorage()
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	defer os.Remove(st.FilePath())

	testBatch(t, st)
}

func TestGetHandler_Postgres(t *testing.T) {

	dsn := os.Getenv("TEST_DB_DSN")

	if dsn == "" {
		t.Fatal("TEST_DB_DSN не установлен")
	}
	pg, err := postgres.NewPostgresStorage(dsn)

	if err != nil {
		t.Fatal(err)
	}

	testBatch(t, pg)
}

func testBatch(t *testing.T, st service.Repository) {

	service := service.NewService(st)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.JWTAutoIssue(testutils.TestSecretKey))

	baseURL := "http://localhost:8080/"
	r.Post("/api/shorten/batch", GetHandler(service, baseURL, logger.Log))

	cookie, _ := testutils.CreateSignedCookie()

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()
	tests := []struct {
		cookie         *http.Cookie
		name           string
		request        string
		wantStatusCode int
	}{
		{
			name: "positive test #1",
			request: `[
									{
										"correlation_id": "123",
										"original_url": "https://example.com/page1"
									},
									{
										"correlation_id": "456",
										"original_url": "https://example.com/page2"
									}
								]`,
			cookie:         cookie,
			wantStatusCode: 201,
		},
		{
			name:           "negative test #2",
			request:        "{}",
			cookie:         cookie,
			wantStatusCode: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := bytes.NewBuffer(nil)
			zb := gzip.NewWriter(buf)
			_, err := zb.Write([]byte(tt.request))
			assert.NoError(t, err)
			err = zb.Close()
			assert.NoError(t, err)

			resp, err := resty.New().R().
				SetBody(buf).
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Accept-Encoding", "gzip").
				Post(srv.URL + "/api/shorten/batch")

			assert.NoError(t, err)

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())
			if tt.wantStatusCode == 201 {
				// проверим, что получили данные в нужном формате
				var response []models.BatchResponse
				err = json.Unmarshal(resp.Body(), &response)
				assert.NoError(t, err)
			}

		})
	}

}
