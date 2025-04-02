package batch

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/service/mocks"

	"github.com/ryabkov82/shortener/internal/app/logger"

	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"

	"github.com/ryabkov82/shortener/internal/app/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetHandler(t *testing.T) {

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockRepository(ctrl)

	m.EXPECT().GetExistingURLs(gomock.Any(), gomock.Any()).Return(nil, nil)
	m.EXPECT().SaveNewURLs(gomock.Any(), gomock.Any()).Return(nil)

	// инициализируем service объектом-заглушкой
	service := service.NewService(m)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)

	baseURL := "http://localhost:8080/"
	r.Post("/api/shorten/batch", GetHandler(service, baseURL, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()
	tests := []struct {
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
			wantStatusCode: 201,
		},
		{
			name:           "negative test #2",
			request:        "{}",
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
