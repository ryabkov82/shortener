package deluserurls_test

import (
	"testing"

	"github.com/ryabkov82/shortener/internal/app/handlers/deluserurls"
	"github.com/ryabkov82/shortener/internal/app/handlers/redirect"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"

	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/test/testhandlers"
	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/go-chi/chi/v5"
)

func TestGetHandler_InMemory(t *testing.T) {

	st, err := testutils.InitializeInMemoryStorage()

	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	service := service.NewService(st)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	baseURL := "http://localhost:8080/"

	tc := testutils.NewTestClient(func(r chi.Router) {
		r.Use(mwlogger.RequestLogging(logger.Log))
		r.Use(mwgzip.Gzip)
		r.Use(auth.StrictJWTAutoIssue(testutils.TestSecretKey))

		r.Delete("/api/user/urls", deluserurls.GetHandler(service, baseURL, logger.Log))
		r.Get("/{id}", redirect.GetHandler(service, logger.Log))
	})
	defer tc.Close()

	testhandlers.TestDelUserUrls(t, service, tc.Client)

}
