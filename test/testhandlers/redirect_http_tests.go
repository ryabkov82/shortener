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

	const (
		shortKey    = "EYm7J2zF"
		originalURL = "https://practicum.yandex.ru/"
	)

	mapping := models.URLMapping{
		ShortURL:    shortKey,
		OriginalURL: originalURL,
	}

	var redirectAttemptedError = errors.New("redirect")
	redirectPolicy := resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		// return nil for continue redirect otherwise return error to stop/prevent redirect
		return redirectAttemptedError
	})

	cookie, userID := testutils.CreateSignedCookie()
	ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, userID)
	repo.SaveURL(ctx, &mapping)

	tests := CommonRedirectTestCases(shortKey, originalURL)

	for _, tt := range tests {
		t.Run("HTTP_"+tt.Name, func(t *testing.T) {

			client.SetRedirectPolicy(redirectPolicy)
			req := client.R().SetCookie(cookie)
			req.Method = http.MethodGet
			req.URL = "/" + tt.ShortKey

			resp, err := req.Send()

			if errors.Is(err, redirectAttemptedError) {
				// эту ошибку игнорируем
				err = nil
			}

			assert.NoError(t, err)

			// Проверяем статус ответа
			assert.Equal(t, tt.ExpectedStatus, testutils.HTTPStatusToStatusCode(resp.StatusCode()))
			if tt.ExpectedStatus == testutils.StatusTemporaryRedirect {
				assert.Equal(t, tt.ExpectedURL, resp.Header().Get("Location"))
			}
		})
	}

}
