package userurls

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/models"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/service"
	storage "github.com/ryabkov82/shortener/internal/app/storage/inmemory"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

var (
	testSecretKey = []byte("test-secret-key")
)

func createSignedCookie() (*http.Cookie, string) {

	tokenString, userID, err := jwtauth.GenerateNewToken(testSecretKey)
	if err != nil {
		panic(err)
	}

	return &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
		//Secure:   true, // HTTPS-only
		SameSite: http.SameSiteStrictMode,
	}, userID

}

func TestGetHandler(t *testing.T) {

	fileStorage := "test.dat"
	err := os.Remove(fileStorage)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	st, err := storage.NewInMemoryStorage(fileStorage)
	if err != nil {
		panic(err)
	}
	st.Load(fileStorage)

	service := service.NewService(st)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.StrictJWTAutoIssue(testSecretKey))

	baseURL := "http://localhost:8080/"
	r.Get("/api/user/urls", GetHandler(service, baseURL, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	// Клиент resty
	client := resty.New().SetBaseURL(srv.URL)

	// Тестовые данные
	cookie1, user1 := createSignedCookie()
	_, user2 := createSignedCookie()
	testURLs := []models.UserURLMapping{
		{UserID: user1, OriginalURL: "https://example.com/1"},
		{UserID: user1, OriginalURL: "https://example.com/2"},
		{UserID: user2, OriginalURL: "https://example.com/3"},
	}

	// Заполняем хранилище
	for _, url := range testURLs {
		ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, url.UserID)
		_, err := service.GetShortKey(ctx, url.OriginalURL)
		if err != nil {
			panic(err)
		}
		//url.ShortURL = shortURL
	}

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
		cookie, _ := createSignedCookie()

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
