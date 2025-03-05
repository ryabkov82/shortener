package shorturl

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetHandler(t *testing.T) {

	storage := storage.NewInMemoryStorage()

	service := service.NewService(storage)

	r := chi.NewRouter()
	baseURL := "http://localhost:8080/"
	r.Post("/", GetHandler(service, baseURL))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

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

			resp, err := resty.New().R().
				SetBody(tt.originalURL).
				Post(srv.URL)

			assert.NoError(t, err)

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())
			if tt.wantStatusCode == 201 {
				shortURL := resp.Body()
				// Проверяем, что получен URL
				_, err = url.Parse(string(shortURL))
				assert.NoError(t, err)
			}
		})
	}
}
