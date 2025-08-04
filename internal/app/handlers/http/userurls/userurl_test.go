package userurls_test

import (
	"testing"

	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/http/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/mwgzip"

	"github.com/ryabkov82/shortener/internal/app/handlers/http/userurls"
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

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	service := service.NewService(st)

	baseURL := "http://localhost:8080/"

	tc := testutils.NewTestClient(func(r chi.Router) {
		r.Use(mwlogger.RequestLogging(logger.Log))
		r.Use(mwgzip.Gzip)
		r.Use(auth.StrictJWTAutoIssue(testutils.TestSecretKey))

		r.Get("/api/user/urls", userurls.GetHandler(service, baseURL, logger.Log))
	})
	defer tc.Close()

	testhandlers.TestUserUrls(t, service, tc.Client)
}
