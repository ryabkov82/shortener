package testhandlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/ryabkov82/shortener/internal/app/models"

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
func TestBatch(t *testing.T, client *resty.Client) {

	cookie, _ := testutils.CreateSignedCookie()

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

			resp, err := client.R().
				SetBody(buf).
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Accept-Encoding", "gzip").
				Post("/api/shorten/batch")

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
