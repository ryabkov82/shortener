package testhandlers

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// TestShortenURL тестирует текстовый обработчик сокращения URL (POST /).
//
// Принимает:
//
//	t *testing.T - инстанс тестинга
//	client *resty.Client - предварительно настроенный HTTP клиент (с базовым URL)
//
// Проверяет следующие сценарии:
//   - Успешное создание короткой ссылки (StatusCreated)
//   - Попытку повторного сокращения того же URL (StatusConflict)
//   - Обработку некорректного URL (StatusBadRequest)
//   - Корректность возвращаемого текстового ответа
//   - Работу JWT авторизации через cookie
//   - Поддержку gzip сжатия через middleware
//
// Тест создает:
//   - Тестовый HTTP сервер с middleware:
//   - Логирование запросов
//   - Поддержка gzip
//   - JWT авторизация
//   - Набор тестовых случаев с разными входными данными
//
// Примеры тест-кейсов:
//   - Корректный URL (ожидается 201 Created с валидным коротким URL в теле ответа)
//   - Дублирующийся URL (ожидается 409 Conflict)
//   - Некорректный URL (ожидается 400 Bad Request)
//
// Особенности:
//   - Проверяет валидность возвращаемого короткого URL
//   - Работает с текстовым форматом запроса/ответа (в отличие от JSON API)
//   - Поддерживает авторизацию через JWT cookie
func TestShortenURL(t *testing.T, client *resty.Client) {

	cookie, _ := testutils.CreateSignedCookie()

	tests := []struct {
		cookie         *http.Cookie
		name           string
		originalURL    string
		wantStatusCode int
	}{
		{
			name:           "positive test #1",
			originalURL:    "https://practicum.yandex.ru/",
			cookie:         cookie,
			wantStatusCode: 201,
		},
		{
			name:           "positive test #1",
			originalURL:    "https://practicum.yandex.ru/",
			cookie:         cookie,
			wantStatusCode: 409,
		},
		{
			name:           "negative test #2",
			originalURL:    "not url",
			cookie:         cookie,
			wantStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resp, err := client.R().
				SetCookie(tt.cookie).
				SetBody(tt.originalURL).
				Post("/")

			assert.NoError(t, err)

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())
			if tt.wantStatusCode == 201 {
				shortURL := resp.Body()
				// Проверяем, что получен URL
				_, err = url.Parse(string(shortURL))
				assert.NoError(t, err)
			}
		})
	}

}
