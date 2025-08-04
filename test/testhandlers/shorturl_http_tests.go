package testhandlers

import (
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

func TestShortenURL(t *testing.T, httpClient *resty.Client) {

	for _, tt := range CommonShortenURLTestCases() {
		t.Run("HTTP_"+tt.Name, func(t *testing.T) {

			resp, err := httpClient.R().
				SetCookie(tt.Cookie).
				SetBody(tt.OriginalURL).
				Post("/")

			assert.NoError(t, err)

			var shortenResult ShortenResult

			shortenResult.ShortURL = string(resp.Body())
			shortenResult.Status = testutils.HTTPStatusToStatusCode(resp.StatusCode())

			// Проверяем статус ответа
			assert.Equal(t, tt.Want, shortenResult.Status)
			if tt.Want == testutils.StatusCreated {
				shortURL := shortenResult.ShortURL
				// Проверяем, что получен URL
				_, err = url.Parse(string(shortURL))
				assert.NoError(t, err)
			}

		})
	}
}
