package userurls_test

import (
	"testing"

	"github.com/ryabkov82/shortener/test/testutils"
	"github.com/ryabkov82/shortener/test/testutils/handlers"
)

func TestGetHandler_InMemory(t *testing.T) {

	st, err := testutils.InitializeInMemoryStorage()

	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	handlers.TestUserUrls(t, st)
}
