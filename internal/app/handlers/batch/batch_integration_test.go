package batch_test

import (
	"os"
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
	defer os.Remove(st.FilePath())

	handlers.TestBatch(t, st)
}
