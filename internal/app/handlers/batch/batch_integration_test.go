package batch_test

import (
	"os"
	"testing"

	"github.com/ryabkov82/shortener/test/testhandlers"
	"github.com/ryabkov82/shortener/test/testutils"
)

func TestGetHandler_InMemory(t *testing.T) {

	st, err := testutils.InitializeInMemoryStorage()
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	defer os.Remove(st.FilePath())

	testhandlers.TestBatch(t, st)
}
