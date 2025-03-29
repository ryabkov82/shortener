package ping

import (
	"net/http/httptest"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/logger"

	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/storage/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetHandler(t *testing.T) {

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	tests := []struct {
		name           string
		DbConnect      string
		wantStatusCode int
	}{
		{
			name:           "positive test #1",
			DbConnect:      "host=localhost port=5432 user=shortener password=shortener dbname=shortener sslmode=disable",
			wantStatusCode: 200,
		},
		{
			name:           "negative test #2",
			DbConnect:      "",
			wantStatusCode: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			pg, err := postgres.NewPostgresStorage(tt.DbConnect)

			if err != nil {
				panic(err)
			}

			r := chi.NewRouter()
			r.Use(mwlogger.RequestLogging(logger.Log))
			r.Use(mwgzip.Gzip)

			r.Get("/ping", GetHandler(pg, logger.Log))

			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(r)
			// останавливаем сервер после завершения теста
			defer srv.Close()

			resp, err := resty.New().R().
				Get(srv.URL + "/ping")

			assert.NoError(t, err)

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())

		})
	}

}
