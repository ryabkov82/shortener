package testhandlers

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"

	"github.com/ryabkov82/shortener/internal/app/models"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// TestRedirect тестирует обработчик редиректа по короткой ссылке (GET /{id}).
//
// Проверяет следующие сценарии:
//   - Успешный редирект на оригинальный URL (StatusTemporaryRedirect)
//   - Обработку несуществующего короткого URL (StatusNotFound)
//   - Корректность заголовка Location при редиректе
//   - Работу JWT авторизации через cookie
//   - Обработку gzip сжатия через middleware
//
// Тест создает:
//   - Тестовое хранилище с предзаполненным URL
//   - HTTP-сервер с middleware:
//   - Логирование запросов
//   - Поддержка gzip
//   - JWT авторизация
//   - Специальную политику обработки редиректов для resty
//
// Примеры тест-кейсов:
//   - Запрос существующего URL (ожидается 307 с правильным Location)
//   - Запрос несуществующего URL (ожидается 404)
//
// Особенности:
//   - Использует кастомную политику редиректов для проверки Location
//   - Сохраняет тестовые данные перед выполнением тестов
//   - Проверяет как положительные, так и отрицательные сценарии
func TestRedirect(t *testing.T, repo service.Repository, client *resty.Client) {

	mapping := models.URLMapping{
		ShortURL:    "EYm7J2zF",
		OriginalURL: "https://practicum.yandex.ru/",
	}

	var redirectAttemptedError = errors.New("redirect")
	redirectPolicy := resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		// return nil for continue redirect otherwise return error to stop/prevent redirect
		return redirectAttemptedError
	})

	cookie, userID := testutils.CreateSignedCookie()
	ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, userID)
	repo.SaveURL(ctx, &mapping)

	tests := []struct {
		cookie         *http.Cookie
		name           string
		originalURL    string
		shortKey       string
		wantStatusCode int
	}{
		{
			name:           "positive test #1",
			originalURL:    "https://practicum.yandex.ru/",
			shortKey:       "EYm7J2zF",
			cookie:         cookie,
			wantStatusCode: 307,
		},
		{
			name:           "negative test #2",
			shortKey:       "RrixjW0q",
			cookie:         cookie,
			wantStatusCode: 404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client.SetRedirectPolicy(redirectPolicy)
			req := client.R().SetCookie(tt.cookie)
			req.Method = http.MethodGet
			req.URL = "/" + tt.shortKey

			resp, err := req.Send()

			if errors.Is(err, redirectAttemptedError) {
				// эту ошибку игнорируем
				err = nil
			}

			assert.NoError(t, err)

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())
			if tt.wantStatusCode == 307 {
				assert.Equal(t, tt.originalURL, resp.Header().Get("Location"))
			}
		})
	}

}
