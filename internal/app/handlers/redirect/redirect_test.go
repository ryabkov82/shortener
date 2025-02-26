package redirect

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/storage"

	"github.com/stretchr/testify/assert"
)

func TestGetHandler(t *testing.T) {

	storage := storage.New()

	storage.SaveURL("https://practicum.yandex.ru/", "EYm7J2zF")

	tests := []struct {
		name           string
		originalURL    string
		shortKey       string
		wantStatusCode int
	}{
		{
			name:           "positive test #1",
			originalURL:    "https://practicum.yandex.ru/",
			shortKey:       "EYm7J2zF",
			wantStatusCode: 307,
		},
		{
			name:           "negative test #2",
			shortKey:       "RrixjW0q",
			wantStatusCode: 404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/"+tt.shortKey, nil)
			request.SetPathValue("id", tt.shortKey)
			w := httptest.NewRecorder()
			h := GetHandler(storage)
			h(w, request)
			result := w.Result()
			result.Body.Close()
			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, result.StatusCode)
			if tt.wantStatusCode == 307 {
				assert.Equal(t, tt.originalURL, w.Header().Get("Location"))
			}
		})
	}
}
