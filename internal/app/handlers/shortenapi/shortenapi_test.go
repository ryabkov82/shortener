package shortenapi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/logger"

	"github.com/ryabkov82/shortener/internal/app/service"
	storage "github.com/ryabkov82/shortener/internal/app/storage/inmemory"

	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"

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

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.JWTAutoIssue(testSecretKey))

	cookie := createSignedCookie()

	baseURL := "http://localhost:8080/"
	r.Post("/api/shorten", GetHandler(service, baseURL, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	tests := []struct {
		name           string
		request        Request
		cookie         *http.Cookie
		wantStatusCode int
	}{
		{
			name:           "positive test #1",
			request:        Request{URL: "https://practicum.yandex.ru/"},
			cookie:         cookie,
			wantStatusCode: 201,
		},
		{
			name:           "positive test #2",
			request:        Request{URL: "https://practicum.yandex.ru/"},
			cookie:         cookie,
			wantStatusCode: 409,
		},
		{
			name:           "negative test #2",
			request:        Request{URL: "not url"},
			cookie:         cookie,
			wantStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			buf := bytes.NewBuffer(nil)
			zb := gzip.NewWriter(buf)
			_, err = zb.Write([]byte(req))
			assert.NoError(t, err)
			err = zb.Close()
			assert.NoError(t, err)

			resp, err := resty.New().R().
				SetCookie(tt.cookie).
				SetBody(buf).
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Accept-Encoding", "gzip").
				Post(srv.URL + "/api/shorten")

			assert.NoError(t, err)

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())
			if tt.wantStatusCode == 201 {
				var response Response
				err = json.Unmarshal(resp.Body(), &response)
				assert.NoError(t, err)
				shortURL := response.Result
				// Проверяем, что получен URL
				_, err = url.Parse(shortURL)
				assert.NoError(t, err)
			}
		})
	}
}
