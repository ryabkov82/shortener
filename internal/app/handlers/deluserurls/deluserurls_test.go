package deluserurls

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ryabkov82/shortener/internal/app/models"
	"github.com/ryabkov82/shortener/internal/app/storage/postgres"

	"github.com/ryabkov82/shortener/internal/app/handlers/redirect"
	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/testutils"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
)

func TestGetHandler_InMemory(t *testing.T) {

	st, err := testutils.InitializeInMemoryStorage()

	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	testDelUserUrls(t, st)

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

	testDelUserUrls(t, pg)
}

func testDelUserUrls(t *testing.T, st service.Repository) {

	service := service.NewService(st)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.StrictJWTAutoIssue(testutils.TestSecretKey))

	baseURL := "http://localhost:8080/"
	r.Delete("/api/user/urls", GetHandler(service, baseURL, logger.Log))
	r.Get("/{id}", redirect.GetHandler(service, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	// Клиент resty
	client := resty.New().SetBaseURL(srv.URL)

	// Тестовые данные
	cookie1, user1 := testutils.CreateSignedCookie()
	cookie2, user2 := testutils.CreateSignedCookie()
	testURLs := []models.UserURLMapping{
		{UserID: user1, OriginalURL: "https://example.com/1"},
		{UserID: user1, OriginalURL: "https://example.com/2"},
		{UserID: user1, OriginalURL: "https://example.com/3"},
		{UserID: user1, OriginalURL: "https://example.com/4"},
		{UserID: user2, OriginalURL: "https://example.com/5"},
	}

	// Заполняем хранилище
	for i, url := range testURLs {
		ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, url.UserID)
		shortURL, err := service.GetShortKey(ctx, url.OriginalURL)
		if err != nil {
			panic(err)
		}
		testURLs[i].ShortURL = shortURL
	}

	tests := []struct {
		cookie         *http.Cookie
		name           string
		userID         string
		codesToDelete  []string
		shouldBeMarked []string
		wantStatus     int
	}{
		{
			name:           "successful deletion",
			userID:         user1,
			cookie:         cookie1,
			codesToDelete:  []string{testURLs[0].ShortURL},
			wantStatus:     http.StatusAccepted,
			shouldBeMarked: []string{testURLs[0].ShortURL},
		},
		{
			name:           "delete multiple",
			userID:         user1,
			cookie:         cookie1,
			codesToDelete:  []string{testURLs[1].ShortURL, testURLs[2].ShortURL},
			wantStatus:     http.StatusAccepted,
			shouldBeMarked: []string{testURLs[1].ShortURL, testURLs[2].ShortURL},
		},
		{
			name:           "delete non-existent",
			userID:         user1,
			cookie:         cookie1,
			codesToDelete:  []string{"nonexistent"},
			wantStatus:     http.StatusAccepted,
			shouldBeMarked: []string{},
		},
		{
			name:           "delete other user's url",
			userID:         user2,
			cookie:         cookie2,
			codesToDelete:  []string{testURLs[3].ShortURL},
			wantStatus:     http.StatusAccepted,
			shouldBeMarked: []string{}, // Не должно пометить как удаленный
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Подготовка запроса
			body, _ := json.Marshal(tt.codesToDelete)
			buf := bytes.NewBuffer(nil)
			zb := gzip.NewWriter(buf)
			_, err := zb.Write([]byte(body))
			assert.NoError(t, err)
			err = zb.Close()
			assert.NoError(t, err)

			// Запрос
			resp, err := client.R().
				SetBody(buf).
				SetCookie(tt.cookie).
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Accept-Encoding", "gzip").
				Delete("/api/user/urls")

			// Проверки
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode())

			time.Sleep(500 * time.Millisecond)
			// Проверка что URL помечены как удаленные
			for _, code := range tt.codesToDelete {

				resp, err := client.R().
					SetCookie(tt.cookie).
					Get("/" + code)

				// Проверки
				assert.NoError(t, err)
				if resp.StatusCode() == http.StatusGone {
					assert.Contains(t, tt.shouldBeMarked, code)
				} else {
					assert.NotContains(t, tt.shouldBeMarked, code)
				}

			}
		})
	}

}
