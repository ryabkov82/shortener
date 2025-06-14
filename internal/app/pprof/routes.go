package pprof

import (
	"net/http"
	"net/http/pprof"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/ryabkov82/shortener/internal/app/config"
)

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
		Handler: r, // Ваш роутер
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to serve server", zap.Error(err))
		}
	}()
}

func basicAuthMiddleware(expectedUser, expectedPass string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем credentials из заголовка
			user, password, ok := r.BasicAuth()

			// Проверяем их
			if !ok || user != expectedUser || password != expectedPass {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Если проверка пройдена - передаем запрос дальше
			next.ServeHTTP(w, r)
		})
	}
}
