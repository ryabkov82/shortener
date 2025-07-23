package pprof

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/ryabkov82/shortener/internal/app/config"
)

func TestStartPProf_Disabled(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.PProfConfig{
		Enabled: false,
	}

	// Просто проверяем что функция не паникует при отключенном pprof
	StartPProf(logger, cfg)
}

func TestBasicAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		user           string
		pass           string
		expectedStatus int
	}{
		{"Valid credentials", "admin", "secret", http.StatusOK},
		{"Invalid user", "wrong", "secret", http.StatusUnauthorized},
		{"Invalid pass", "admin", "wrong", http.StatusUnauthorized},
		{"No auth", "", "", http.StatusUnauthorized},
	}

	handler := basicAuthMiddleware("admin", "secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.user != "" || tt.pass != "" {
				req.SetBasicAuth(tt.user, tt.pass)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestPProfRoutesRegistration(t *testing.T) {

	cfg := config.PProfConfig{
		Enabled:  true,
		BindAddr: "localhost:0", // 0 означает случайный порт
		Endpoint: "/debug/pprof",
		AuthUser: "admin",
		AuthPass: "secret",
	}

	// Создаем роутер и запускаем pprof
	r := chi.NewRouter()
	registerPProfRoutes(r, cfg)

	// Тестируем только GET-роуты (кроме profile)
	testCases := []struct {
		method string
		path   string
	}{
		{"GET", "/debug/pprof/"},
		{"GET", "/debug/pprof/cmdline"},
		{"GET", "/debug/pprof/symbol"},
		{"GET", "/debug/pprof/trace"},
		// Не тестируем "/debug/pprof/profile" чтобы избежать зависания
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.SetBasicAuth("admin", "secret")

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.NotEqual(t, http.StatusNotFound, rr.Code, "route %s not found", tc.path)
			assert.NotEqual(t, http.StatusUnauthorized, rr.Code, "unauthorized for route %s", tc.path)
		})
	}

	// Отдельно тестируем handlers (без выполнения профилирования)
	handlerCases := []struct {
		path        string
		profileType string
	}{
		{"/debug/pprof/goroutine", "goroutine"},
		{"/debug/pprof/heap", "heap"},
		{"/debug/pprof/allocs", "allocs"},
		{"/debug/pprof/threadcreate", "threadcreate"},
		{"/debug/pprof/block", "block"},
		{"/debug/pprof/mutex", "mutex"},
	}

	for _, hc := range handlerCases {
		t.Run(hc.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", hc.path, nil)
			req.SetBasicAuth("admin", "secret")

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "unexpected status for %s", hc.path)
		})
	}
}
