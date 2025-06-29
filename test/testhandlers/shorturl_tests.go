package testhandlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/handlers/shorturl"
	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// TestShortenURL тестирует текстовый обработчик сокращения URL (POST /).
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
func TestShortenURL(t *testing.T, st service.Repository) {

	service := service.NewService(st)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.JWTAutoIssue(testutils.TestSecretKey))

	baseURL := "http://localhost:8080/"
	r.Post("/", shorturl.GetHandler(service, baseURL, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

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

			resp, err := resty.New().R().
				SetCookie(tt.cookie).
				SetBody(tt.originalURL).
				Post(srv.URL)

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
