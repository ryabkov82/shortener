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
	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
	"github.com/ryabkov82/shortener/internal/testutils"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetHandler_InMemory(t *testing.T) {

	st, err := testutils.InitializeInMemoryStorage()

	if err != nil {
		t.Fatal(err)
	}

	testRedirect(t, st)
}

func TestGetHandler_Postgres(t *testing.T) {

	dsn := os.Getenv("TEST_DB_DSN")

	if dsn == "" {
		t.Fatal("TEST_DB_DSN не установлен")
	}
	pg, err := postgres.NewPostgresStorage(dsn)

	if err != nil {
		t.Fatal(err)
	}

	testRedirect(t, pg)
}

func testRedirect(t *testing.T, st service.Repository) {
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
	r.Use(auth.JWTAutoIssue(testutils.TestSecretKey))

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

	cookie, userID := testutils.CreateSignedCookie()
	ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, userID)
	st.SaveURL(ctx, &mapping)

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
