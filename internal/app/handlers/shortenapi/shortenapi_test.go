package shortenapi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/logger"

	"github.com/ryabkov82/shortener/internal/app/service"

	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
	"github.com/ryabkov82/shortener/internal/testutils"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func testShortenApi(t *testing.T, st service.Repository) {

	service := service.NewService(st)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.JWTAutoIssue(testutils.TestSecretKey))

	cookie, _ := testutils.CreateSignedCookie()

	baseURL := "http://localhost:8080/"
	r.Post("/api/shorten", GetHandler(service, baseURL, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	tests := []struct {
		cookie         *http.Cookie
		name           string
		request        Request
		wantStatusCode int
	}{
		{
			name:           "positive test #1",
			request:        Request{URL: "https://practicum.yandex.ru/"},
			cookie:         cookie,
			wantStatusCode: 201,
		},
		{
			name:           "positive test #2",
			request:        Request{URL: "https://practicum.yandex.ru/"},
			cookie:         cookie,
			wantStatusCode: 409,
		},
		{
			name:           "negative test #2",
			request:        Request{URL: "not url"},
			cookie:         cookie,
			wantStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			buf := bytes.NewBuffer(nil)
			zb := gzip.NewWriter(buf)
			_, err = zb.Write([]byte(req))
			assert.NoError(t, err)
			err = zb.Close()
			assert.NoError(t, err)

			resp, err := resty.New().R().
				SetCookie(tt.cookie).
				SetBody(buf).
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Accept-Encoding", "gzip").
				Post(srv.URL + "/api/shorten")

			assert.NoError(t, err)

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())
			if tt.wantStatusCode == 201 {
				var response Response
				err = json.Unmarshal(resp.Body(), &response)
				assert.NoError(t, err)
				shortURL := response.Result
				// Проверяем, что получен URL
				_, err = url.Parse(shortURL)
				assert.NoError(t, err)
			}
		})
	}

}

func TestGetHandler_InMemory(t *testing.T) {

	st, err := testutils.InitializeInMemoryStorage()

	if err != nil {
		t.Fatal(err)
	}

	testShortenApi(t, st)
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

	testShortenApi(t, pg)
}
