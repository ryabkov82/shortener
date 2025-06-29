package testhandlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/service"

	"github.com/ryabkov82/shortener/internal/app/logger"

	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/ryabkov82/shortener/internal/app/handlers/batch"
	"github.com/ryabkov82/shortener/internal/app/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// TestBatchHandler тестирует обработчик пакетного создания сокращённых URL (/api/shorten/batch).
//
// Проверяет следующие сценарии:
//   - Успешное создание сокращённых URL для пакета ссылок
//   - Обработку некорректных входных данных
//   - Работу сжатия gzip на входящих и исходящих данных
//   - Авторизацию через JWT cookie
//   - Соответствие форматов запроса и ответа API
//
// Тест создаёт полноценное HTTP-окружение с:
//   - Роутером Chi
//   - Middleware: логирование, gzip, JWT-авторизация
//   - Тестовым сервером httptest
//   - Поддержкой сжатия на клиенте и сервере
//
// Примеры тест-кейсов:
//   - Корректный запрос с валидными URL (ожидается 201 Created)
//   - Некорректный JSON (ожидается 400 Bad Request)
//
// Используемые компоненты:
//   - service.Service: бизнес-логика сервиса
//   - logger: система логирования
//   - testutils: утилиты для генерации тестовых данных
//   - resty: HTTP-клиент для тестирования
func TestBatch(t *testing.T, st service.Repository) {

	service := service.NewService(st)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.JWTAutoIssue(testutils.TestSecretKey))

	baseURL := "http://localhost:8080/"
	r.Post("/api/shorten/batch", batch.GetHandler(service, baseURL, logger.Log))

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
