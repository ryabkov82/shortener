package shorturl

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

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

func createSignedCookie() *http.Cookie {

	tokenString, _, err := jwtauth.GenerateNewToken(testSecretKey)
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
	}

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
	r.Use(auth.JWTAutoIssue(testSecretKey))

	baseURL := "http://localhost:8080/"
	r.Post("/", GetHandler(service, baseURL, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	cookie := createSignedCookie()

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
