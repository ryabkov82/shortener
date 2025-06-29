package testhandlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/handlers/shortenapi"
	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// TestShortenAPI тестирует JSON API обработчик сокращения URL (POST /api/shorten).
//
// Проверяет следующие сценарии:
//   - Успешное создание короткой ссылки (StatusCreated)
//   - Попытку повторного сокращения того же URL (StatusConflict)
//   - Обработку некорректного URL (StatusBadRequest)
//   - Корректность формата JSON ответа
//   - Работу JWT авторизации через cookie
//   - Поддержку gzip сжатия запросов и ответов
//
// Тест создает:
//   - Тестовый HTTP сервер с middleware:
//   - Логирование запросов
//   - Поддержка gzip
//   - JWT авторизация
//   - Набор тестовых случаев с разными входными данными
//
// Примеры тест-кейсов:
//   - Корректный URL (ожидается 201 Created с валидным коротким URL)
//   - Дублирующийся URL (ожидается 409 Conflict)
//   - Некорректный URL (ожидается 400 Bad Request)
//
// Особенности:
//   - Проверяет валидность возвращаемого короткого URL
//   - Использует gzip сжатие для запросов
//   - Поддерживает авторизацию через JWT cookie
func TestShortenAPI(t *testing.T, client *resty.Client) {

	cookie, _ := testutils.CreateSignedCookie()
	tests := []struct {
		cookie         *http.Cookie
		name           string
		request        shortenapi.Request
		wantStatusCode int
	}{
		{
			name:           "positive test #1",
			request:        shortenapi.Request{URL: "https://practicum.yandex.ru/"},
			cookie:         cookie,
			wantStatusCode: 201,
		},
		{
			name:           "positive test #2",
			request:        shortenapi.Request{URL: "https://practicum.yandex.ru/"},
			cookie:         cookie,
			wantStatusCode: 409,
		},
		{
			name:           "negative test #2",
			request:        shortenapi.Request{URL: "not url"},
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

			resp, err := client.R().
				SetCookie(tt.cookie).
				SetBody(buf).
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Accept-Encoding", "gzip").
				Post("/api/shorten")

			assert.NoError(t, err)

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())
			if tt.wantStatusCode == 201 {
				var response shortenapi.Response
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
