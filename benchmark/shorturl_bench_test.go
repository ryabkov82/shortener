package benchmark

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/logger"

	"github.com/ryabkov82/shortener/internal/app/handlers/http/shorturl"

	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/http/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/mwgzip"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/storage/inmemory"
	"github.com/ryabkov82/shortener/internal/app/storage/postgres"

	"github.com/go-chi/chi/v5"
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

func BenchmarkHandleShorten_InMemory(b *testing.B) {

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

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.JWTAutoIssue(testSecretKey))

	baseURL := "http://localhost:8080/"
	r.Post("/", shorturl.GetHandler(srv, baseURL, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	serv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer serv.Close()

	benchmarkHandleShorten(b, serv)

}

func BenchmarkHandleShorten_Postgres(b *testing.B) {
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

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)
	r.Use(auth.JWTAutoIssue(testSecretKey))

	baseURL := "http://localhost:8080/"
	r.Post("/", shorturl.GetHandler(srv, baseURL, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	serv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer serv.Close()

	benchmarkHandleShorten(b, serv)
}

func benchmarkHandleShorten(b *testing.B, serv *httptest.Server) {

	cookie, _ := createSignedCookie()

	log.Printf("benchmarkHandleShorten start, b.N = %d", b.N)

	client := resty.New() // Один раз перед циклом

	b.ResetTimer()

	var urlBuilder strings.Builder
	for i := 0; i < b.N; i++ {

		// originalURL := "https://example.com/" + strconv.Itoa(i)
		urlBuilder.Reset()
		urlBuilder.WriteString("https://example.com/")
		urlBuilder.WriteString(strconv.Itoa(i))
		originalURL := urlBuilder.String()

		resp, err := client.R().
			SetCookie(cookie).
			SetBody(originalURL).
			Post(serv.URL)

		if err != nil {
			b.Errorf("Ошибка выполнения запроса: %v", err)
		}
		if resp.StatusCode() != http.StatusCreated {
			b.Errorf("Expected status 201, got %d. Body: %s", resp.StatusCode(), string(resp.Body()))
		}

	}

}
