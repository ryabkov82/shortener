package shorturl

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHandler(t *testing.T) {

	storage := storage.New()

	tests := []struct {
		name           string
		originalURL    string
		wantStatusCode int
	}{
		{
			name:           "positive test #1",
			originalURL:    "https://practicum.yandex.ru/",
			wantStatusCode: 201,
		},
		{
			name:           "negative test #2",
			originalURL:    "not url",
			wantStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.originalURL))
			w := httptest.NewRecorder()
			h := GetHandler(storage)
			h(w, request)
			result := w.Result()
			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, result.StatusCode)
			if tt.wantStatusCode == 201 {
				shortURL, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				// Проверяем, что получен URL
				_, err = url.Parse(string(shortURL))
				assert.NoError(t, err)
			}
			err := result.Body.Close()
			require.NoError(t, err)
		})
	}
}
