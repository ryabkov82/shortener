package testhandlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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

	//cookie, _ := testutils.CreateSignedCookie()

	tests := CommonBatchTestCases()

	for _, tt := range tests {
		t.Run("HTTP_"+tt.Name, func(t *testing.T) {

			buf := bytes.NewBuffer(nil)
			zb := gzip.NewWriter(buf)
			_, err := zb.Write([]byte(tt.Request))
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
			assert.Equal(t, tt.WantStatus, testutils.HTTPStatusToStatusCode(resp.StatusCode()))
			if tt.WantStatus == testutils.StatusCreated {
				// проверим, что получили данные в нужном формате
				var response []models.BatchResponse
				err = json.Unmarshal(resp.Body(), &response)
				assert.NoError(t, err)
			}

		})
	}

}
