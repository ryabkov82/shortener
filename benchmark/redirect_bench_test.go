package benchmark

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"

	"github.com/ryabkov82/shortener/internal/app/logger"

	"github.com/ryabkov82/shortener/internal/app/handlers/redirect"
	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/storage/inmemory"
	"github.com/ryabkov82/shortener/internal/app/storage/postgres"
)

var (
	shortURLs []string
)

// Предварительное заполнение (1000 записей)
const prefillCount = 1000

func initStorage(userID string, srv *service.Service) {
	shortURLs = make([]string, prefillCount)
	ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, userID)
	for i := 0; i < prefillCount; i++ {
		originalURL := fmt.Sprintf("https://example.com/%d", i)
		shortURL, _ := srv.GetShortKey(ctx, originalURL)
		shortURLs[i] = shortURL
	}
}

// go test -bench=HandleRedirect_InMemory -benchmem -memprofile="profiles/base_1.pprof" -cpu=4 -benchtime=5s ./benchmark
func BenchmarkHandleRedirect_InMemory(b *testing.B) {

	fileStorage := "test.dat"
	_ = os.Remove(fileStorage)
	/*
		if err != nil && !os.IsNotExist(err) {
			panic(err)
		}
	*/

	st, err := inmemory.NewInMemoryStorage(fileStorage)
	if err != nil {
		panic(err)
	}
	st.Load(fileStorage)

	srv := service.NewService(st)

	if err := logger.Initialize("info"); err != nil {
		b.Fatalf("Не удалось инициализировать logger: %v", err)
	}

	cookie, userID := createSignedCookie()

	// Предварительное заполнение (1000 записей)
	initStorage(userID, srv)

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.JWTAutoIssue(testSecretKey))

	r.Get("/{id}", redirect.GetHandler(srv, logger.Log))
	serv := httptest.NewServer(r)
	defer serv.Close()

	benchmarkHandleRedirect(b, serv, cookie)

}

// go test -bench=HandleRedirect_Postgres -benchmem -memprofile="profiles/base_1_pg.pprof" -cpu=4 -benchtime=5s ./benchmark
func BenchmarkHandleRedirect_Postgres(b *testing.B) {

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		b.Skip("TEST_DB_DSN не установлен, пропускаем бенчмарк с Postgres")
	}

	pg, err := postgres.NewPostgresStorage(dsn)

	if err != nil {
		b.Fatalf("Не удалось инициализировать Postgres: %v", err)
	}
	srv := service.NewService(pg)

	if err := logger.Initialize("info"); err != nil {
		b.Fatalf("Не удалось инициализировать logger: %v", err)
	}

	cookie, userID := createSignedCookie()

	// Предварительное заполнение (1000 записей)
	initStorage(userID, srv)

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.JWTAutoIssue(testSecretKey))

	r.Get("/{id}", redirect.GetHandler(srv, logger.Log))
	serv := httptest.NewServer(r)
	defer serv.Close()

	benchmarkHandleRedirect(b, serv, cookie)

}

func benchmarkHandleRedirect(b *testing.B, serv *httptest.Server, cookie *http.Cookie) {

	var redirectAttemptedError = errors.New("redirect")
	redirectPolicy := resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		// return nil for continue redirect otherwise return error to stop/prevent redirect
		return redirectAttemptedError
	})

	client := resty.New() // Один раз перед циклом

	var counter int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&counter, 1)
			// Используем round-robin по предзаполненным URL
			shortCode := shortURLs[counter%prefillCount]
			client.SetRedirectPolicy(redirectPolicy)
			req := client.R().SetCookie(cookie)
			req.Method = http.MethodGet
			req.URL = serv.URL + "/" + shortCode

			resp, err := req.Send()

			if errors.Is(err, redirectAttemptedError) {
				// эту ошибку игнорируем
				err = nil
			}

			if err != nil {
				b.Errorf("Request failed: %v", err)
			}

			if resp.StatusCode() != http.StatusTemporaryRedirect {
				b.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, resp.StatusCode())
			}

		}
	})

}
