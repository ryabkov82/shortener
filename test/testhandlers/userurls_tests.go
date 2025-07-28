package testhandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/models"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// TestUserUrls тестирует обработчик получения списка URL пользователя (GET /api/user/urls).
//
// Проверяет следующие сценарии:
//   - Успешное получение списка URL авторизованного пользователя (StatusOK)
//   - Пустой результат для нового пользователя (StatusNoContent)
//   - Ошибку авторизации при отсутствии cookie (StatusUnauthorized)
//   - Корректность формата JSON ответа
//   - Автоматическую выдачу новой cookie при неавторизованном доступе
//   - Работу JWT авторизации через StrictJWTAutoIssue middleware
//
// Тест создает:
//   - Тестовое хранилище с URL для двух пользователей
//   - HTTP-сервер с middleware:
//   - Логирование запросов
//   - Поддержка gzip
//   - Строгая JWT авторизация с автоматической выдачей
//   - Набор тестовых случаев с разными условиями авторизации
//
// Примеры тест-кейсов:
//   - Авторизованный пользователь с URL (ожидается 200 OK с корректным списком)
//   - Новый пользователь без URL (ожидается 204 NoContent)
//   - Запрос без авторизации (ожидается 401 Unauthorized с новой cookie)
//
// Особенности:
//   - Проверяет строгую авторизацию через StrictJWTAutoIssue
//   - Тестирует автоматическую выдачу cookie при неавторизованном доступе
//   - Поддерживает проверку структуры JSON ответа
func TestUserUrls(t *testing.T, serv *service.Service, client *resty.Client) {

	// Тестовые данные
	cookie1, user1 := testutils.CreateSignedCookie()
	_, user2 := testutils.CreateSignedCookie()

	testURLs := []models.UserURLMapping{
		{UserID: user1, OriginalURL: "https://example.com/1"},
		{UserID: user1, OriginalURL: "https://example.com/2"},
		{UserID: user2, OriginalURL: "https://example.com/3"},
	}

	prepareTestUserURLs(serv, testURLs)

	t.Run("Успешное получение ссылок пользователя", func(t *testing.T) {

		// Запрос
		resp, err := client.R().
			SetCookie(cookie1).
			Get("/api/user/urls")

		// Проверки
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		var urls []models.URLMapping
		err = json.Unmarshal(resp.Body(), &urls)
		assert.NoError(t, err)
		assert.Len(t, urls, 2) // user1 имеет 2 ссылки

	})
	t.Run("Пустой результат для нового пользователя", func(t *testing.T) {
		cookie, _ := testutils.CreateSignedCookie()

		resp, err := client.R().
			SetCookie(cookie).
			Get("/api/user/urls")

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode())

	})

	t.Run("Ошибка без куки", func(t *testing.T) {
		resp, err := client.R().
			Get("/api/user/urls")

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode())

		// Проверяем что установлена новая кука
		assert.NotEmpty(t, resp.Cookies())
	})

}

func TestUserUrlsGRPC(t *testing.T, serv *service.Service, grpcClient pb.ShortenerClient) {

	// Тестовые данные
	cookie1, user1 := testutils.CreateSignedCookie()
	_, user2 := testutils.CreateSignedCookie()

	testURLs := []models.UserURLMapping{
		{UserID: user1, OriginalURL: "https://example.com/1"},
		{UserID: user1, OriginalURL: "https://example.com/2"},
		{UserID: user2, OriginalURL: "https://example.com/3"},
	}

	prepareTestUserURLs(serv, testURLs)

	t.Run("Успешное получение ссылок пользователя", func(t *testing.T) {

		// Запрос
		token := cookie1.Value
		ctx := testutils.ContextWithJWT(context.Background(), token)

		resp, err := grpcClient.GetUserURLs(ctx, &pb.UserURLsRequest{})

		// Проверки
		assert.NoError(t, err)

		statusGetUserURLs := testutils.StatusOK
		if len(resp.Urls) == 0 {
			statusGetUserURLs = testutils.StatusNoContent
		}

		var urls []models.URLMapping
		if statusGetUserURLs == testutils.StatusOK {
			for _, u := range resp.Urls {
				urls = append(urls, models.URLMapping{
					ShortURL:    u.ShortUrl,
					OriginalURL: u.OriginalUrl,
				})
			}
		}

		assert.Equal(t, testutils.StatusOK, statusGetUserURLs)

		assert.Len(t, urls, 2) // user1 имеет 2 ссылки

	})
	t.Run("Пустой результат для нового пользователя", func(t *testing.T) {

		cookie, _ := testutils.CreateSignedCookie()
		ctx := testutils.ContextWithJWT(context.Background(), cookie.Value)

		resp, err := grpcClient.GetUserURLs(ctx, &pb.UserURLsRequest{})

		assert.NoError(t, err)
		statusGetUserURLs := testutils.StatusOK
		if len(resp.Urls) == 0 {
			statusGetUserURLs = testutils.StatusNoContent
		}

		assert.Equal(t, testutils.StatusNoContent, statusGetUserURLs)

	})
}

func prepareTestUserURLs(serv *service.Service, testURLs []models.UserURLMapping) {

	for _, url := range testURLs {
		ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, url.UserID)
		_, err := serv.GetShortKey(ctx, url.OriginalURL)
		if err != nil {
			panic(err)
		}
		// url.ShortURL = shortURL
	}

}
