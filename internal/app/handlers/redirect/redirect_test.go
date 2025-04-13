package redirect

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/logger"

	"github.com/ryabkov82/shortener/internal/app/models"
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
	if err != nil && os.IsNotExist(err) {
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

	mapping := models.URLMapping{
		ShortURL:    "EYm7J2zF",
		OriginalURL: "https://practicum.yandex.ru/",
	}

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.JWTAutoIssueMiddleware(testSecretKey))

	r.Get("/{id}", GetHandler(service, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	var redirectAttemptedError = errors.New("redirect")
	redirectPolicy := resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		// return nil for continue redirect otherwise return error to stop/prevent redirect
		return redirectAttemptedError
	})

	cookie, userID := createSignedCookie()
	ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, userID)
	st.SaveURL(ctx, &mapping)

	tests := []struct {
		name           string
		originalURL    string
		shortKey       string
		cookie         *http.Cookie
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

			client := resty.New()
			client.SetRedirectPolicy(redirectPolicy)
			req := client.R().SetCookie(tt.cookie)
			req.Method = http.MethodGet
			req.URL = srv.URL + "/" + tt.shortKey

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
