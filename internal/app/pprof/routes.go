// Package pprof предоставляет HTTP-интерфейс для профилирования приложения
// с использованием стандартного пакета net/http/pprof.
package pprof

import (
	"net/http"
	"net/http/pprof"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"

	"github.com/ryabkov82/shortener/internal/app/config"
)

// StartPProf запускает HTTP-сервер для профилирования приложения.
//
// Параметры:
//   - log: логер для записи ошибок
//   - config: конфигурация сервера профилирования
//   - Enabled: флаг включения сервера
//   - BindAddr: адрес для прослушивания (например, "localhost:6060")
//   - Endpoint: базовый URL для эндпоинтов (например, "/debug/pprof")
//   - AuthUser: логин для HTTP Basic Auth
//   - AuthPass: пароль для HTTP Basic Auth
//
// Пример использования:
//
//	cfg := config.PProfConfig{
//	    Enabled:  true,
//	    BindAddr: "localhost:6060",
//	    Endpoint: "/debug/pprof",
//	    AuthUser: "admin",
//	    AuthPass: "secret",
//	}
//	pprof.StartPProf(logger.Log, cfg)
func StartPProf(log *zap.Logger, config config.PProfConfig) {
	if !config.Enabled {
		return
	}
	r := chi.NewRouter()

	// Регистрируем стандартные обработчики pprof
	r.Route(config.Endpoint, func(r chi.Router) {
		// Применяем аутентификацию ко всем под-роутам
		r.Use(basicAuthMiddleware(config.AuthUser, config.AuthPass))

		// Регистрируем стандартные обработчики pprof
		r.Get("/", http.HandlerFunc(pprof.Index))
		r.Get("/cmdline", http.HandlerFunc(pprof.Cmdline))
		r.Get("/profile", http.HandlerFunc(pprof.Profile))
		r.Get("/symbol", http.HandlerFunc(pprof.Symbol))
		r.Get("/trace", http.HandlerFunc(pprof.Trace))

		// Регистрируем обработчики профилей
		r.Handle("/goroutine", pprof.Handler("goroutine"))
		r.Handle("/heap", pprof.Handler("heap"))
		r.Handle("/allocs", pprof.Handler("allocs"))
		r.Handle("/threadcreate", pprof.Handler("threadcreate"))
		r.Handle("/block", pprof.Handler("block"))
		r.Handle("/mutex", pprof.Handler("mutex"))
	})

	server := &http.Server{
		Addr:    config.BindAddr,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to serve pprof server", zap.Error(err))
		}
	}()
}

// basicAuthMiddleware создает middleware для HTTP Basic Authentication.
//
// Параметры:
//   - expectedUser: ожидаемое имя пользователя
//   - expectedPass: ожидаемый пароль
//
// Возвращает:
//   - middleware функцию для chi.Router
func basicAuthMiddleware(expectedUser, expectedPass string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, password, ok := r.BasicAuth()

			if !ok || user != expectedUser || password != expectedPass {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
