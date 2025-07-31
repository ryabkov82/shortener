package redirect_test

import (
	"testing"

	"github.com/ryabkov82/shortener/internal/app/handlers/http/redirect"
	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/auth"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/http/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/mwgzip"

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

	tc := testutils.NewTestClient(func(r chi.Router) {
		r.Use(mwlogger.RequestLogging(logger.Log))
		r.Use(mwgzip.Gzip)
		r.Use(auth.JWTAutoIssue(testutils.TestSecretKey))

		r.Get("/{id}", redirect.GetHandler(service, logger.Log))
	})
	defer tc.Close()

	testhandlers.TestRedirect(t, st, tc.Client)
}
